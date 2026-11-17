package llms

import (
	"context"
	"errors"
	"reflect"
	"testing"

	auth "github.com/GarroshIcecream/yummy/internal/auth/chatgpt"
	"github.com/GarroshIcecream/yummy/internal/config"
)

func chatgptConfig(models ...string) config.ChatConfig {
	cfg := config.NewDefaultChatConfig()
	cfg.Provider = "chatgpt"
	cfg.ChatGPT.Models = models
	return cfg
}

func TestNewProviderChatGPT(t *testing.T) {
	p, err := NewProvider(chatgptConfig(), t.TempDir())
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	if p.Name() != "chatgpt" || p.DisplayName() != "ChatGPT" {
		t.Errorf("Name/DisplayName = %q/%q", p.Name(), p.DisplayName())
	}
}

func TestChatGPTProviderNotLoggedIn(t *testing.T) {
	p, err := NewProvider(chatgptConfig(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Check(context.Background()); !errors.Is(err, auth.ErrNotLoggedIn) {
		t.Fatalf("Check: err = %v, want ErrNotLoggedIn", err)
	}
}

func TestChatGPTProviderListsConfiguredModels(t *testing.T) {
	// Not logged in, so the live catalog can't be fetched: the configured
	// list is the fallback.
	p, err := NewProvider(chatgptConfig("m1", "m2"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	got, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if want := []string{"m1", "m2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ListModels = %v, want %v", got, want)
	}
	if p.DefaultModel() != "m1" {
		t.Errorf("DefaultModel = %q, want m1 (first listed)", p.DefaultModel())
	}
}

func TestChatGPTProviderConfiguredDefaultWins(t *testing.T) {
	cfg := chatgptConfig("m1", "m2")
	cfg.ChatGPT.DefaultModel = "m2"
	p, err := NewProvider(cfg, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if p.DefaultModel() != "m2" {
		t.Errorf("DefaultModel = %q, want m2", p.DefaultModel())
	}
}
