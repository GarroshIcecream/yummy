package callbacks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

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

// DefaultAgentCallbackHandler is the default implementation of AgentCallbackHandler
// It logs all agent events and optionally sends status updates to the UI
type DefaultAgentCallbackHandler struct {
	onStatusFunc      StatusUpdateFunc
	streamingCallback func(chunk string)
}

// NewDefaultAgentCallbackHandler creates a new default callback handler
func NewDefaultAgentCallbackHandler(statusFunc StatusUpdateFunc) *DefaultAgentCallbackHandler {
	return &DefaultAgentCallbackHandler{
		onStatusFunc: statusFunc,
	}
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
	if len(res.Choices) > 0 {
		stopReason = res.Choices[0].StopReason
	}

	slog.Debug("Agent Callback: LLM Generate Content End",
		"choices", len(res.Choices),
		"stop_reason", stopReason)

	content := res.Choices[0].Content
	slog.Debug("Agent Callback: LLM Response Preview", "content", content)
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

// HandleStreamingFunc handles streaming chunks from the LLM
func (h *DefaultAgentCallbackHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	if len(chunk) > 0 {
		slog.Debug("Agent Callback: Streaming Chunk", "size", len(chunk), "content", string(chunk))
		chunkStr := string(chunk)

		// Send to streaming callback if available
		if h.streamingCallback != nil {
			h.streamingCallback(chunkStr)
		}

		// Also send as status update
		h.sendStatus(chunkStr)
	}
}
