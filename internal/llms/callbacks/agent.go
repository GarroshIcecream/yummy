package callbacks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type AgentCallbackHandler interface {
	HandleText(ctx context.Context, text string)
	HandleLLMStart(ctx context.Context, prompts []string)
	HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent)
	HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse)
	HandleLLMError(ctx context.Context, err error)
	HandleChainStart(ctx context.Context, inputs map[string]any)
	HandleChainEnd(ctx context.Context, outputs map[string]any)
	HandleChainError(ctx context.Context, err error)
	HandleToolStart(ctx context.Context, input string)
	HandleToolEnd(ctx context.Context, output string)
	HandleToolError(ctx context.Context, err error)
	HandleAgentAction(ctx context.Context, action schema.AgentAction)
	HandleAgentFinish(ctx context.Context, finish schema.AgentFinish)
	HandleRetrieverStart(ctx context.Context, query string)
	HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document)
	HandleStreamingFunc(ctx context.Context, chunk []byte)
}

// StatusUpdateFunc is a callback function type for sending status updates to the UI
type StatusUpdateFunc func(status string)

// generationKey is the context key under which the active generation ID is
// stored. langchaingo threads the generation's context into every callback,
// so tagging it lets the handler drop events from cancelled generations that
// are still draining.
type generationKey struct{}

// WithGeneration returns a context tagged with the given generation ID.
func WithGeneration(ctx context.Context, gen uint64) context.Context {
	return context.WithValue(ctx, generationKey{}, gen)
}

