package chatgpt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	auth "github.com/GarroshIcecream/yummy/internal/auth/chatgpt"
	langllms "github.com/tmc/langchaingo/llms"
)

type fakeTokens struct {
	mu        sync.Mutex
	access    string
	refreshes int
	refreshed string // access token handed out after a refresh
}

func (f *fakeTokens) Token(context.Context) (*auth.Tokens, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return &auth.Tokens{AccessToken: f.access, AccountID: "acc_1"}, nil
}

func (f *fakeTokens) ForceRefresh(context.Context) (*auth.Tokens, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refreshes++
	f.access = f.refreshed
	return &auth.Tokens{AccessToken: f.access, AccountID: "acc_1"}, nil
}

// sseEvents renders Responses API stream events.
func sseEvents(events ...map[string]any) string {
	var b strings.Builder
	for _, e := range events {
		data, _ := json.Marshal(e)
		fmt.Fprintf(&b, "event: %s\ndata: %s\n\n", e["type"], data)
	}
	return b.String()
}

func delta(s string) map[string]any {
	return map[string]any{"type": "response.output_text.delta", "delta": s}
}

func completed(in, out int) map[string]any {
	return map[string]any{"type": "response.completed", "response": map[string]any{
		"usage": map[string]any{"input_tokens": in, "output_tokens": out, "total_tokens": in + out},
	}}
}

type captured struct {
	header http.Header
	body   map[string]any
}

// newServer serves the given SSE body and records each request.
func newServer(t *testing.T, sse string) (*httptest.Server, *[]captured) {
	t.Helper()
	var reqs []captured
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(data, &body)
		mu.Lock()
		reqs = append(reqs, captured{header: r.Header.Clone(), body: body})
		mu.Unlock()
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, sse)
	}))
	t.Cleanup(srv.Close)
	return srv, &reqs
}

func newTestLLM(srv *httptest.Server, tokens TokenProvider) *LLM {
	return New("gpt-test", tokens, WithEndpoint(srv.URL), WithHTTPClient(srv.Client()))
}

func TestRequestShape(t *testing.T) {
	srv, reqs := newServer(t, sseEvents(delta("hi"), completed(1, 1)))
	llm := newTestLLM(srv, &fakeTokens{access: "tok"})
	msgs := []langllms.MessageContent{
		langllms.TextParts(langllms.ChatMessageTypeSystem, "be a chef"),
		langllms.TextParts(langllms.ChatMessageTypeHuman, "hello"),
		langllms.TextParts(langllms.ChatMessageTypeAI, "hi there"),
		langllms.TextParts(langllms.ChatMessageTypeHuman, "soup?"),
	}
	_, err := llm.GenerateContent(context.Background(), msgs,
		langllms.WithTemperature(0.5), langllms.WithMaxTokens(100), langllms.WithStopWords([]string{"X"}))
	if err != nil {
		t.Fatalf("GenerateContent: %v", err)
	}
	if len(*reqs) != 1 {
		t.Fatalf("requests = %d, want 1", len(*reqs))
	}
	req := (*reqs)[0]
	for k, v := range map[string]string{
		"Authorization":      "Bearer tok",
		"Chatgpt-Account-Id": "acc_1",
		"Originator":         auth.Originator,
		"Accept":             "text/event-stream",
		"Content-Type":       "application/json",
	} {
		if got := req.header.Get(k); got != v {
			t.Errorf("header %s = %q, want %q", k, got, v)
		}
	}
	b := req.body
	if b["model"] != "gpt-test" || b["stream"] != true || b["store"] != false {
		t.Errorf("model/stream/store = %v/%v/%v", b["model"], b["stream"], b["store"])
	}
	for _, k := range []string{"temperature", "max_output_tokens", "stop", "instructions"} {
		if _, ok := b[k]; ok {
			t.Errorf("body has forbidden field %q", k)
		}
	}
	input, _ := b["input"].([]any)
	wantRoles := []string{"developer", "user", "assistant", "user"}
	wantTypes := []string{"input_text", "input_text", "output_text", "input_text"}
	wantText := []string{"be a chef", "hello", "hi there", "soup?"}
	if len(input) != len(wantRoles) {
		t.Fatalf("input = %v", input)
	}
	for i, it := range input {
		m := it.(map[string]any)
		content := m["content"].([]any)[0].(map[string]any)
		if m["type"] != "message" || m["role"] != wantRoles[i] || content["type"] != wantTypes[i] || content["text"] != wantText[i] {
			t.Errorf("input[%d] = %v", i, m)
		}
	}
}

