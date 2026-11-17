package llms

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	auth "github.com/GarroshIcecream/yummy/internal/auth/chatgpt"
	"github.com/GarroshIcecream/yummy/internal/config"
	"github.com/GarroshIcecream/yummy/internal/llms/chatgpt"
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

// loggedInProvider returns a chatgpt provider with a stored session whose
// model catalog is served by handler.
func loggedInProvider(t *testing.T, cfg config.ChatConfig, handler http.HandlerFunc) Provider {
	t.Helper()
	dir := t.TempDir()
	if err := auth.NewStore(dir).Save(&auth.Tokens{AccessToken: "a", RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	p, err := NewProvider(cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	p.(*chatgptProvider).modelOpts = []chatgpt.Option{chatgpt.WithModelsURL(srv.URL)}
	return p
}

func TestChatGPTProviderCatalogDownUsesConfiguredDefault(t *testing.T) {
	cfg := chatgptConfig()
	cfg.ChatGPT.DefaultModel = "gpt-x"
	p := loggedInProvider(t, cfg, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	if err := p.Check(context.Background()); err != nil {
		t.Fatalf("Check: %v", err)
	}
	got, err := p.ListModels(context.Background())
	if err != nil || !reflect.DeepEqual(got, []string{"gpt-x"}) {
		t.Fatalf("ListModels = %v, %v; want [gpt-x]", got, err)
	}
}

func TestChatGPTProviderCheckSurfacesCatalogError(t *testing.T) {
	p := loggedInProvider(t, chatgptConfig(), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	err := p.Check(context.Background())
	if err == nil || !strings.Contains(err.Error(), "418") {
		t.Fatalf("Check err = %v, want the catalog HTTP 418 error", err)
	}
}

func TestChatGPTProvidersShareTokenSource(t *testing.T) {
	dir := t.TempDir()
	a, _ := NewProvider(chatgptConfig(), dir)
	b, _ := NewProvider(chatgptConfig(), dir)
	if a.(*chatgptProvider).ts != b.(*chatgptProvider).ts {
		t.Error("providers for the same data dir should share one TokenSource")
	}
}