// generationFromContext extracts the generation ID from a context, if present.
func generationFromContext(ctx context.Context) (uint64, bool) {
	gen, ok := ctx.Value(generationKey{}).(uint64)
	return gen, ok
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// DefaultAgentCallbackHandler is the default implementation of AgentCallbackHandler
// It logs all agent events and optionally sends status updates to the UI
type DefaultAgentCallbackHandler struct {
	// mu guards the mutable fields below: callbacks are swapped from the main
	// (Bubble Tea) goroutine while Handle* methods run on the LLM goroutine.
	// Callbacks are always invoked OUTSIDE the lock so a blocking send in a
	// callback can never deadlock against a callback swap.
	mu                sync.Mutex
	onStatusFunc      StatusUpdateFunc
	streamingCallback func(chunk string)
	toolCallCallback  func(tool, input string)
	tokenUsage        TokenUsage

	// generation is the ID of the currently active generation. Events whose
	// context carries a different generation ID (i.e. from a cancelled run
	// that is still draining) are dropped so they can't contaminate the
	// active generation's stream buffer or channels.
	generation uint64

	// streamBuffer accumulates raw streaming output so we can detect the
	// "AI:" marker and only forward the actual answer to the UI.
	streamBuffer    string
	streamingAnswer bool
}

// NewDefaultAgentCallbackHandler creates a new default callback handler
func NewDefaultAgentCallbackHandler(statusFunc StatusUpdateFunc) *DefaultAgentCallbackHandler {
	return &DefaultAgentCallbackHandler{
		onStatusFunc: statusFunc,
		tokenUsage:   TokenUsage{},
	}
}

// SetGeneration marks the given generation ID as current; events tagged with
// any other generation are dropped.
func (h *DefaultAgentCallbackHandler) SetGeneration(gen uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.generation = gen
}

// isStaleGeneration reports whether the context belongs to a generation other
// than the current one. Untagged contexts are never considered stale.
func (h *DefaultAgentCallbackHandler) isStaleGeneration(ctx context.Context) bool {
	gen, ok := generationFromContext(ctx)
	if !ok {
		return false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return gen != h.generation
}

// SetStatusFunc replaces the status update callback.
func (h *DefaultAgentCallbackHandler) SetStatusFunc(fn StatusUpdateFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onStatusFunc = fn
}

// SetStreamingCallback sets the function called for each streaming chunk from the LLM.
func (h *DefaultAgentCallbackHandler) SetStreamingCallback(cb func(chunk string)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.streamingCallback = cb
}

// SetToolCallCallback sets the function called when the agent invokes a tool.
func (h *DefaultAgentCallbackHandler) SetToolCallCallback(cb func(tool, input string)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.toolCallCallback = cb
}

// GetTokenUsage returns the accumulated token usage
func (h *DefaultAgentCallbackHandler) GetTokenUsage() TokenUsage {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.tokenUsage
}

// ResetTokenUsage resets the token usage counter
func (h *DefaultAgentCallbackHandler) ResetTokenUsage() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.tokenUsage = TokenUsage{}
}

// sendStatus sends a status update to the UI if the callback function is set
func (h *DefaultAgentCallbackHandler) sendStatus(status string) {
	h.mu.Lock()
	fn := h.onStatusFunc
	h.mu.Unlock()
	if fn != nil {
		fn(status)
	}
}

// HandleText handles text output from the agent
func (h *DefaultAgentCallbackHandler) HandleText(ctx context.Context, text string) {
	slog.Debug("Agent Callback: Text", "text", text)
	h.sendStatus(fmt.Sprintf("💬 %s", text))
}

// HandleLLMStart handles the start of LLM processing
func (h *DefaultAgentCallbackHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	slog.Debug("Agent Callback: LLM Start", "prompt_count", len(prompts))
	h.sendStatus("🤖 Thinking...")
}

// HandleLLMGenerateContentStart handles the start of LLM content generation
func (h *DefaultAgentCallbackHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	slog.Debug("Agent Callback: LLM Generate Content Start", "message_count", len(ms))
	for i, msg := range ms {
		slog.Debug("Agent Callback: Message Content", "index", i, "role", msg.Role, "parts", len(msg.Parts))
	}
	h.sendStatus("🤖 Generating response...")
}

// HandleLLMGenerateContentEnd handles the end of LLM content generation
func (h *DefaultAgentCallbackHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	if res == nil {
		slog.Debug("Agent Callback: LLM Generate Content End", "response", "nil")
		return
	}
	// Don't let a draining cancelled run skew the next generation's counters.
	if h.isStaleGeneration(ctx) {
		return
	}

	stopReason := ""
	var promptTokens, completionTokens, totalTokens int

	// Extract token usage from GenerationInfo (Ollama provides this)
	for _, choice := range res.Choices {
		if choice.GenerationInfo != nil {
			stopReason = choice.StopReason

			// Ollama uses "prompt_eval_count" and "eval_count" keys
			if val, ok := choice.GenerationInfo["prompt_eval_count"]; ok {
				switch v := val.(type) {
				case int:
					promptTokens += v
				case float64:
					promptTokens += int(v)
				case int64:
					promptTokens += int(v)
				}
			}

			if val, ok := choice.GenerationInfo["eval_count"]; ok {
				switch v := val.(type) {
				case int:
					completionTokens += v
				case float64:
					completionTokens += int(v)
				case int64:
					completionTokens += int(v)
				}
			}
		}
	}

	// Calculate total if not provided directly
	totalTokens = promptTokens + completionTokens

	// Accumulate token usage
	h.mu.Lock()
	h.tokenUsage.PromptTokens += promptTokens
	h.tokenUsage.CompletionTokens += completionTokens
	h.tokenUsage.TotalTokens += totalTokens
	h.mu.Unlock()

	slog.Debug("Agent Callback: LLM Generate Content End",
		"choices", len(res.Choices),
		"stop_reason", stopReason,
		"prompt_tokens", promptTokens,
		"completion_tokens", completionTokens,
		"total_tokens", totalTokens,
		"generation_info", res.Choices[0].GenerationInfo)

	if len(res.Choices) > 0 {
		content := res.Choices[0].Content
		slog.Debug("Agent Callback: LLM Response Preview", "content", content)
	}
}

// HandleLLMError handles LLM errors
func (h *DefaultAgentCallbackHandler) HandleLLMError(ctx context.Context, err error) {
	slog.Error("Agent Callback: LLM Error", "error", err)
	h.sendStatus(fmt.Sprintf("❌ LLM Error: %v", err))
}

// HandleChainStart handles the start of a chain execution
func (h *DefaultAgentCallbackHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	inputJSON, _ := json.Marshal(inputs)
	slog.Debug("Agent Callback: Chain Start", "inputs", string(inputJSON))
	h.sendStatus("⛓️  Starting chain...")
}

// HandleChainEnd handles the end of a chain execution
func (h *DefaultAgentCallbackHandler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	outputJSON, _ := json.Marshal(outputs)
	slog.Debug("Agent Callback: Chain End", "outputs", string(outputJSON))
	h.sendStatus("✅ Chain completed")
}

// HandleChainError handles chain execution errors
func (h *DefaultAgentCallbackHandler) HandleChainError(ctx context.Context, err error) {
	slog.Error("Agent Callback: Chain Error", "error", err)
	h.sendStatus(fmt.Sprintf("❌ Chain Error: %v", err))
}

// HandleToolStart handles the start of a tool execution
func (h *DefaultAgentCallbackHandler) HandleToolStart(ctx context.Context, input string) {
	slog.Debug("Agent Callback: Tool Start", "input", input)
	h.sendStatus(fmt.Sprintf("🔧 Using tool: %s", input))
}

// HandleToolEnd handles the end of a tool execution
func (h *DefaultAgentCallbackHandler) HandleToolEnd(ctx context.Context, output string) {
	slog.Debug("Agent Callback: Tool End", "output_length", len(output))
	h.sendStatus("✅ Tool completed")
}

// HandleToolError handles tool execution errors
func (h *DefaultAgentCallbackHandler) HandleToolError(ctx context.Context, err error) {
	slog.Error("Agent Callback: Tool Error", "error", err)
	h.sendStatus(fmt.Sprintf("❌ Tool Error: %v", err))
}

// HandleAgentAction handles agent actions (tool calls)
func (h *DefaultAgentCallbackHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	if h.isStaleGeneration(ctx) {
		return
	}
	slog.Debug("Agent Callback: Agent Action",
		"tool_id", action.ToolID,
		"tool", action.Tool,
		"tool_input", action.ToolInput,
		"log", action.Log,
	)

	h.sendStatus(fmt.Sprintf("🎯 Action: %s", action.Tool))
	h.mu.Lock()
	cb := h.toolCallCallback
	h.mu.Unlock()
	if cb != nil {
		cb(action.Tool, action.ToolInput)
	}
}

// HandleAgentFinish handles agent completion
func (h *DefaultAgentCallbackHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	outputJSON, _ := json.Marshal(finish.ReturnValues)
	slog.Debug("Agent Callback: Agent Finish",
		"return_values", string(outputJSON),
		"log", finish.Log)

	h.sendStatus("✅ Agent finished")
}

// HandleRetrieverStart handles the start of a retrieval operation
func (h *DefaultAgentCallbackHandler) HandleRetrieverStart(ctx context.Context, query string) {
	slog.Debug("Agent Callback: Retriever Start", "query", query)
	h.sendStatus(fmt.Sprintf("🔍 Searching: %s", query))
}

// HandleRetrieverEnd handles the end of a retrieval operation
func (h *DefaultAgentCallbackHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	slog.Debug("Agent Callback: Retriever End", "query", query, "document_count", len(documents))
	h.sendStatus(fmt.Sprintf("📚 Found %d documents", len(documents)))
}

// ResetStreamBuffer resets the streaming filter state so the next generation
// starts fresh. Call this before each new LLM request.
func (h *DefaultAgentCallbackHandler) ResetStreamBuffer() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.streamBuffer = ""
	h.streamingAnswer = false
}

