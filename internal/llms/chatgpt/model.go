// Package chatgpt is a langchaingo model backed by a ChatGPT subscription:
// it calls the ChatGPT Codex backend (an OpenAI Responses API endpoint) with
// the user's OAuth session from internal/auth/chatgpt.
package chatgpt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	auth "github.com/GarroshIcecream/yummy/internal/auth/chatgpt"
	"github.com/tmc/langchaingo/callbacks"
	langllms "github.com/tmc/langchaingo/llms"
)

// DefaultEndpoint is the Responses endpoint for ChatGPT-plan accounts.
const DefaultEndpoint = auth.CodexBaseURL + "/responses"

// TokenProvider supplies the OAuth session. *auth.TokenSource implements it.
type TokenProvider interface {
	Token(ctx context.Context) (*auth.Tokens, error)
	ForceRefresh(ctx context.Context) (*auth.Tokens, error)
}

// LLM is a langchaingo model served by a ChatGPT subscription.
type LLM struct {
	// CallbacksHandler, when set, receives the same LLM lifecycle events
	// langchaingo's own providers emit.
	CallbacksHandler callbacks.Handler

	model     string
	endpoint  string
	modelsURL string
	http      *http.Client
	tokens    TokenProvider
}

var _ langllms.Model = (*LLM)(nil)

// Option configures an LLM.
type Option func(*LLM)

// WithEndpoint overrides the Responses endpoint URL.
func WithEndpoint(url string) Option { return func(l *LLM) { l.endpoint = url } }

// WithHTTPClient overrides the HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(l *LLM) { l.http = c } }

// New returns a model that serves the given ChatGPT model ID.
func New(model string, tokens TokenProvider, opts ...Option) *LLM {
	l := &LLM{
		model:    model,
		endpoint: DefaultEndpoint,
		http:     &http.Client{}, // streaming responses: no overall timeout, ctx bounds it
		tokens:   tokens,
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

// Call implements langllms.Model.
func (l *LLM) Call(ctx context.Context, prompt string, opts ...langllms.CallOption) (string, error) {
	return langllms.GenerateFromSinglePrompt(ctx, l, prompt, opts...)
}

// GenerateContent implements langllms.Model. Temperature, max tokens and
// stop words are not sent (the backend rejects them); stop words are
// applied client-side instead.
func (l *LLM) GenerateContent(ctx context.Context, messages []langllms.MessageContent, options ...langllms.CallOption) (*langllms.ContentResponse, error) {
	if l.CallbacksHandler != nil {
		l.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}
	resp, err := l.generate(ctx, messages, options...)
	if err != nil {
		if l.CallbacksHandler != nil {
			l.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}
	if l.CallbacksHandler != nil {
		l.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, resp)
	}
	return resp, nil
}

func (l *LLM) generate(ctx context.Context, messages []langllms.MessageContent, options ...langllms.CallOption) (*langllms.ContentResponse, error) {
	opts := langllms.CallOptions{}
	for _, o := range options {
		o(&opts)
	}
	body, err := json.Marshal(buildRequest(l.model, messages, opts.JSONMode))
	if err != nil {
		return nil, err
	}

	tok, err := l.tokens.Token(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := l.post(ctx, body, tok)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		if tok, err = l.tokens.ForceRefresh(ctx); err != nil {
			return nil, err
		}
		if resp, err = l.post(ctx, body, tok); err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			_ = resp.Body.Close()
			return nil, errors.New("ChatGPT rejected the session; run `yummy login chatgpt` to sign in again")
		}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("ChatGPT request failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	res, err := readStream(ctx, resp.Body, opts.StopWords, opts.StreamingFunc)
	if err != nil {
		return nil, err
	}
	return &langllms.ContentResponse{Choices: []*langllms.ContentChoice{{
		Content:    res.text,
		StopReason: res.stopReason,
		GenerationInfo: map[string]any{
			// Keys read by internal/llms/callbacks (Ollama naming)...
			"prompt_eval_count": res.inputTokens,
			"eval_count":        res.outputTokens,
			// ...and langchaingo's standardized ones.
			"PromptTokens":     res.inputTokens,
			"CompletionTokens": res.outputTokens,
			"TotalTokens":      res.inputTokens + res.outputTokens,
		},
	}}}, nil
}

func (l *LLM) post(ctx context.Context, body []byte, tok *auth.Tokens) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	if tok.AccountID != "" {
		req.Header.Set("chatgpt-account-id", tok.AccountID)
	}
	req.Header.Set("originator", auth.Originator)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := l.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ChatGPT request: %w", err)
	}
	return resp, nil
}

type contentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type inputItem struct {
	Type    string        `json:"type"`
	Role    string        `json:"role"`
	Content []contentPart `json:"content"`
}

type textFormat struct {
	Format struct {
		Type string `json:"type"`
	} `json:"format"`
}

type request struct {
	Model  string      `json:"model"`
	Stream bool        `json:"stream"`
	Store  bool        `json:"store"`
	Input  []inputItem `json:"input"`
	Text   *textFormat `json:"text,omitempty"`
}

// buildRequest maps langchaingo messages onto Responses API input items.
// System prompts become developer messages; tool results are fed back as
// user text, matching the text-based agent loop.
func buildRequest(model string, messages []langllms.MessageContent, jsonMode bool) request {
	req := request{Model: model, Stream: true, Store: false, Input: []inputItem{}}
	for _, m := range messages {
		var parts []string
		for _, p := range m.Parts {
			switch v := p.(type) {
			case langllms.TextContent:
				parts = append(parts, v.Text)
			case langllms.ToolCallResponse:
				parts = append(parts, v.Content)
			}
		}
		text := strings.Join(parts, "\n")
		if text == "" {
			continue
		}
		role, partType := "user", "input_text"
		switch m.Role {
		case langllms.ChatMessageTypeSystem:
			role = "developer"
		case langllms.ChatMessageTypeAI:
			role, partType = "assistant", "output_text"
		case langllms.ChatMessageTypeTool:
			text = "Observation: " + text
		}
		req.Input = append(req.Input, inputItem{
			Type:    "message",
			Role:    role,
			Content: []contentPart{{Type: partType, Text: text}},
		})
	}
	if jsonMode {
		req.Text = &textFormat{}
		req.Text.Format.Type = "json_object"
	}
	return req
}
