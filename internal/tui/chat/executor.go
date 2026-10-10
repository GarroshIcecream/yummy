package chat

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/llms"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	langllms "github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
)

// ExecutorService provides agent-based LLM interactions using langchaingo executor
type ExecutorService struct {
	ai           *llms.LLMService
	executor     *agents.Executor
	sessionLog   *db.SessionLog
	cookbook     *db.CookBook
	sessionStats db.SessionStats
	systemPrompt string

	// mu guards memSnapshot and the executor field above: the executor is
	// reassigned on the main goroutine (SetModelByName) while the generation
	// goroutine reads it.
	//
	// memSnapshot holds a frozen copy of conversation messages taken before
	// generation starts. While non-nil, GetMemoryConversation returns this
	// snapshot instead of reading from the (concurrently mutated) memory.
	mu          sync.Mutex
	memSnapshot []langllms.ChatMessage

	// toolCalls holds, per assistant message (by its position among the
	// conversation's assistant messages), the tool calls that produced it, so
	// the thread keeps showing them after the conversation moves on. Guarded
	// by mu; persisted in SessionMessage.ToolCalls and restored by LoadSession.
	toolCalls map[int][]llms.ToolCallEvent
}

// getExecutor returns the current agent executor under the lock so that reads
// from the generation goroutine don't race with SetModelByName reassigning it.
func (e *ExecutorService) getExecutor() *agents.Executor {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.executor
}

// NewExecutorService creates a new executor service instance
func NewExecutorService(ai *llms.LLMService, cookbook *db.CookBook, sessionLog *db.SessionLog) (*ExecutorService, error) {
	chatConfig := config.GetChatConfig()

	mem := memory.NewConversationBuffer(
		memory.WithInputKey("input"),
		memory.WithOutputKey("output"),
	)

	executor := ai.NewAgent(mem)

	service := &ExecutorService{
		ai:           ai,
		executor:     executor,
		cookbook:     cookbook,
		sessionLog:   sessionLog,
		sessionStats: db.SessionStats{},
		systemPrompt: chatConfig.SystemPrompt,
	}

	return service, nil
}

func (e *ExecutorService) GetSystemPrompt() string {
	return e.systemPrompt
}

// GetStreamCh returns the read-only channel that emits streaming chunks from the LLM.
func (e *ExecutorService) GetStreamCh() <-chan string {
	return e.ai.GetStreamCh()
}

func (e *ExecutorService) GetToolCallCh() <-chan llms.ToolCallEvent {
	return e.ai.GetToolCallCh()
}

// GetDoneCh returns a channel closed when the stream/tool-call channels are
// replaced, signalling listeners on the old channels to exit.
func (e *ExecutorService) GetDoneCh() <-chan struct{} {
	return e.ai.GetDoneCh()
}

// CancelStreaming cancels the in-flight LLM generation (if any), drains and
// replaces the stream channel, and creates a fresh cancellable context so the
// next generation can proceed.
func (e *ExecutorService) CancelStreaming() {
	e.ai.CancelStreaming()
}

func (e *ExecutorService) GetMemory() *memory.ConversationBuffer {
	return e.getExecutor().GetMemory().(*memory.ConversationBuffer)
}

