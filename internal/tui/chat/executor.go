package chat

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/tui/chat/callbacks"
	tools "github.com/GarroshIcecream/yummy/internal/tui/chat/tools"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/memory"
)

// ExecutorService provides agent-based LLM interactions using langchaingo executor
type ExecutorService struct {
	executor        *agents.Executor
	llm             *ollama.LLM
	toolManager     *tools.ToolManager
	cookbook        *db.CookBook
	sessionLog      *db.SessionLog
	ollamaStatus    *OllamaServiceStatus
	modelName       string
	ctx             context.Context
	cancelCtx       context.CancelFunc
	sessionStats    db.SessionStats
	systemPrompt    string
	maxIterations   int
	callbackHandler *callbacks.DefaultAgentCallbackHandler
	streamCh        chan string
}

// NewExecutorService creates a new executor service instance
func NewExecutorService(cookbook *db.CookBook, sessionLog *db.SessionLog) (*ExecutorService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	chatConfig := config.GetChatConfig()

	// Get Ollama service status
	ollamaStatus, err := GetOllamaServiceStatus(chatConfig.DefaultModel)
	if err != nil {
		slog.Error("Failed to get ollama service status", "error", err)
		cancel()
		return nil, err
	}

	// Create tool manager with cookbook access
	toolManager := tools.NewToolManager(cookbook)

	// Initialize the LLM
	llm, err := ollama.New(
		ollama.WithModel(chatConfig.DefaultModel),
	)
	if err != nil {
		slog.Error("Failed to create LLM", "error", err)
		cancel()
		return nil, err
	}

	mem := memory.NewConversationBuffer(
		memory.WithInputKey("input"),
		memory.WithOutputKey("output"),
	)

	// Create a buffered channel for streaming chunks from the LLM callback to the TUI
	streamCh := make(chan string, 64)

	// Create the executor
	tools := toolManager.GetTools()
	callbackHandler := callbacks.NewDefaultAgentCallbackHandler(
		func(status string) {
			slog.Debug("Agent Callback: Status", "status", status)
		},
	)
	callbackHandler.SetStreamingCallback(func(chunk string) {
		streamCh <- chunk
	})
	executor := agents.NewExecutor(
		agents.NewConversationalAgent(llm, tools, agents.WithCallbacksHandler(callbackHandler)),
		agents.WithMaxIterations(chatConfig.MaxIterations),
		agents.WithMemory(mem),
		agents.WithReturnIntermediateSteps(),
		agents.WithCallbacksHandler(callbackHandler),
	)

	emptySessionStats := db.SessionStats{}
	service := &ExecutorService{
		executor:        executor,
		llm:             llm,
		modelName:       chatConfig.DefaultModel,
		cookbook:        cookbook,
		sessionLog:      sessionLog,
		ctx:             ctx,
		cancelCtx:       cancel,
		toolManager:     toolManager,
		ollamaStatus:    ollamaStatus,
		sessionStats:    emptySessionStats,
		systemPrompt:    chatConfig.SystemPrompt,
		maxIterations:   chatConfig.MaxIterations,
		callbackHandler: callbackHandler,
		streamCh:        streamCh,
	}

	return service, nil
}

func (e *ExecutorService) GetSystemPrompt() string {
	return e.systemPrompt
}

// GetStreamCh returns the read-only channel that emits streaming chunks from the LLM.
func (e *ExecutorService) GetStreamCh() <-chan string {
	return e.streamCh
}

