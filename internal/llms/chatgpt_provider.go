package llms

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	auth "github.com/GarroshIcecream/yummy/internal/auth/chatgpt"
	"github.com/GarroshIcecream/yummy/internal/config"
	"github.com/GarroshIcecream/yummy/internal/llms/chatgpt"
	langllms "github.com/tmc/langchaingo/llms"
)

// chatgptProvider serves models through the user's ChatGPT subscription.
type chatgptProvider struct {
	cfg config.ChatGPTConfig
	ts  *auth.TokenSource

	// modelOpts are passed to chatgpt.FetchModels (tests point it at a stub).
	modelOpts []chatgpt.Option

	mu     sync.Mutex
	models []string // cached after the first listing
}

// tokenSources holds one TokenSource per data directory so every provider
// in the process shares a session and refresh-token rotations never race.
var tokenSources sync.Map // dataDir -> *auth.TokenSource

func sharedTokenSource(dataDir string) *auth.TokenSource {
	if ts, ok := tokenSources.Load(dataDir); ok {
		return ts.(*auth.TokenSource)
	}
	ts, _ := tokenSources.LoadOrStore(dataDir, auth.NewTokenSource(auth.NewStore(dataDir), auth.NewClient()))
	return ts.(*auth.TokenSource)
}

func newChatGPTProvider(cfg config.ChatGPTConfig, dataDir string) *chatgptProvider {
	return &chatgptProvider{cfg: cfg, ts: sharedTokenSource(dataDir)}
}

func (p *chatgptProvider) Name() string        { return "chatgpt" }
func (p *chatgptProvider) DisplayName() string { return "ChatGPT" }

// DefaultModel is the configured model, else the first model on the account.
func (p *chatgptProvider) DefaultModel() string {
	if p.cfg.DefaultModel != "" {
		return p.cfg.DefaultModel
	}
	models, err := p.ListModels(context.Background())
	if err != nil || len(models) == 0 {
		return ""
	}
	return models[0]
}

// Check verifies there is a usable ChatGPT session and a model to use.
func (p *chatgptProvider) Check(ctx context.Context) error {
	if _, err := p.ts.Token(ctx); err != nil {
		return err
	}
	if _, err := p.ListModels(ctx); err != nil {
		return err
	}
	if p.DefaultModel() == "" {
		return errors.New("no ChatGPT models available; set chat.chatgpt.default_model in config.json")
	}
	return nil
}

// ListModels returns the account's model catalog. If it can't be fetched,
// the configured default model and model list are used instead.
func (p *chatgptProvider) ListModels(ctx context.Context) ([]string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.models != nil {
		return p.models, nil
	}
	models, err := chatgpt.FetchModels(ctx, p.ts, p.modelOpts...)
	if err == nil && len(models) > 0 {
		p.models = models
		return models, nil
	}
	if err == nil {
		err = errors.New("ChatGPT account offers no models")
	}
	var fallback []string
	if p.cfg.DefaultModel != "" {
		fallback = append(fallback, p.cfg.DefaultModel)
	}
	for _, m := range p.cfg.Models {
		if m != p.cfg.DefaultModel {
			fallback = append(fallback, m)
		}
	}
	if len(fallback) == 0 {
		return nil, err
	}
	slog.Warn("Could not fetch ChatGPT models, using configured ones", "error", err)
	p.models = fallback
	return fallback, nil
}

func (p *chatgptProvider) NewModel(model string) (langllms.Model, error) {
	return chatgpt.New(model, p.ts), nil
}