func TestStreamingDeltas(t *testing.T) {
	srv, _ := newServer(t, sseEvents(delta("Hel"), delta("lo"), completed(3, 2)))
	llm := newTestLLM(srv, &fakeTokens{access: "tok"})
	var streamed strings.Builder
	resp, err := llm.GenerateContent(context.Background(),
		[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")},
		langllms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			streamed.Write(chunk)
			return nil
		}))
	if err != nil {
		t.Fatalf("GenerateContent: %v", err)
	}
	if resp.Choices[0].Content != "Hello" {
		t.Errorf("Content = %q, want Hello", resp.Choices[0].Content)
	}
	if streamed.String() != "Hello" {
		t.Errorf("streamed = %q, want Hello", streamed.String())
	}
}

func TestStopWordsTruncateAndStopStreaming(t *testing.T) {
	srv, _ := newServer(t, sseEvents(
		delta("Thought: x\nAction: list\nAction Input: {}\nObs"),
		delta("ervation: fake"),
		delta(" more invented text"),
		completed(1, 1),
	))
	llm := newTestLLM(srv, &fakeTokens{access: "tok"})
	var streamed strings.Builder
	resp, err := llm.GenerateContent(context.Background(),
		[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")},
		langllms.WithStopWords([]string{"\nObservation:", "\n\tObservation:"}),
		langllms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			streamed.Write(chunk)
			return nil
		}))
	if err != nil {
		t.Fatalf("GenerateContent: %v", err)
	}
	want := "Thought: x\nAction: list\nAction Input: {}"
	if resp.Choices[0].Content != want {
		t.Errorf("Content = %q, want %q", resp.Choices[0].Content, want)
	}
	if streamed.String() != want {
		t.Errorf("streamed = %q, want %q", streamed.String(), want)
	}
}

func TestStreamingKeepsMultibyteRunesWhole(t *testing.T) {
	srv, _ := newServer(t, sseEvents(delta("crème brûlée"), completed(1, 1)))
	llm := newTestLLM(srv, &fakeTokens{access: "tok"})
	var chunks []string
	_, err := llm.GenerateContent(context.Background(),
		[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")},
		langllms.WithStopWords([]string{"\nObservation:"}),
		langllms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			chunks = append(chunks, string(chunk))
			return nil
		}))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range chunks {
		if !utf8.ValidString(c) {
			t.Errorf("chunk %q splits a rune", c)
		}
	}
	if strings.Join(chunks, "") != "crème brûlée" {
		t.Errorf("chunks = %q", chunks)
	}
}

