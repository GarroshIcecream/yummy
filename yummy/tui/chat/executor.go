package chat

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/GarroshIcecream/yummy/yummy/tui/chat/callbacks"
	tools "github.com/GarroshIcecream/yummy/yummy/tui/chat/tools"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/memory"
)

// ExecutorService provides agent-based LLM interactions using langchaingo executor
type ExecutorService struct {
	executor     *agents.Executor
	llm          *ollama.LLM
	toolManager  *tools.ToolManager
	sessionLog   *db.SessionLog
	ollamaStatus *OllamaServiceStatus
	modelName    string
	ctx          context.Context
	sessionStats db.SessionStats
	systemPrompt string
}

// NewExecutorService creates a new executor service instance
func NewExecutorService(cookbook *db.CookBook, sessionLog *db.SessionLog, modelName string, systemPrompt string) (*ExecutorService, error) {
	ctx := context.Background()

	// Get Ollama service status
	ollamaStatus, err := GetOllamaServiceStatus(modelName)
	if err != nil {
		slog.Error("Failed to get ollama service status", "error", err)
		return nil, err
	}

	// Create tool manager with cookbook access
	toolManager := tools.NewToolManager(cookbook)

	// Initialize the LLM
	llm, err := ollama.New(
		ollama.WithModel(modelName),
	)
	if err != nil {
		slog.Error("Failed to create LLM", "error", err)
		return nil, err
	}

	mem := memory.NewConversationBuffer(
		memory.WithInputKey("input"),
		memory.WithOutputKey("output"),
	)

	// Create the agent with tools
	agent := agents.NewConversationalAgent(
		llm,
		toolManager.GetTools(),
		agents.WithCallbacksHandler(callbacks.NewDefaultAgentCallbackHandler(
			func(status string) {
				slog.Debug("Agent Callback: Status", "status", status)
			},
		)),
	)

	// Create the executor
	executor := agents.NewExecutor(
		agent,
		agents.WithMaxIterations(5),
		agents.WithMemory(mem),
		agents.WithReturnIntermediateSteps(),
		agents.WithCallbacksHandler(callbacks.NewDefaultAgentCallbackHandler(
			func(status string) {
				slog.Debug("Agent Callback: Status", "status", status)
			}),
		),
	)

	emptySessionStats := db.SessionStats{}
	service := &ExecutorService{
		executor:     executor,
		llm:          llm,
		modelName:    modelName,
		sessionLog:   sessionLog,
		ctx:          ctx,
		toolManager:  toolManager,
		ollamaStatus: ollamaStatus,
		sessionStats: emptySessionStats,
		systemPrompt: systemPrompt,
	}

	return service, nil
}

func (e *ExecutorService) GetSystemPrompt() string {
	return e.systemPrompt
}

func (e *ExecutorService) GetMemory() *memory.ConversationBuffer {
	return e.executor.GetMemory().(*memory.ConversationBuffer)
}

func (e *ExecutorService) SaveMessage(message string, role llms.ChatMessageType) error {
	err := e.sessionLog.SaveSessionMessage(
		e.sessionStats.SessionID,
		message,
		role,
		e.GetCurrentModelName(),
		0,
		0,
		0,
	)
	if err != nil {
		slog.Error("Failed to save message to database", "error", err)
		return err
	}

	return nil
}

// GenerateResponse generates a response using the executor
func (e *ExecutorService) GenerateResponse(message string) (string, error) {
	slog.Debug("Generating response with executor", "model", e.modelName, "input", message)
	if message == "" {
		return "", fmt.Errorf("no input provided")
	}

	if e.GetSessionID() == 0 {
		slog.Debug("No session selected, creating new session")
		err := e.NewSession()
		if err != nil {
			slog.Error("Failed to create new session", "error", err)
			return "", err
		}
		slog.Debug("New session created", "sessionID", e.GetSessionID())
	}

	// Update memory
	err := e.SaveMessage(message, llms.ChatMessageTypeHuman)
	if err != nil {
		slog.Error("Failed to register message", "error", err)
		return "", err
	}

	result, err := chains.Run(e.ctx, e.executor, message)
	if err != nil {
		slog.Error("Executor execution error", "error", err)
		return "", err
	}

	slog.Debug("Generated response", "result", result)
	err = e.SaveMessage(result, llms.ChatMessageTypeAI)
	if err != nil {
		slog.Error("Failed to register message", "error", err)
		return "", err
	}

	return result, nil
}

func (e *ExecutorService) GetMemoryConversation() ([]llms.ChatMessage, error) {
	messages, err := e.GetMemory().ChatHistory.Messages(e.ctx)
	if err != nil {
		slog.Error("Failed to get conversation", "error", err)
		return nil, err
	}
	return messages, nil
}

// GetSessionID returns the current session ID
func (e *ExecutorService) GetSessionID() uint {
	return e.sessionStats.SessionID
}

func (e *ExecutorService) SetMemory(conversation []llms.ChatMessage) error {
	err := e.ClearMemory()
	if err != nil {
		slog.Error("Failed to clear memory", "error", err)
		return err
	}
	for _, message := range conversation {
		err := e.GetMemory().ChatHistory.AddMessage(e.ctx, message)
		if err != nil {
			slog.Error("Failed to add message to memory", "error", err)
			return err
		}
	}
	return nil
}

