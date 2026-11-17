package chat

import (
	"context"
	"testing"

	"github.com/GarroshIcecream/yummy/internal/config"
	"github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/llms"
	langllms "github.com/tmc/langchaingo/llms"
)

type stubModel struct{}

func (stubModel) GenerateContent(context.Context, []langllms.MessageContent, ...langllms.CallOption) (*langllms.ContentResponse, error) {
	return &langllms.ContentResponse{Choices: []*langllms.ContentChoice{{Content: "ok"}}}, nil
}

func (m stubModel) Call(ctx context.Context, prompt string, opts ...langllms.CallOption) (string, error) {
	return langllms.GenerateFromSinglePrompt(ctx, m, prompt, opts...)
}

type stubProvider struct{}

func (stubProvider) Name() string                                 { return "stub" }
func (stubProvider) DisplayName() string                          { return "Stub" }
func (stubProvider) DefaultModel() string                         { return "a" }
func (stubProvider) Check(context.Context) error                  { return nil }
func (stubProvider) ListModels(context.Context) ([]string, error) { return []string{"a"}, nil }
func (stubProvider) NewModel(string) (langllms.Model, error)      { return stubModel{}, nil }

func TestLoadSessionUnknownModelKeepsCurrent(t *testing.T) {
	dir := t.TempDir()
	sessionLog, err := db.NewSessionLog(dir, &config.NewDefaultConfig().Database)
	if err != nil {
		t.Fatalf("NewSessionLog: %v", err)
	}
	sessionID, err := sessionLog.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	// Session recorded on a model the active provider doesn't offer.
	if err := sessionLog.SaveSessionMessage(sessionID, "system", langllms.ChatMessageTypeSystem, "gemma3:4b", 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := sessionLog.SaveSessionMessage(sessionID, "hi", langllms.ChatMessageTypeHuman, "gemma3:4b", 0, 0, 0); err != nil {
		t.Fatal(err)
	}

	ai, err := llms.NewLLMServiceWithProvider(nil, stubProvider{})
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewExecutorService(ai, nil, sessionLog)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.LoadSession(sessionID); err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if got := ai.GetModelName(); got != "a" {
		t.Errorf("model = %q, want current model a kept", got)
	}
}
