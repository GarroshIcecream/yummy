package llms

import (
	"context"
	"errors"
	"fmt"

	"github.com/GarroshIcecream/yummy/internal/config"
	langllms "github.com/tmc/langchaingo/llms"
)

// Provider is an LLM backend (local Ollama, ChatGPT subscription, ...). It
// owns availability checks, model discovery and model construction; the rest
// of the app only deals in langchaingo models.
type Provider interface {
	// Name is the config identifier, e.g. "ollama" or "chatgpt".
	Name() string
	// DisplayName is the human-readable name shown in the UI.
	DisplayName() string
	// DefaultModel is the model used when none is chosen explicitly.
	DefaultModel() string
	// Check reports whether the provider is ready to serve DefaultModel.
	Check(ctx context.Context) error
	// ListModels returns the models the user can switch between.
	ListModels(ctx context.Context) ([]string, error)
	// NewModel constructs a langchaingo model for the given model name.
	NewModel(model string) (langllms.Model, error)
}

// ErrUnknownProvider is returned when chat.provider names no known backend.
var ErrUnknownProvider = errors.New("unknown llm provider")

// NewProvider returns the provider selected by cfg.Provider. An empty value
// means Ollama, for configs written before the setting existed. dataDir is
// the Yummy data directory (used by providers that store credentials).
func NewProvider(cfg config.ChatConfig, dataDir string) (Provider, error) {
	switch cfg.Provider {
	case "", "ollama":
		return &ollamaProvider{defaultModel: cfg.DefaultModel}, nil
	default:
		return nil, fmt.Errorf("%w %q (valid: ollama, chatgpt)", ErrUnknownProvider, cfg.Provider)
	}
}

// NewIngredientModel builds a model on the active provider for ingredient
// parsing. An empty name uses the provider's default model.
func NewIngredientModel(name string) (langllms.Model, error) {
	p, err := NewProvider(config.GetChatConfig(), config.DataDir())
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = p.DefaultModel()
	}
	return p.NewModel(name)
}
