package chatgpt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// makeJWT builds an unsigned JWT whose payload is the given claims.
func makeJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	enc := base64.RawURLEncoding
	return enc.EncodeToString([]byte(`{"alg":"none"}`)) + "." + enc.EncodeToString(payload) + ".sig"
}

func idTokenFor(t *testing.T, accountID string) string {
	return makeJWT(t, map[string]any{
		"https://api.openai.com/auth": map[string]any{"chatgpt_account_id": accountID},
	})
}

func TestPKCEChallengeRFC7636(t *testing.T) {
	// RFC 7636 Appendix B test vector.
	got := challengeFor("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk")
	if got != "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" {
		t.Fatalf("challengeFor = %q", got)
	}
	v, err := newVerifier()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[A-Za-z0-9\-._~]{43,128}$`).MatchString(v) {
		t.Fatalf("verifier %q is not a valid RFC 7636 verifier", v)
	}
}

func TestStoreRoundTripAndPerms(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	if _, err := s.Load(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("Load on empty dir: err = %v, want ErrNotLoggedIn", err)
	}
	want := &Tokens{AccessToken: "a", RefreshToken: "r", IDToken: "i", AccountID: "acc", ExpiresAt: time.Now().Add(time.Hour).Round(time.Second)}
	if err := s.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.AccessToken != want.AccessToken || got.RefreshToken != want.RefreshToken || got.AccountID != want.AccountID || !got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Fatalf("Load = %+v, want %+v", got, want)
	}
	fi, err := os.Stat(filepath.Join(dir, "auth", "chatgpt.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("file mode = %v, want 0600", fi.Mode().Perm())
	}
	di, err := os.Stat(filepath.Join(dir, "auth"))
	if err != nil {
		t.Fatal(err)
	}
	if di.Mode().Perm() != 0o700 {
		t.Errorf("dir mode = %v, want 0700", di.Mode().Perm())
	}
	if err := s.Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := s.Delete(); err != nil {
		t.Fatalf("second Delete: %v", err)
	}
	if _, err := s.Load(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("Load after Delete: err = %v", err)
	}
}

func TestAccountIDFromJWT(t *testing.T) {
	id, err := accountIDFromJWT(idTokenFor(t, "acc_123"))
	if err != nil || id != "acc_123" {
		t.Fatalf("accountIDFromJWT = %q, %v; want acc_123", id, err)
	}
	if _, err := accountIDFromJWT(makeJWT(t, map[string]any{"sub": "x"})); err == nil {
		t.Fatal("missing claim: want error")
	}
	if _, err := accountIDFromJWT("not-a-jwt"); err == nil {
		t.Fatal("malformed token: want error")
	}
}

// tokenServer is a stub OAuth token endpoint.
type tokenServer struct {
	t        *testing.T
	srv      *httptest.Server
	mu       sync.Mutex
	forms    []url.Values
	refreshN atomic.Int32
}

func newTokenServer(t *testing.T, accountID string) *tokenServer {
	ts := &tokenServer{t: t}
	ts.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		ts.mu.Lock()
		ts.forms = append(ts.forms, r.PostForm)
		ts.mu.Unlock()
		if r.PostForm.Get("grant_type") == "refresh_token" {
			n := ts.refreshN.Add(1)
			time.Sleep(20 * time.Millisecond) // widen the race window for single-flight tests
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "access-" + strconv.Itoa(int(n)),
				"id_token":     idTokenFor(t, accountID),
				"expires_in":   3600,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access-0",
			"refresh_token": "refresh-0",
			"id_token":      idTokenFor(t, accountID),
			"expires_in":    3600,
		})
	}))
	t.Cleanup(ts.srv.Close)
	return ts
}

func (ts *tokenServer) lastForm() url.Values {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.forms[len(ts.forms)-1]
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

func testClient(ts *tokenServer, ports ...int) *Client {
	c := NewClient()
	c.TokenURL = ts.srv.URL
	c.CallbackPorts = ports
	return c
}

func TestLoginHappyPath(t *testing.T) {
	ts := newTokenServer(t, "acc_123")
	port := freePort(t)
	c := testClient(ts, port)

	var challenge string
	browser := func(authURL string) error {
		u, err := url.Parse(authURL)
		if err != nil {
			return err
		}
		q := u.Query()
		wantRedirect := fmt.Sprintf("http://localhost:%d/auth/callback", port)
		checks := map[string]string{
			"response_type":              "code",
			"client_id":                  ClientID,
			"redirect_uri":               wantRedirect,
			"scope":                      Scope,
			"code_challenge_method":      "S256",
			"id_token_add_organizations": "true",
			"codex_cli_simplified_flow":  "true",
			"originator":                 Originator,
		}
		for k, v := range checks {
			if q.Get(k) != v {
				t.Errorf("authorize param %s = %q, want %q", k, q.Get(k), v)
			}
		}
		challenge = q.Get("code_challenge")
		go func() {
			resp, err := http.Get(wantRedirect + "?code=xyz&state=" + url.QueryEscape(q.Get("state")))
			if err != nil {
				t.Errorf("callback GET: %v", err)
				return
			}
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("callback status = %d", resp.StatusCode)
			}
		}()
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tok, err := c.Login(ctx, browser)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	form := ts.lastForm()
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "xyz" || form.Get("client_id") != ClientID {
		t.Errorf("token form = %v", form)
	}
	if challengeFor(form.Get("code_verifier")) != challenge {
		t.Errorf("code_verifier does not match the code_challenge")
	}
	if tok.AccessToken != "access-0" || tok.RefreshToken != "refresh-0" || tok.AccountID != "acc_123" {
		t.Errorf("tokens = %+v", tok)
	}
	if d := time.Until(tok.ExpiresAt); d < 59*time.Minute || d > 61*time.Minute {
		t.Errorf("ExpiresAt in %v, want ~1h", d)
	}
}

func TestLoginStateMismatch(t *testing.T) {
	ts := newTokenServer(t, "acc")
	port := freePort(t)
	c := testClient(ts, port)
	status := make(chan int, 1)
	browser := func(string) error {
		go func() {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/auth/callback?code=xyz&state=WRONG", port))
			if err != nil {
				status <- 0
				return
			}
			_ = resp.Body.Close()
			status <- resp.StatusCode
		}()
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := c.Login(ctx, browser); err == nil {
		t.Fatal("Login with wrong state: want error")
	}
	if got := <-status; got != http.StatusBadRequest {
		t.Errorf("callback status = %d, want 400", got)
	}
}

func TestLoginPortInUse(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	port := l.Addr().(*net.TCPAddr).Port
	c := testClient(newTokenServer(t, "acc"), port)
	_, err = c.Login(context.Background(), func(string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), strconv.Itoa(port)) {
		t.Fatalf("err = %v, want error naming port %d", err, port)
	}
}

func TestLoginFallsBackToSecondPort(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	busy := l.Addr().(*net.TCPAddr).Port
	free := freePort(t)
	c := testClient(newTokenServer(t, "acc"), busy, free)
	browser := func(authURL string) error {
		u, _ := url.Parse(authURL)
		redirect := u.Query().Get("redirect_uri")
		if !strings.Contains(redirect, strconv.Itoa(free)) {
			t.Errorf("redirect_uri = %q, want port %d", redirect, free)
		}
		go func() {
			resp, err := http.Get(redirect + "?code=c&state=" + url.QueryEscape(u.Query().Get("state")))
			if err == nil {
				_ = resp.Body.Close()
			}
		}()
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := c.Login(ctx, browser); err != nil {
		t.Fatalf("Login: %v", err)
	}
}

func TestLoginContextTimeout(t *testing.T) {
	port := freePort(t)
	c := testClient(newTokenServer(t, "acc"), port)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.Login(ctx, func(string) error { return nil })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want DeadlineExceeded", err)
	}
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("port %d still bound after Login returned: %v", port, err)
	}
	_ = l.Close()
}

func TestTokenSourceRefreshesNearExpiry(t *testing.T) {
	ts := newTokenServer(t, "acc_9")
	store := NewStore(t.TempDir())
	if err := store.Save(&Tokens{AccessToken: "old", RefreshToken: "refresh-old", AccountID: "acc_9", ExpiresAt: time.Now().Add(30 * time.Second)}); err != nil {
		t.Fatal(err)
	}
	src := NewTokenSource(store, testClient(ts))
	tok, err := src.Token(context.Background())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if tok.AccessToken != "access-1" {
		t.Errorf("AccessToken = %q, want refreshed access-1", tok.AccessToken)
	}
	if tok.RefreshToken != "refresh-old" {
		t.Errorf("RefreshToken = %q, want old one kept when response omits it", tok.RefreshToken)
	}
	if f := ts.lastForm(); f.Get("grant_type") != "refresh_token" || f.Get("refresh_token") != "refresh-old" {
		t.Errorf("refresh form = %v", f)
	}
	saved, err := store.Load()
	if err != nil || saved.AccessToken != "access-1" {
		t.Errorf("persisted = %+v, %v", saved, err)
	}
	// A fresh token is served without another refresh.
	if _, err := src.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	if n := ts.refreshN.Load(); n != 1 {
		t.Errorf("refreshes = %d, want 1", n)
	}
}

func TestForceRefreshSingleFlight(t *testing.T) {
	ts := newTokenServer(t, "acc")
	store := NewStore(t.TempDir())
	if err := store.Save(&Tokens{AccessToken: "old", RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	src := NewTokenSource(store, testClient(ts))
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := src.ForceRefresh(context.Background()); err != nil {
				t.Errorf("ForceRefresh: %v", err)
			}
		}()
	}
	wg.Wait()
	if n := ts.refreshN.Load(); n != 1 {
		t.Errorf("refreshes = %d, want 1", n)
	}
}

func TestTokenSourceNotLoggedIn(t *testing.T) {
	src := NewTokenSource(NewStore(t.TempDir()), NewClient())
	if _, err := src.Token(context.Background()); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("err = %v, want ErrNotLoggedIn", err)
	}
}