// _aiMarker is the token the conversational agent uses to prefix its final answer.
const _aiMarker = "AI:"

// HandleStreamingFunc handles streaming chunks from the LLM.
// The conversational agent emits "Thought: … AI: <answer>" — we buffer until
// we see the "AI:" marker and only forward the answer portion to the UI.
func (h *DefaultAgentCallbackHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	if len(chunk) == 0 || h.isStaleGeneration(ctx) {
		return
	}

	chunkStr := string(chunk)
	slog.Debug("Agent Callback: Streaming Chunk", "size", len(chunk), "content", chunkStr)

	// Mutate buffer state under the lock, but invoke the callback outside it
	// so a blocking channel send can't deadlock against a callback swap.
	h.mu.Lock()
	cb := h.streamingCallback
	var toSend string
	switch {
	case h.streamingAnswer:
		// Already past the marker — forward everything directly.
		toSend = chunkStr
	default:
		// Still buffering the thought process; accumulate and check for marker.
		h.streamBuffer += chunkStr
		if idx := strings.Index(h.streamBuffer, _aiMarker); idx >= 0 {
			h.streamingAnswer = true
			// Everything after "AI:" is the beginning of the real answer.
			answer := h.streamBuffer[idx+len(_aiMarker):]
			h.streamBuffer = "" // free memory
			if len(strings.TrimSpace(answer)) > 0 {
				toSend = strings.TrimLeft(answer, " ")
			}
		}
	}
	h.mu.Unlock()

	if toSend != "" && cb != nil {
		cb(toSend)
	}
}
