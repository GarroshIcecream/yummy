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

	mu     sync.Mutex
	models []string // cached after the first successful listing
}

func newChatGPTProvider(cfg config.ChatGPTConfig, dataDir string) *chatgptProvider {
	return &chatgptProvider{
		cfg: cfg,
		ts:  auth.NewTokenSource(auth.NewStore(dataDir), auth.NewClient()),
	}
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
	if p.DefaultModel() == "" {
		return errors.New("no ChatGPT models available; set chat.chatgpt.default_model in config.json")
	}
	return nil
}

// ListModels returns the account's model catalog, falling back to the
// configured list when it can't be fetched.
func (p *chatgptProvider) ListModels(ctx context.Context) ([]string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.models != nil {
		return p.models, nil
	}
	models, err := chatgpt.FetchModels(ctx, p.ts)
	if err == nil && len(models) > 0 {
		p.models = models
		return models, nil
	}
	if len(p.cfg.Models) > 0 {
		if err != nil {
			slog.Warn("Could not fetch ChatGPT models, using configured list", "error", err)
		}
		return p.cfg.Models, nil
	}
	if err == nil {
		err = errors.New("ChatGPT account offers no models")
	}
	return nil, err
}

func (p *chatgptProvider) NewModel(model string) (langllms.Model, error) {
	return chatgpt.New(model, p.ts), nil
}
