package chatgpt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client runs the OAuth flows. The zero value is not usable; use NewClient.
type Client struct {
	HTTP          *http.Client
	AuthorizeURL  string
	TokenURL      string
	CallbackPorts []int
}

// NewClient returns a Client pointed at OpenAI's auth server.
func NewClient() *Client {
	return &Client{
		HTTP:          &http.Client{Timeout: 30 * time.Second},
		AuthorizeURL:  AuthorizeURL,
		TokenURL:      TokenURL,
		CallbackPorts: append([]int(nil), DefaultCallbackPorts...),
	}
}

type callbackResult struct {
	code string
	err  error
}

const callbackPage = `<!doctype html><meta charset="utf-8"><title>Yummy</title>
<body style="font-family:system-ui;text-align:center;padding-top:4em">
<h2>%s</h2><p>You can close this tab and return to the terminal.</p></body>`

// Login runs the browser sign-in: it starts a loopback callback server,
// hands the authorization URL to openBrowser, waits for the redirect and
// exchanges the code for tokens. It returns when ctx is done.
func (c *Client) Login(ctx context.Context, openBrowser func(url string) error) (*Tokens, error) {
	ln, port, err := c.listen()
	if err != nil {
		return nil, err
	}

	verifier, err := newVerifier()
	if err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("generate PKCE verifier: %w", err)
	}
	state, err := randomToken(32)
	if err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("generate state: %w", err)
	}
	redirectURI := fmt.Sprintf("http://localhost:%d%s", port, callbackPath)

	results := make(chan callbackResult, 1)
	deliver := func(r callbackResult) {
		select {
		case results <- r:
		default:
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		switch {
		case q.Get("state") != state:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, callbackPage, "Sign-in failed: state mismatch")
			deliver(callbackResult{err: errors.New("OAuth state mismatch; please try again")})
		case q.Get("error") != "":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, callbackPage, "Sign-in was not completed")
			deliver(callbackResult{err: fmt.Errorf("authorization failed: %s %s", q.Get("error"), q.Get("error_description"))})
		case q.Get("code") == "":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, callbackPage, "Sign-in failed: missing code")
			deliver(callbackResult{err: errors.New("authorization callback had no code")})
		default:
			_, _ = fmt.Fprintf(w, callbackPage, "Signed in to ChatGPT")
			deliver(callbackResult{code: q.Get("code")})
		}
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = srv.Serve(ln) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	authURL := c.AuthorizeURL + "?" + url.Values{
		"response_type":              {"code"},
		"client_id":                  {ClientID},
		"redirect_uri":               {redirectURI},
		"scope":                      {Scope},
		"code_challenge":             {challengeFor(verifier)},
		"code_challenge_method":      {"S256"},
		"id_token_add_organizations": {"true"},
		"codex_cli_simplified_flow":  {"true"},
		"state":                      {state},
		"originator":                 {Originator},
	}.Encode()
	if err := openBrowser(authURL); err != nil {
		slog.Warn("Could not open browser for ChatGPT sign-in", "error", err)
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("waiting for ChatGPT sign-in: %w", ctx.Err())
	case r := <-results:
		if r.err != nil {
			return nil, r.err
		}
		return c.requestTokens(ctx, url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {r.code},
			"redirect_uri":  {redirectURI},
			"client_id":     {ClientID},
			"code_verifier": {verifier},
		}, "")
	}
}

// listen binds the first free callback port.
func (c *Client) listen() (net.Listener, int, error) {
	var errs []string
	for _, port := range c.CallbackPorts {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return ln, port, nil
		}
		errs = append(errs, fmt.Sprintf("port %d: %v", port, err))
	}
	return nil, 0, fmt.Errorf("cannot start sign-in callback server (is another login, e.g. Codex, running?): %s", strings.Join(errs, "; "))
}

// Refresh trades a refresh token for a new session. If the response omits a
// refresh token the old one is kept.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	return c.requestTokens(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {ClientID},
	}, refreshToken)
}

func (c *Client) requestTokens(ctx context.Context, form url.Values, prevRefresh string) (*Tokens, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s request: %w", form.Get("grant_type"), err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s failed: HTTP %d: %s", form.Get("grant_type"), resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var tr struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if tr.AccessToken == "" {
		return nil, errors.New("token response had no access token")
	}
	t := &Tokens{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		IDToken:      tr.IDToken,
	}
	if t.RefreshToken == "" {
		t.RefreshToken = prevRefresh
	}
	if tr.ExpiresIn > 0 {
		t.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}
	if id, err := accountIDFromJWT(tr.IDToken); err == nil {
		t.AccountID = id
	} else if id, err := accountIDFromJWT(tr.AccessToken); err == nil {
		t.AccountID = id
	}
	return t, nil
}