func (e *ExecutorService) SaveMessage(message string, role langllms.ChatMessageType) error {
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
	err := e.SaveMessage(message, langllms.ChatMessageTypeHuman)
	if err != nil {
		slog.Error("Failed to register message", "error", err)
		return err
	}

	e.ai.ResetCallbackState()
	e.ai.TakeToolCalls() // drop anything left over from a previous run

	// Snapshot the current conversation so that renders during generation
	// read a stable copy instead of the concurrently-mutated memory.
	ctx := e.ai.GetContext()
	msgs, err := e.GetMemory().ChatHistory.Messages(ctx)
	if err != nil {
		slog.Error("Failed to snapshot memory", "error", err)
		// Non-fatal — we'll just read live memory (racy but won't crash).
	} else {
		snap := make([]langllms.ChatMessage, len(msgs))
		copy(snap, msgs)
		e.mu.Lock()
		e.memSnapshot = snap
		e.mu.Unlock()
	}

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
	slog.Debug("Generating response with executor", "model", e.ai.GetModelName(), "input", promptMessage)

	result, err := chains.Run(e.ai.GetContext(), e.getExecutor(), promptMessage)

	// Generation done — clear the snapshot so renders read live memory again.
	e.mu.Lock()
	e.memSnapshot = nil
	e.mu.Unlock()

	if err != nil {
		slog.Error("Executor execution error", "error", err)
		return "", err
	}

	// If we have a compact display message, replace the augmented prompt in
	// the conversation memory so that re-renders show the clean version.
	if displayMessage != "" && displayMessage != promptMessage {
		e.replaceLastHumanMessage(displayMessage)
	}

	usage := e.ai.GetTokenUsage()
	slog.Debug("Generated response",
		"result", result,
		"prompt_tokens", usage.PromptTokens,
		"completion_tokens", usage.CompletionTokens,
		"total_tokens", usage.TotalTokens)

	calls := e.ai.TakeToolCalls()
	var encoded string
	if len(calls) > 0 {
		if b, err := json.Marshal(calls); err == nil {
			encoded = string(b)
		}
		// The executor has already appended this answer to memory, so it is the
		// last assistant message. Count before locking: GetMemory takes e.mu.
		index := e.aiMessageCount() - 1
		e.mu.Lock()
		if e.toolCalls == nil {
			e.toolCalls = map[int][]llms.ToolCallEvent{}
		}
		e.toolCalls[index] = calls
		e.mu.Unlock()
	}

	err = e.sessionLog.SaveSessionMessageWithToolCalls(
		e.sessionStats.SessionID, result, langllms.ChatMessageTypeAI,
		e.GetCurrentModelName(), usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, encoded,
	)
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
	ctx := e.ai.GetContext()
	msgs, err := e.GetMemory().ChatHistory.Messages(ctx)
	if err != nil {
		slog.Error("Failed to read memory for mention cleanup", "error", err)
		return
	}

	// Find the last human message and replace it.
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].GetType() == langllms.ChatMessageTypeHuman {
			msgs[i] = langllms.HumanChatMessage{Content: displayText}
			break
		}
	}

	// Rewrite the full conversation buffer.
	if err := e.SetMemory(msgs); err != nil {
		slog.Error("Failed to rewrite memory after mention cleanup", "error", err)
	}
}

func (e *ExecutorService) GetMemoryConversation() ([]langllms.ChatMessage, error) {
	e.mu.Lock()
	snap := e.memSnapshot
	e.mu.Unlock()
	if snap != nil {
		return snap, nil
	}
	ctx := e.ai.GetContext()
	messages, err := e.GetMemory().ChatHistory.Messages(ctx)
	if err != nil {
		slog.Error("Failed to get conversation", "error", err)
		return nil, err
	}
	return messages, nil
}

// ToolCallsFor returns the tool calls behind the n-th assistant message
// (0-based, counting only assistant messages) of the current conversation.
func (e *ExecutorService) ToolCallsFor(n int) []llms.ToolCallEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.toolCalls[n]
}

// aiMessageCount counts assistant messages in live memory. It must not be
// called with e.mu held (GetMemory takes it).
func (e *ExecutorService) aiMessageCount() int {
	msgs, err := e.GetMemory().ChatHistory.Messages(e.ai.GetContext())
	if err != nil {
		return 0
	}
	n := 0
	for _, m := range msgs {
		if m.GetType() == langllms.ChatMessageTypeAI {
			n++
		}
	}
	return n
}

// VisibleMessageCount returns the number of user and assistant messages in the
// current conversation (system and tool messages excluded).
func (e *ExecutorService) VisibleMessageCount() int {
	msgs, err := e.GetMemoryConversation()
	if err != nil {
		return 0
	}
	n := 0
	for _, m := range msgs {
		switch m.GetType() {
		case langllms.ChatMessageTypeHuman, langllms.ChatMessageTypeAI:
			n++
		}
	}
	return n
}

// GetProviderName returns the display name of the LLM provider.
func (e *ExecutorService) GetProviderName() string {
	return "Ollama"
}

// GetSessionID returns the current session ID
func (e *ExecutorService) GetSessionID() uint {
	return e.sessionStats.SessionID
}

func (e *ExecutorService) SetMemory(conversation []langllms.ChatMessage) error {
	ctx := e.ai.GetContext()
	err := e.ClearMemory()
	if err != nil {
		slog.Error("Failed to clear memory", "error", err)
		return err
	}
	for _, message := range conversation {
		err := e.GetMemory().ChatHistory.AddMessage(ctx, message)
		if err != nil {
			slog.Error("Failed to add message to memory", "error", err)
			return err
		}
	}
	return nil
}

func (e *ExecutorService) ClearMemory() error {
	ctx := e.ai.GetContext()
	err := e.GetMemory().Clear(ctx)
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
	e.setToolCalls(nil)
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
	e.setToolCalls(nil)
	return nil
}

func (e *ExecutorService) setToolCalls(calls map[int][]llms.ToolCallEvent) {
	e.mu.Lock()
	e.toolCalls = calls
	e.mu.Unlock()
}