// CancelStreaming cancels the in-flight LLM generation (if any), drains and
// replaces the stream channel, and creates a fresh cancellable context so the
// next generation can proceed.
func (e *ExecutorService) CancelStreaming() {
	// Cancel the running context to abort chains.Run
	e.cancelCtx()

	// Drain any leftover chunks so the listener goroutine unblocks
	old := e.streamCh
	go func() {
		for range old {
		}
	}()
	close(old)

	// Create a fresh channel and context for the next request
	e.streamCh = make(chan string, 64)
	e.ctx, e.cancelCtx = context.WithCancel(context.Background())

	// Re-point the callback handler at the new channel
	e.callbackHandler.SetStreamingCallback(func(chunk string) {
		e.streamCh <- chunk
	})
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

// PrepareForGeneration ensures a session exists, saves the user message to
// memory/DB, and resets callback state. It must be called synchronously (on the
// main Bubble Tea goroutine) so the user message is visible in the conversation
// before the streaming goroutine starts.
func (e *ExecutorService) PrepareForGeneration(message string) error {
	if message == "" {
		return fmt.Errorf("no input provided")
	}

	if e.GetSessionID() == 0 {
		slog.Debug("No session selected, creating new session")
		err := e.NewSession()
		if err != nil {
			slog.Error("Failed to create new session", "error", err)
			return err
		}
		slog.Debug("New session created", "sessionID", e.GetSessionID())
	}

	// Save the user message so it appears in the conversation immediately
	err := e.SaveMessage(message, llms.ChatMessageTypeHuman)
	if err != nil {
		slog.Error("Failed to register message", "error", err)
		return err
	}

	// Reset callback state before running the chain
	e.callbackHandler.ResetTokenUsage()
	e.callbackHandler.ResetStreamBuffer()

	return nil
}

// GenerateResponse runs the LLM chain and returns the final response.
// PrepareForGeneration must be called before this method.
//
// promptMessage is sent to the LLM (may include resolved recipe context).
// displayMessage, if non-empty, replaces the human message in the in-memory
// conversation buffer after generation so that the UI shows the compact text
// (e.g. with @[Recipe] intact) rather than the expanded prompt.
func (e *ExecutorService) GenerateResponse(promptMessage, displayMessage string) (string, error) {
	slog.Debug("Generating response with executor", "model", e.modelName, "input", promptMessage)

	result, err := chains.Run(e.ctx, e.executor, promptMessage)
	if err != nil {
		slog.Error("Executor execution error", "error", err)
		return "", err
	}

	// If we have a compact display message, replace the augmented prompt in
	// the conversation memory so that re-renders show the clean version.
	if displayMessage != "" && displayMessage != promptMessage {
		e.replaceLastHumanMessage(displayMessage)
	}

	// Get token usage after execution
	usage := e.callbackHandler.GetTokenUsage()
	slog.Debug("Generated response",
		"result", result,
		"prompt_tokens", usage.PromptTokens,
		"completion_tokens", usage.CompletionTokens,
		"total_tokens", usage.TotalTokens)

	err = e.SaveMessage(result, llms.ChatMessageTypeAI)
	if err != nil {
		slog.Error("Failed to register message", "error", err)
		return "", err
	}

	// Generate and update session summary asynchronously
	go e.GenerateAndUpdateSessionSummary()

	return result, nil
}

// replaceLastHumanMessage walks the conversation memory backwards and replaces
// the last human message with the given text. This is used to swap out the
// augmented LLM prompt (which includes full recipe data) with the compact
// display text (which keeps @[RecipeName] intact).
func (e *ExecutorService) replaceLastHumanMessage(displayText string) {
	msgs, err := e.GetMemory().ChatHistory.Messages(e.ctx)
	if err != nil {
		slog.Error("Failed to read memory for mention cleanup", "error", err)
		return
	}

	// Find the last human message and replace it.
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].GetType() == llms.ChatMessageTypeHuman {
			msgs[i] = llms.HumanChatMessage{Content: displayText}
			break
		}
	}

	// Rewrite the full conversation buffer.
	if err := e.SetMemory(msgs); err != nil {
		slog.Error("Failed to rewrite memory after mention cleanup", "error", err)
	}
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

	// Set session stats directly — GetSessionStats excludes system messages so
	// it would return SessionID=0 for a brand-new session that only has the
	// system prompt.
	e.sessionStats = db.SessionStats{SessionID: sessionID}
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

	// Get session stats (excludes system messages)
	newSessionStats, err := e.sessionLog.GetSessionStats(sessionID)
	if err != nil {
		slog.Error("Failed to get session stats", "error", err)
		return err
	}

	e.sessionStats = newSessionStats
	return nil
}

// GetCurrentModelName returns the name of the current model
func (e *ExecutorService) GetCurrentModelName() string {
	return e.modelName
}

// GetInstalledModels returns the list of Ollama models available locally.
func (e *ExecutorService) GetInstalledModels() []string {
	return e.ollamaStatus.InstalledModels
}

func (e *ExecutorService) GetSessionLog() *db.SessionLog {
	return e.sessionLog
}

