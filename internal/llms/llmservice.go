package llms

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/GarroshIcecream/yummy/internal/config"
	"github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/llms/callbacks"
	llmtools "github.com/GarroshIcecream/yummy/internal/llms/tools"
	"github.com/tmc/langchaingo/agents"
	langllms "github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
)

// ToolCallEvent represents a tool invocation by the agent.
type ToolCallEvent struct {
	Tool  string
	Input string
}

// LLMService provides shared LLM infrastructure for both chat and cooking modes.
type LLMService struct {
	// mu guards the mutable fields (llm, modelName, ctx, cancelCtx and the
	// channels): they are swapped on the main goroutine while LLM goroutines
	// read them via the getters.
	mu        sync.Mutex
	llm       langllms.Model
	modelName string
	provider  Provider
	// models is the provider's model list, fetched once at construction.
	models          []string
	ctx             context.Context
	cancelCtx       context.CancelFunc
	callbackHandler *callbacks.DefaultAgentCallbackHandler
	streamCh        chan string
	toolCallCh      chan ToolCallEvent
	// doneCh is closed whenever streamCh/toolCallCh are replaced so that
	// listeners blocked on the old channels can terminate instead of leaking.
	doneCh chan struct{}
	// generation increments on every cancellation; it is embedded in the
	// generation's context so the callback handler can drop events from
	// cancelled runs that are still draining.
	generation    uint64
	toolManager   *llmtools.ToolManager
	maxIterations int

	// toolCallLog records the tool calls of the current generation so the
	// caller can persist them alongside the answer (TakeToolCalls). Guarded
	// by mu; reset whenever the channels are replaced.
	toolCallLog []ToolCallEvent
}

// NewLLMService creates the shared LLM service on the provider selected in the
// chat config.
func NewLLMService(cookbook *db.CookBook) (*LLMService, error) {
	provider, err := NewProvider(config.GetChatConfig(), config.DataDir())
	if err != nil {
		slog.Error("Failed to create LLM provider", "error", err)
		return nil, err
	}
	return NewLLMServiceWithProvider(cookbook, provider)
}

// NewLLMServiceWithProvider creates the shared LLM service on the given
// provider, starting with the provider's default model.
func NewLLMServiceWithProvider(cookbook *db.CookBook, provider Provider) (*LLMService, error) {
	const initialGeneration = 1
	chatConfig := config.GetChatConfig()

	models, err := provider.ListModels(context.Background())
	if err != nil {
		slog.Error("Failed to list models", "provider", provider.Name(), "error", err)
		return nil, err
	}

	modelName := provider.DefaultModel()
	llm, err := provider.NewModel(modelName)
	if err != nil {
		slog.Error("Failed to create LLM", "error", err)
		return nil, err
	}

	toolManager := llmtools.NewToolManager(cookbook)
	ctx, cancel := context.WithCancel(callbacks.WithGeneration(context.Background(), initialGeneration))

	streamCh := make(chan string, 64)
	toolCallCh := make(chan ToolCallEvent, 8)
	doneCh := make(chan struct{})
	callbackHandler := callbacks.NewDefaultAgentCallbackHandler(nil)
	callbackHandler.SetGeneration(initialGeneration)

	s := &LLMService{
		llm:             llm,
		modelName:       modelName,
		provider:        provider,
		models:          models,
		ctx:             ctx,
		cancelCtx:       cancel,
		callbackHandler: callbackHandler,
		streamCh:        streamCh,
		toolCallCh:      toolCallCh,
		doneCh:          doneCh,
		generation:      initialGeneration,
		toolManager:     toolManager,
		maxIterations:   chatConfig.MaxIterations,
	}
	s.wireCallbacks(callbackHandler, streamCh, toolCallCh, doneCh)
	return s, nil
}

// TakeToolCalls returns the tool calls recorded since the last call (or since
// the channels were last replaced) and clears the record.
func (s *LLMService) TakeToolCalls() []ToolCallEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	calls := s.toolCallLog
	s.toolCallLog = nil
	return calls
}

func (s *LLMService) recordToolCall(ev ToolCallEvent) {
	s.mu.Lock()
	s.toolCallLog = append(s.toolCallLog, ev)
	s.mu.Unlock()
}

func (s *LLMService) GetStreamCh() <-chan string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.streamCh
}

// GetToolCallCh returns a channel that emits tool call events from the agent.
func (s *LLMService) GetToolCallCh() <-chan ToolCallEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.toolCallCh
}

// GetDoneCh returns a channel that is closed when the current stream/tool-call
// channels are replaced. Listeners should select on it alongside the data
// channels so they terminate instead of blocking forever on orphaned channels.
func (s *LLMService) GetDoneCh() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.doneCh
}

// replaceChannels swaps the stream and tool-call channels for fresh ones,
// re-points the callbacks at them, and closes the previous done channel so
// listeners still blocked on the old channels can exit. We intentionally do
// NOT close the old data channels to avoid send-on-closed-channel panics from
// in-flight LLM goroutines.
func (s *LLMService) replaceChannels() {
	s.mu.Lock()
	close(s.doneCh)
	s.doneCh = make(chan struct{})
	s.streamCh = make(chan string, 64)
	s.toolCallCh = make(chan ToolCallEvent, 8)
	s.toolCallLog = nil
	streamCh, toolCallCh, doneCh := s.streamCh, s.toolCallCh, s.doneCh
	s.mu.Unlock()
	s.wireCallbacks(s.callbackHandler, streamCh, toolCallCh, doneCh)
}

