package llms

import (
	"context"
	"testing"

	langllms "github.com/tmc/langchaingo/llms"
)

// fakeModel is a langllms.Model that records its name and returns a canned reply.
type fakeModel struct{ name string }

func (m *fakeModel) GenerateContent(_ context.Context, _ []langllms.MessageContent, _ ...langllms.CallOption) (*langllms.ContentResponse, error) {
	return &langllms.ContentResponse{Choices: []*langllms.ContentChoice{{Content: "ok from " + m.name}}}, nil
}

func (m *fakeModel) Call(ctx context.Context, prompt string, opts ...langllms.CallOption) (string, error) {
	return langllms.GenerateFromSinglePrompt(ctx, m, prompt, opts...)
}

type fakeProvider struct {
	models       []string
	defaultModel string
}

func (p *fakeProvider) Name() string                                 { return "fake" }
func (p *fakeProvider) DisplayName() string                          { return "Fake" }
func (p *fakeProvider) DefaultModel() string                         { return p.defaultModel }
func (p *fakeProvider) Check(context.Context) error                  { return nil }
func (p *fakeProvider) ListModels(context.Context) ([]string, error) { return p.models, nil }
func (p *fakeProvider) NewModel(model string) (langllms.Model, error) {
	return &fakeModel{name: model}, nil
}

func newFakeService(t *testing.T) *LLMService {
	t.Helper()
	s, err := NewLLMServiceWithProvider(nil, &fakeProvider{models: []string{"a", "b"}, defaultModel: "a"})
	if err != nil {
		t.Fatalf("NewLLMServiceWithProvider: %v", err)
	}
	return s
}

func TestSetModelByNameRejectsUnlisted(t *testing.T) {
	s := newFakeService(t)
	if err := s.SetModelByName("c"); err == nil {
		t.Fatal("SetModelByName(c) = nil, want error")
	}
	if s.GetModelName() != "a" {
		t.Errorf("GetModelName() = %q, want a", s.GetModelName())
	}
}

func TestSetModelByNameSwapsModel(t *testing.T) {
	s := newFakeService(t)
	if err := s.SetModelByName("b"); err != nil {
		t.Fatalf("SetModelByName(b): %v", err)
	}
	if s.GetModelName() != "b" {
		t.Errorf("GetModelName() = %q, want b", s.GetModelName())
	}
	fm, ok := s.GetLLM().(*fakeModel)
	if !ok || fm.name != "b" {
		t.Errorf("GetLLM() = %#v, want fakeModel b", s.GetLLM())
	}
	if s.Provider().DisplayName() != "Fake" {
		t.Errorf("Provider().DisplayName() = %q", s.Provider().DisplayName())
	}
}
