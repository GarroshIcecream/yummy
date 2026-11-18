package llms

import (
	"context"
	"errors"
	"strings"
	"testing"

	langllms "github.com/tmc/langchaingo/llms"
)

type brokenProvider struct{ *fakeProvider }

func (brokenProvider) Name() string                { return "broken" }
func (brokenProvider) DisplayName() string         { return "Broken" }
func (brokenProvider) Check(context.Context) error { return errors.New("not signed in") }

func TestPrepareProviderUnavailable(t *testing.T) {
	st := PrepareProvider(context.Background(), brokenProvider{&fakeProvider{}})
	if st.Err == nil {
		t.Fatal("Err = nil, want the Check error")
	}
	if st.LLM == nil {
		t.Fatal("LLM = nil, want a model that reports the error")
	}
	_, err := st.LLM.GenerateContent(context.Background(), []langllms.MessageContent{langllms.TextParts(langllms.ChatMessageTypeHuman, "hi")})
	if err == nil || !strings.Contains(err.Error(), "not signed in") || !strings.Contains(err.Error(), "command palette") {
		t.Fatalf("GenerateContent err = %v, want the cause plus a command-palette hint", err)
	}
}

func TestNewLLMServiceToleratesUnreadyProvider(t *testing.T) {
	s, err := NewLLMServiceWithProvider(nil, brokenProvider{&fakeProvider{}})
	if err != nil {
		t.Fatalf("NewLLMServiceWithProvider: %v", err)
	}
	if s.ProviderError() == nil {
		t.Error("ProviderError() = nil, want the readiness error")
	}
	if len(s.GetInstalledModels()) != 0 {
		t.Errorf("models = %v, want none", s.GetInstalledModels())
	}
}

func TestApplyProviderSwitches(t *testing.T) {
	s, err := NewLLMServiceWithProvider(nil, brokenProvider{&fakeProvider{}})
	if err != nil {
		t.Fatal(err)
	}
	st := PrepareProvider(context.Background(), &fakeProvider{models: []string{"x", "y"}, defaultModel: "y"})
	if st.Err != nil {
		t.Fatalf("PrepareProvider: %v", st.Err)
	}
	s.ApplyProvider(st)
	if s.ProviderError() != nil || s.GetModelName() != "y" || s.Provider().DisplayName() != "Fake" {
		t.Errorf("after apply: err=%v model=%q provider=%q", s.ProviderError(), s.GetModelName(), s.Provider().DisplayName())
	}
	if got := s.GetInstalledModels(); len(got) != 2 {
		t.Errorf("models = %v", got)
	}
}