// LoadSession loads a session into the executor service
func (e *ExecutorService) LoadSession(sessionID uint) error {
	sessionMessages, err := e.sessionLog.GetSessionMessages(sessionID)
	if err != nil {
		slog.Error("Failed to get session messages", "error", err)
		return err
	}

	chatMessages := make([]langllms.ChatMessage, len(sessionMessages))
	toolCalls := map[int][]llms.ToolCallEvent{}
	aiIndex := 0
	for i, msg := range sessionMessages {
		if msg.Role == string(langllms.ChatMessageTypeAI) {
			if msg.ToolCalls != "" {
				var calls []llms.ToolCallEvent
				if err := json.Unmarshal([]byte(msg.ToolCalls), &calls); err == nil {
					toolCalls[aiIndex] = calls
				}
			}
			aiIndex++
		}
		if msg.Role == string(langllms.ChatMessageTypeSystem) {
			chatMessages[i] = langllms.SystemChatMessage{Content: msg.Message}
		} else if msg.Role == string(langllms.ChatMessageTypeHuman) {
			chatMessages[i] = langllms.HumanChatMessage{Content: msg.Message}
		} else if msg.Role == string(langllms.ChatMessageTypeAI) {
			chatMessages[i] = langllms.AIChatMessage{Content: msg.Message}
		} else if msg.Role == string(langllms.ChatMessageTypeTool) {
			chatMessages[i] = langllms.ToolChatMessage{Content: msg.Message}
		}
	}

	// We need to preserve the system prompt so the LLM can remember previous messages
	err = e.SetMemory(chatMessages)
	if err != nil {
		slog.Error("Failed to set memory", "error", err)
		return err
	}

	// Set the model for the session
	err = e.SetModelByName(sessionMessages[0].ModelName)
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
	e.setToolCalls(toolCalls)
	return nil
}

// GetCurrentModelName returns the name of the current model
func (e *ExecutorService) GetCurrentModelName() string {
	return e.ai.GetModelName()
}

// GetInstalledModels returns the list of Ollama models available locally.
func (e *ExecutorService) GetInstalledModels() []string {
	return e.ai.GetInstalledModels()
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
	ctx := e.ai.GetContext()
	messages, err := e.GetMemory().ChatHistory.Messages(ctx)
	if err != nil {
		return false
	}

	for _, message := range messages {
		if message.GetType() == langllms.ChatMessageTypeSystem {
			return true
		}
	}
	return false
}

// SetModelByName sets the model for the executor service and recreates the agent.
func (e *ExecutorService) SetModelByName(modelName string) error {
	err := e.ai.SetModelByName(modelName)
	if err != nil {
		return err
	}

	// Recreate the agent executor with the new model (memory already has system prompt)
	mem := e.GetMemory()
	executor := e.ai.NewAgent(mem)
	e.mu.Lock()
	e.executor = executor
	e.mu.Unlock()

	return nil
}

// AppendSystemPrompt appends the system prompt to the memory and saves it to the database
func (e *ExecutorService) AppendSystemPrompt(systemPrompt string, sessionID uint) error {
	ctx := e.ai.GetContext()
	systemMessage := langllms.SystemChatMessage{Content: systemPrompt}
	err := e.GetMemory().ChatHistory.AddMessage(ctx, systemMessage)
	if err != nil {
		slog.Error("Failed to add system prompt to memory", "error", err)
		return err
	}

	err = e.sessionLog.SaveSessionMessage(sessionID, systemPrompt, langllms.ChatMessageTypeSystem, e.ai.GetModelName(), 0, 0, 0)
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
		case langllms.ChatMessageTypeSystem:
			continue
		case langllms.ChatMessageTypeHuman:
			conversationText.WriteString("User: " + msg.GetContent() + "\n")
		case langllms.ChatMessageTypeAI:
			conversationText.WriteString("Assistant: " + msg.GetContent() + "\n")
		}
	}

	// Get summary prompt from config
	chatConfig := config.GetChatConfig()
	summaryPrompt := fmt.Sprintf(chatConfig.SummaryPrompt, conversationText.String())

	// Generate summary using the LLM
	msgContent := []langllms.MessageContent{
		langllms.TextParts(langllms.ChatMessageTypeHuman, summaryPrompt),
	}

	summaryResponse, err := e.ai.GenerateContent(msgContent, langllms.WithMaxLength(chatConfig.SummaryMaxLength))
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

// GetOllamaStatus returns the Ollama service status.
func (e *ExecutorService) GetOllamaStatus() *llms.OllamaServiceStatus {
	return e.ai.GetOllamaStatus()
}

// GetAI returns the underlying LLMService.
func (e *ExecutorService) GetAI() *llms.LLMService {
	return e.ai
}

// RecipeSuggestion holds a recipe name and ID for autocomplete.
type RecipeSuggestion struct {
	ID   uint
	Name string
}