func (e *ExecutorService) ClearMemory() error {
	err := e.GetMemory().Clear(e.ctx)
	if err != nil {
		slog.Error("Failed to clear memory", "error", err)
		return err
	}
	return nil
}

func (e *ExecutorService) ResetSession() error {
	err := e.ClearMemory()
	if err != nil {
		slog.Error("Failed to clear memory", "error", err)
		return err
	}

	e.sessionStats = db.SessionStats{}
	return nil
}

func (e *ExecutorService) NewSession() error {
	err := e.ClearMemory()
	if err != nil {
		slog.Error("Failed to clear memory", "error", err)
		return err
	}

	sessionID, err := e.sessionLog.CreateSession()
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		return err
	}

	err = e.AppendSystemPrompt(e.systemPrompt, sessionID)
	if err != nil {
		slog.Error("Failed to add system prompt to memory", "error", err)
		return err
	}

	newSessionStats, err := e.sessionLog.GetSessionStats(sessionID)
	if err != nil {
		slog.Error("Failed to get session stats", "error", err)
		return err
	}

	e.sessionStats = newSessionStats
	return nil
}

// LoadSession loads a session into the executor service
func (e *ExecutorService) LoadSession(sessionID uint) error {
	sessionMessages, err := e.sessionLog.GetSessionMessages(sessionID)
	if err != nil {
		slog.Error("Failed to get session messages", "error", err)
		return err
	}

	chatMessages := make([]llms.ChatMessage, len(sessionMessages))
	for i, msg := range sessionMessages {
		if msg.Role == string(llms.ChatMessageTypeSystem) {
			chatMessages[i] = llms.SystemChatMessage{Content: msg.Message}
		} else if msg.Role == string(llms.ChatMessageTypeHuman) {
			chatMessages[i] = llms.HumanChatMessage{Content: msg.Message}
		} else if msg.Role == string(llms.ChatMessageTypeAI) {
			chatMessages[i] = llms.AIChatMessage{Content: msg.Message}
		} else if msg.Role == string(llms.ChatMessageTypeTool) {
			chatMessages[i] = llms.ToolChatMessage{Content: msg.Message}
		}
	}

	// We need to preserve the system prompt so the LLM can remember previous messages
	err = e.SetMemory(chatMessages)
	if err != nil {
		slog.Error("Failed to set memory", "error", err)
		return err
	}

	// Set the model for the session
	err = e.SetModelByName(sessionMessages[0].ModelName, e.ollamaStatus)
	if err != nil {
		slog.Error("Failed to set model", "error", err)
		return err
	}

	// Reset counters
	newSessionStats := db.SessionStats{
		SessionID:    sessionMessages[0].SessionID,
		MessageCount: len(sessionMessages),
		InputTokens:  0,
		OutputTokens: 0,
		TotalTokens:  0,
	}

	// Calculate token counts from loaded messages
	for _, msg := range sessionMessages {
		newSessionStats.InputTokens += msg.InputTokens
		newSessionStats.OutputTokens += msg.OutputTokens
		newSessionStats.TotalTokens += msg.TotalTokens
	}

	e.sessionStats = newSessionStats
	return nil
}

// GetCurrentModelName returns the name of the current model
func (e *ExecutorService) GetCurrentModelName() string {
	return e.modelName
}

func (e *ExecutorService) GetSessionLog() *db.SessionLog {
	return e.sessionLog
}

// HasSystemPrompt checks if a system prompt is present in the conversation
func (e *ExecutorService) HasSystemPrompt() bool {
	messages, err := e.GetMemory().ChatHistory.Messages(e.ctx)
	if err != nil {
		return false
	}

	for _, message := range messages {
		if message.GetType() == llms.ChatMessageTypeSystem {
			return true
		}
	}
	return false
}

// SetModel sets the model for the executor service
func (e *ExecutorService) SetModelByName(modelName string, ollamaStatus *OllamaServiceStatus) error {
	if !slices.Contains(ollamaStatus.InstalledModels, modelName) {
		slog.Error("Model not installed", "model", modelName)
		return fmt.Errorf("model %s is not installed", modelName)
	}

	e.modelName = modelName
	llm, err := ollama.New(
		ollama.WithModel(modelName),
	)
	if err != nil {
		slog.Error("Failed to create LLM", "error", err)
		return err
	}

	// Recreate the agent and executor with the new model
	e.llm = llm
	agent := agents.NewConversationalAgent(
		llm,
		e.toolManager.GetTools(),
	)

	e.executor = agents.NewExecutor(
		agent,
		agents.WithMaxIterations(5),
		agents.WithMemory(e.GetMemory()),
		agents.WithReturnIntermediateSteps(),
		agents.WithCallbacksHandler(callbacks.NewDefaultAgentCallbackHandler(func(status string) {
			slog.Debug("Agent Callback: Status", "status", status)
		}),
		),
	)

	return nil
}

func (e *ExecutorService) AppendSystemPrompt(systemPrompt string, sessionID uint) error {
	systemMessage := llms.SystemChatMessage{Content: systemPrompt}
	err := e.GetMemory().ChatHistory.AddMessage(e.ctx, systemMessage)
	if err != nil {
		slog.Error("Failed to add system prompt to memory", "error", err)
		return err
	}

	err = e.sessionLog.SaveSessionMessage(sessionID, systemPrompt, llms.ChatMessageTypeSystem, e.modelName, 0, 0, 0)
	if err != nil {
		slog.Error("Failed to save system prompt to database", "error", err)
		return err
	}

	return nil
}
