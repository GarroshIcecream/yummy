package callbacks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

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

// TokenUsage tracks token consumption
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// DefaultAgentCallbackHandler is the default implementation of AgentCallbackHandler
// It logs all agent events and optionally sends status updates to the UI
type DefaultAgentCallbackHandler struct {
	onStatusFunc      StatusUpdateFunc
	streamingCallback func(chunk string)
	tokenUsage        TokenUsage

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

// SetStreamingCallback sets the function called for each streaming chunk from the LLM.
func (h *DefaultAgentCallbackHandler) SetStreamingCallback(cb func(chunk string)) {
	h.streamingCallback = cb
}

// GetTokenUsage returns the accumulated token usage
func (h *DefaultAgentCallbackHandler) GetTokenUsage() TokenUsage {
	return h.tokenUsage
}

// ResetTokenUsage resets the token usage counter
func (h *DefaultAgentCallbackHandler) ResetTokenUsage() {
	h.tokenUsage = TokenUsage{}
}

// sendStatus sends a status update to the UI if the callback function is set
func (h *DefaultAgentCallbackHandler) sendStatus(status string) {
	if h.onStatusFunc != nil {
		h.onStatusFunc(status)
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
	h.tokenUsage.PromptTokens += promptTokens
	h.tokenUsage.CompletionTokens += completionTokens
	h.tokenUsage.TotalTokens += totalTokens

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
	slog.Debug("Agent Callback: Agent Action",
		"tool_id", action.ToolID,
		"tool", action.Tool,
		"tool_input", action.ToolInput,
		"log", action.Log,
	)

	h.sendStatus(fmt.Sprintf("🎯 Action: %s", action.Tool))
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
	h.streamBuffer = ""
	h.streamingAnswer = false
}

// _aiMarker is the token the conversational agent uses to prefix its final answer.
const _aiMarker = "AI:"

// HandleStreamingFunc handles streaming chunks from the LLM.
// The conversational agent emits "Thought: … AI: <answer>" — we buffer until
// we see the "AI:" marker and only forward the answer portion to the UI.
func (h *DefaultAgentCallbackHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	if len(chunk) == 0 {
		return
	}

	chunkStr := string(chunk)
	slog.Debug("Agent Callback: Streaming Chunk", "size", len(chunk), "content", chunkStr)

	if h.streamingAnswer {
		// Already past the marker — forward everything directly.
		if h.streamingCallback != nil {
			h.streamingCallback(chunkStr)
		}
		return
	}

	// Still buffering the thought process; accumulate and check for marker.
	h.streamBuffer += chunkStr

	if idx := strings.Index(h.streamBuffer, _aiMarker); idx >= 0 {
		h.streamingAnswer = true
		// Everything after "AI:" is the beginning of the real answer.
		answer := h.streamBuffer[idx+len(_aiMarker):]
		h.streamBuffer = "" // free memory
		if len(strings.TrimSpace(answer)) > 0 && h.streamingCallback != nil {
			h.streamingCallback(strings.TrimLeft(answer, " "))
		}
	}
}