// GetSessionSummary returns the summary for the current session
func (e *ExecutorService) GetSessionSummary() (string, error) {
	sessionID := e.GetSessionID()
	if sessionID == 0 {
		return "", nil
	}
	return e.sessionLog.GetSessionSummary(sessionID)
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

// SetModelByName sets the model for the executor service
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
	mem := e.GetMemory()
	newHandler := callbacks.NewDefaultAgentCallbackHandler(func(status string) {
		slog.Debug("Agent Callback: Status", "status", status)
	})
	newHandler.SetStreamingCallback(func(chunk string) {
		e.streamCh <- chunk
	})
	e.callbackHandler = newHandler
	e.executor = agents.NewExecutor(
		agents.NewConversationalAgent(llm, e.toolManager.GetTools(), agents.WithCallbacksHandler(newHandler)),
		agents.WithMaxIterations(e.maxIterations),
		agents.WithMemory(mem),
		agents.WithReturnIntermediateSteps(),
		agents.WithCallbacksHandler(newHandler),
	)

	return nil
}

// AppendSystemPrompt appends the system prompt to the memory and saves it to the database
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

// GenerateAndUpdateSessionSummary generates a short summary of the conversation and updates the session
func (e *ExecutorService) GenerateAndUpdateSessionSummary() {
	sessionID := e.GetSessionID()
	if sessionID == 0 {
		return
	}

	// Get conversation messages (excluding system messages)
	messages, err := e.GetMemoryConversation()
	if err != nil {
		slog.Error("Failed to get conversation for summary", "error", err)
		return
	}

	// Build conversation text for summarization (only human and AI messages)
	var conversationText strings.Builder
	for _, msg := range messages {
		role := msg.GetType()

		switch role {
		case llms.ChatMessageTypeSystem:
			continue
		case llms.ChatMessageTypeHuman:
			conversationText.WriteString("User: " + msg.GetContent() + "\n")
		case llms.ChatMessageTypeAI:
			conversationText.WriteString("Assistant: " + msg.GetContent() + "\n")
		}
	}

	// Get summary prompt from config
	chatConfig := config.GetChatConfig()
	summaryPrompt := fmt.Sprintf(chatConfig.SummaryPrompt, conversationText.String())

	// Generate summary using the LLM
	msgContent := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, summaryPrompt),
	}

	summaryResponse, err := e.llm.GenerateContent(e.ctx, msgContent, llms.WithMaxLength(chatConfig.SummaryMaxLength))
	if err != nil {
		slog.Error("Failed to generate session summary", "error", err)
		return
	}

	// Extract summary text from response
	var summaryText string
	if len(summaryResponse.Choices) > 0 {
		summaryText = summaryResponse.Choices[0].Content
	}

	if summaryText == "" {
		slog.Error("Empty summary response from LLM")
		return
	}

	summaryText = strings.TrimSpace(summaryText)
	summaryText = strings.Trim(summaryText, "\"'")

	// Update session summary in database
	err = e.sessionLog.UpdateSessionSummary(sessionID, summaryText)
	if err != nil {
		slog.Error("Failed to update session summary", "error", err)
		return
	}

	slog.Debug("Session summary generated", "sessionID", sessionID, "summary", summaryText)
}

// SearchRecipeNames returns recipe names that match the given query (case-insensitive prefix match).
func (e *ExecutorService) SearchRecipeNames(query string) []RecipeSuggestion {
	allRecipes, err := e.cookbook.AllRecipes()
	if err != nil {
		slog.Error("Failed to fetch recipes for autocomplete", "error", err)
		return nil
	}

	queryLower := strings.ToLower(query)
	var results []RecipeSuggestion
	const maxResults = 8

	for _, r := range allRecipes {
		if strings.Contains(strings.ToLower(r.RecipeName), queryLower) {
			results = append(results, RecipeSuggestion{
				ID:   r.RecipeID,
				Name: r.RecipeName,
			})
			if len(results) >= maxResults {
				break
			}
		}
	}
	return results
}

// GetFullRecipe returns the full recipe content for a given ID.
func (e *ExecutorService) GetFullRecipe(recipeID uint) (string, error) {
	recipe, err := e.cookbook.GetFullRecipe(recipeID)
	if err != nil {
		return "", fmt.Errorf("recipe not found: %w", err)
	}
	return recipe.FormatRecipeMarkdown(), nil
}

// RecipeSuggestion holds a recipe name and ID for autocomplete.
type RecipeSuggestion struct {
	ID   uint
	Name string
}