func TestUsageMapping(t *testing.T) {
	srv, _ := newServer(t, sseEvents(delta("ok"), completed(12, 5)))
	llm := newTestLLM(srv, &fakeTokens{access: "tok"})
	resp, err := llm.GenerateContent(context.Background(),
		[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")})
	if err != nil {
		t.Fatal(err)
	}
	info := resp.Choices[0].GenerationInfo
	if info["prompt_eval_count"] != 12 || info["eval_count"] != 5 {
		t.Errorf("GenerationInfo = %v, want prompt_eval_count=12 eval_count=5", info)
	}
}

func TestUnauthorizedRefreshesOnceThenFails(t *testing.T) {
	t.Run("recovers", func(t *testing.T) {
		calls := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if r.Header.Get("Authorization") != "Bearer fresh" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = io.WriteString(w, sseEvents(delta("ok"), completed(1, 1)))
		}))
		defer srv.Close()
		tokens := &fakeTokens{access: "stale", refreshed: "fresh"}
		resp, err := newTestLLM(srv, tokens).GenerateContent(context.Background(),
			[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")})
		if err != nil {
			t.Fatalf("GenerateContent: %v", err)
		}
		if resp.Choices[0].Content != "ok" || tokens.refreshes != 1 || calls != 2 {
			t.Errorf("content=%q refreshes=%d calls=%d", resp.Choices[0].Content, tokens.refreshes, calls)
		}
	})
	t.Run("gives up", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()
		tokens := &fakeTokens{access: "stale", refreshed: "still-bad"}
		_, err := newTestLLM(srv, tokens).GenerateContent(context.Background(),
			[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")})
		if err == nil || !strings.Contains(err.Error(), "yummy login chatgpt") {
			t.Fatalf("err = %v, want hint to run yummy login chatgpt", err)
		}
		if tokens.refreshes != 1 {
			t.Errorf("refreshes = %d, want 1", tokens.refreshes)
		}
	})
}

func TestResponseFailedEvent(t *testing.T) {
	srv, _ := newServer(t, sseEvents(map[string]any{
		"type":     "response.failed",
		"response": map[string]any{"error": map[string]any{"message": "model not supported"}},
	}))
	_, err := newTestLLM(srv, &fakeTokens{access: "tok"}).GenerateContent(context.Background(),
		[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")})
	if err == nil || !strings.Contains(err.Error(), "model not supported") {
		t.Fatalf("err = %v", err)
	}
}

func TestHTTPErrorIncludesBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"detail":"Unsupported parameter: foo"}`)
	}))
	defer srv.Close()
	_, err := newTestLLM(srv, &fakeTokens{access: "tok"}).GenerateContent(context.Background(),
		[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")})
	if err == nil || !strings.Contains(err.Error(), "Unsupported parameter: foo") {
		t.Fatalf("err = %v", err)
	}
}

func TestJSONMode(t *testing.T) {
	srv, reqs := newServer(t, sseEvents(delta("{}"), completed(1, 1)))
	_, err := newTestLLM(srv, &fakeTokens{access: "tok"}).GenerateContent(context.Background(),
		[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")},
		langllms.WithJSONMode())
	if err != nil {
		t.Fatal(err)
	}
	text, _ := (*reqs)[0].body["text"].(map[string]any)
	format, _ := text["format"].(map[string]any)
	if format["type"] != "json_object" {
		t.Errorf("text = %v, want format json_object", (*reqs)[0].body["text"])
	}
}

func TestTokenErrorPropagates(t *testing.T) {
	srv, _ := newServer(t, "")
	_, err := New("m", errTokens{}, WithEndpoint(srv.URL)).GenerateContent(context.Background(),
		[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")})
	if !errors.Is(err, auth.ErrNotLoggedIn) {
		t.Fatalf("err = %v, want ErrNotLoggedIn", err)
	}
}

type errTokens struct{}

func (errTokens) Token(context.Context) (*auth.Tokens, error)        { return nil, auth.ErrNotLoggedIn }
func (errTokens) ForceRefresh(context.Context) (*auth.Tokens, error) { return nil, auth.ErrNotLoggedIn }

func TestStreamWithoutTerminalEventIsError(t *testing.T) {
	srv, _ := newServer(t, sseEvents(delta("partial answ")))
	_, err := newTestLLM(srv, &fakeTokens{access: "tok"}).GenerateContent(context.Background(),
		[]langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "x")})
	if err == nil || !strings.Contains(err.Error(), "ended before") {
		t.Fatalf("err = %v, want stream-ended-early error", err)
	}
}
