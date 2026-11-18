package llms

import (
	"context"
	"fmt"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	langllms "github.com/tmc/langchaingo/llms"
)

// ProviderState is a provider made ready for use (or the reason it isn't).
// It is built off the UI goroutine by PrepareProvider and installed with
// LLMService.ApplyProvider.
type ProviderState struct {
	Provider  Provider
	Models    []string
	ModelName string
	// LLM is never nil: when Err is set it is a model that fails every
	// request with an explanation, so the chat shows why it can't answer.
	LLM langllms.Model
	Err error
}

// PrepareProvider checks the provider, lists its models and builds its
// default model. It may block on the network or on starting Ollama.
func PrepareProvider(ctx context.Context, p Provider) ProviderState {
	st := ProviderState{Provider: p}
	fail := func(err error) ProviderState {
		st.Err = err
		st.Models = nil
		st.ModelName = ""
		st.LLM = unavailableModel{err: fmt.Errorf("%s is not available: %w. Open the command palette%s to sign in or switch provider",
			p.DisplayName(), err, paletteKeyHint())}
		return st
	}
	if err := p.Check(ctx); err != nil {
		return fail(err)
	}
	models, err := p.ListModels(ctx)
	if err != nil {
		return fail(err)
	}
	name := p.DefaultModel()
	if name == "" && len(models) > 0 {
		name = models[0]
	}
	llm, err := p.NewModel(name)
	if err != nil {
		return fail(err)
	}
	st.Models, st.ModelName, st.LLM = models, name, llm
	return st
}

func paletteKeyHint() string {
	cfg := config.GetGlobalConfig()
	if cfg == nil || len(cfg.Keymap.CommandPalette) == 0 {
		return ""
	}
	return " (" + strings.Join(cfg.Keymap.CommandPalette, " / ") + ")"
}

// unavailableModel stands in for a provider that isn't ready.
type unavailableModel struct{ err error }

func (u unavailableModel) GenerateContent(context.Context, []langllms.MessageContent, ...langllms.CallOption) (*langllms.ContentResponse, error) {
	return nil, u.err
}

func (u unavailableModel) Call(context.Context, string, ...langllms.CallOption) (string, error) {
	return "", u.err
}