// wireCallbacks points the handler's callbacks at the given channels. The
// channels are captured as locals so the LLM goroutine never reads the
// channel fields (avoiding data races with channel replacement; the tool-call
// record goes through the mutex-guarded recordToolCall), and the stream send
// aborts once done is closed so a stalled generation can't block forever on
// an orphaned channel.
func (s *LLMService) wireCallbacks(handler *callbacks.DefaultAgentCallbackHandler, streamCh chan string, toolCallCh chan ToolCallEvent, doneCh chan struct{}) {
	handler.SetStreamingCallback(func(chunk string) {
		select {
		case streamCh <- chunk:
		case <-doneCh:
		}
	})
	handler.SetToolCallCallback(func(tool, input string) {
		s.recordToolCall(ToolCallEvent{Tool: tool, Input: input})
		select {
		case toolCallCh <- ToolCallEvent{Tool: tool, Input: input}:
		default:
		}
	})
}

func (s *LLMService) GetContext() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ctx
}

// CancelStreaming cancels the in-flight LLM generation, replaces the stream
// channel, and creates a fresh cancellable context. Safe to call even if no
// generation is in progress.
func (s *LLMService) CancelStreaming() {
	s.mu.Lock()
	s.cancelCtx()
	// Bump the generation and embed it in the fresh context: the cancelled
	// run may drain a few more callback events, and the handler drops any
	// event whose context carries a stale generation ID.
	s.generation++
	gen := s.generation
	s.ctx, s.cancelCtx = context.WithCancel(callbacks.WithGeneration(context.Background(), gen))
	s.mu.Unlock()

	s.callbackHandler.SetGeneration(gen)

	// Replace channels so no new sends target old channels.
	s.replaceChannels()
}

func (s *LLMService) GetLLM() langllms.Model {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.llm
}

func (s *LLMService) GetModelName() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.modelName
}

func (s *LLMService) GetToolManager() *llmtools.ToolManager {
	return s.toolManager
}

// Provider returns the LLM provider backing this service.
func (s *LLMService) Provider() Provider {
	return s.provider
}

// GetInstalledModels returns the models available on the active provider.
func (s *LLMService) GetInstalledModels() []string {
	return s.models
}

// ResetCallbackState resets token usage, stream buffer, and replaces both the
// stream and tool-call channels before a new generation. Replacing channels
// detaches any stale listener goroutines left over from previous generations
// so they can't consume events meant for the current one; closing the done
// channel lets them exit instead of leaking.
func (s *LLMService) ResetCallbackState() {
	s.callbackHandler.ResetTokenUsage()
	s.callbackHandler.ResetStreamBuffer()

	// Replace channels so old listeners detach from defunct channels.
	s.replaceChannels()
}

// GetTokenUsage returns the accumulated token usage from the last generation.
func (s *LLMService) GetTokenUsage() callbacks.TokenUsage {
	return s.callbackHandler.GetTokenUsage()
}

func (s *LLMService) SetModelByName(modelName string) error {
	if !slices.Contains(s.models, modelName) {
		return fmt.Errorf("model %s is not available on %s", modelName, s.provider.DisplayName())
	}

	llm, err := s.provider.NewModel(modelName)
	if err != nil {
		return fmt.Errorf("failed to create LLM: %w", err)
	}

	s.mu.Lock()
	s.llm = llm
	s.modelName = modelName
	s.mu.Unlock()
	return nil
}

// NewAgent creates a conversational agent executor with tools, using the given
// conversation memory. The caller owns the memory and should seed it with a
// system prompt via AddSystemPrompt if needed.
func (s *LLMService) NewAgent(mem *memory.ConversationBuffer) *agents.Executor {
	handler := s.callbackHandler
	tools := s.toolManager.GetTools()

	return agents.NewExecutor(
		agents.NewConversationalAgent(s.GetLLM(), tools, agents.WithCallbacksHandler(handler)),
		agents.WithMaxIterations(s.maxIterations),
		agents.WithMemory(mem),
		agents.WithReturnIntermediateSteps(),
		agents.WithCallbacksHandler(handler),
	)
}

// AddSystemPrompt injects a system prompt into the given conversation memory.
func (s *LLMService) AddSystemPrompt(mem *memory.ConversationBuffer, systemPrompt string) error {
	if systemPrompt == "" {
		return nil
	}
	return mem.ChatHistory.AddMessage(s.GetContext(), langllms.SystemChatMessage{Content: systemPrompt})
}

// GenerateContent calls the LLM directly (bypassing the agent) for one-off
// generations like session summaries.
func (s *LLMService) GenerateContent(msgs []langllms.MessageContent, opts ...langllms.CallOption) (*langllms.ContentResponse, error) {
	return s.GetLLM().GenerateContent(s.GetContext(), msgs, opts...)
}
