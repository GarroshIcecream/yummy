package chatgpt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	auth "github.com/GarroshIcecream/yummy/internal/auth/chatgpt"
)

// DefaultModelsURL lists the models a ChatGPT plan grants.
const DefaultModelsURL = auth.CodexBaseURL + "/models"

// catalogClientVersion is sent as client_version, which the backend
// requires; it hides models from clients it considers too old. Yummy runs
// its own agent loop, so a high version asks for the full catalog.
const catalogClientVersion = "999.0.0"

// WithModelsURL overrides the model catalog URL used by FetchModels.
func WithModelsURL(u string) Option { return func(l *LLM) { l.modelsURL = u } }

// FetchModels returns the IDs of the user-selectable models on the account.
// Accepts WithModelsURL and WithHTTPClient.
func FetchModels(ctx context.Context, tokens TokenProvider, opts ...Option) ([]string, error) {
	cfg := &LLM{modelsURL: DefaultModelsURL, http: &http.Client{Timeout: 30 * time.Second}}
	for _, o := range opts {
		o(cfg)
	}
	tok, err := tokens.Token(ctx)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(cfg.modelsURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("client_version", catalogClientVersion)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	if tok.AccountID != "" {
		req.Header.Set("chatgpt-account-id", tok.AccountID)
	}
	req.Header.Set("originator", auth.Originator)
	resp, err := cfg.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch ChatGPT models: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read ChatGPT models: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch ChatGPT models: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var catalog struct {
		Models []struct {
			Slug       string `json:"slug"`
			Visibility string `json:"visibility"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &catalog); err != nil {
		return nil, fmt.Errorf("decode ChatGPT models: %w", err)
	}
	var ids []string
	for _, m := range catalog.Models {
		// Anything not "list" is internal or retired.
		if m.Slug != "" && m.Visibility == "list" {
			ids = append(ids, m.Slug)
		}
	}
	return ids, nil
}
