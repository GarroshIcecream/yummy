package llms

import (
	"errors"
	"testing"

	"github.com/GarroshIcecream/yummy/internal/config"
)

func TestNewProviderDefaultsToOllama(t *testing.T) {
	p, err := NewProvider(config.NewDefaultChatConfig(), t.TempDir())
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	if p.Name() != "ollama" {
		t.Errorf("Name() = %q, want ollama", p.Name())
	}
	if p.DisplayName() != "Ollama" {
		t.Errorf("DisplayName() = %q, want Ollama", p.DisplayName())
	}
	if p.DefaultModel() != "gemma3:4b" {
		t.Errorf("DefaultModel() = %q, want gemma3:4b", p.DefaultModel())
	}
}

func TestNewProviderEmptyIsOllama(t *testing.T) {
	cfg := config.NewDefaultChatConfig()
	cfg.Provider = ""
	p, err := NewProvider(cfg, t.TempDir())
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	if p.Name() != "ollama" {
		t.Errorf("Name() = %q, want ollama", p.Name())
	}
}

func TestNewProviderUnknown(t *testing.T) {
	cfg := config.NewDefaultChatConfig()
	cfg.Provider = "bogus"
	_, err := NewProvider(cfg, t.TempDir())
	if !errors.Is(err, ErrUnknownProvider) {
		t.Fatalf("err = %v, want ErrUnknownProvider", err)
	}
}
