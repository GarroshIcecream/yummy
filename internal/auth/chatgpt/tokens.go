package chatgpt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Tokens is a ChatGPT OAuth session.
type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IDToken      string    `json:"id_token"`
	AccountID    string    `json:"account_id"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// ErrNotLoggedIn means no ChatGPT session is stored.
var ErrNotLoggedIn = errors.New("not logged in to ChatGPT; run `yummy login chatgpt`")

// newVerifier returns a random RFC 7636 code verifier (86 URL-safe chars).
func newVerifier() (string, error) {
	return randomToken(64)
}

// challengeFor returns the S256 code challenge for a verifier.
func challengeFor(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// accountIDFromJWT reads the ChatGPT account ID from a JWT's payload. The
// signature is not verified: the value only routes requests, and the server
// validates the bearer token itself.
func accountIDFromJWT(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("malformed JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return "", fmt.Errorf("decode JWT payload: %w", err)
	}
	var claims map[string]json.RawMessage
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("decode JWT claims: %w", err)
	}
	var auth struct {
		AccountID string `json:"chatgpt_account_id"`
	}
	if raw, ok := claims[authClaim]; ok {
		_ = json.Unmarshal(raw, &auth)
	}
	if auth.AccountID == "" {
		return "", errors.New("JWT has no ChatGPT account ID")
	}
	return auth.AccountID, nil
}

// Store persists Tokens at <dataDir>/auth/chatgpt.json, readable only by the
// current user.
type Store struct {
	path string
}

// NewStore returns a Store under the given Yummy data directory.
func NewStore(dataDir string) *Store {
	return &Store{path: filepath.Join(dataDir, "auth", "chatgpt.json")}
}

// Load reads the stored session. A missing file yields ErrNotLoggedIn.
func (s *Store) Load() (*Tokens, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotLoggedIn
	}
	if err != nil {
		return nil, fmt.Errorf("read ChatGPT session: %w", err)
	}
	var t Tokens
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parse ChatGPT session: %w", err)
	}
	if t.AccessToken == "" {
		return nil, ErrNotLoggedIn
	}
	return &t, nil
}

// Save writes the session atomically with mode 0600.
func (s *Store) Save(t *Tokens) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create auth dir: %w", err)
	}
	// MkdirAll leaves an existing directory's mode alone; tighten it.
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("secure auth dir: %w", err)
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".chatgpt-*.json")
	if err != nil {
		return fmt.Errorf("write ChatGPT session: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}

// Delete removes the stored session. A missing file is not an error.
func (s *Store) Delete() error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
