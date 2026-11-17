package chatgpt

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// refreshSkew refreshes tokens this long before they expire.
const refreshSkew = 60 * time.Second

// forceRefreshDedupe treats a refresh this recent as satisfying a
// ForceRefresh, so concurrent 401s share one refresh.
const forceRefreshDedupe = 5 * time.Second

// TokenSource hands out a valid session, refreshing and persisting it as
// needed. It is safe for concurrent use.
type TokenSource struct {
	store  *Store
	client *Client

	mu          sync.Mutex
	tokens      *Tokens
	lastRefresh time.Time
}

// NewTokenSource returns a TokenSource backed by store.
func NewTokenSource(store *Store, client *Client) *TokenSource {
	return &TokenSource{store: store, client: client}
}

// Token returns the current session, refreshing it first if it expires
// within a minute. Returns ErrNotLoggedIn when there is no session.
func (ts *TokenSource) Token(ctx context.Context) (*Tokens, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if err := ts.loadLocked(); err != nil {
		return nil, err
	}
	if !ts.tokens.ExpiresAt.IsZero() && time.Until(ts.tokens.ExpiresAt) < refreshSkew {
		if err := ts.refreshLocked(ctx); err != nil {
			return nil, err
		}
	}
	cp := *ts.tokens
	return &cp, nil
}

// ForceRefresh refreshes the session regardless of its expiry, e.g. after
// the server rejected the access token. Concurrent callers share a single
// refresh.
func (ts *TokenSource) ForceRefresh(ctx context.Context) (*Tokens, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if err := ts.loadLocked(); err != nil {
		return nil, err
	}
	if time.Since(ts.lastRefresh) > forceRefreshDedupe {
		if err := ts.refreshLocked(ctx); err != nil {
			return nil, err
		}
	}
	cp := *ts.tokens
	return &cp, nil
}

func (ts *TokenSource) loadLocked() error {
	if ts.tokens != nil {
		return nil
	}
	t, err := ts.store.Load()
	if err != nil {
		return err
	}
	ts.tokens = t
	return nil
}

func (ts *TokenSource) refreshLocked(ctx context.Context) error {
	// Another TokenSource (a second Yummy process, or `yummy login`) may have
	// rotated or replaced the session since we loaded it. Refreshing with
	// our stale refresh token would be rejected, so adopt the newer one.
	if disk, err := ts.store.Load(); err == nil && disk.RefreshToken != ts.tokens.RefreshToken {
		ts.tokens = disk
		if disk.ExpiresAt.IsZero() || time.Until(disk.ExpiresAt) >= refreshSkew {
			ts.lastRefresh = time.Now()
			return nil
		}
	}
	if ts.tokens.RefreshToken == "" {
		return ErrNotLoggedIn
	}
	t, err := ts.client.Refresh(ctx, ts.tokens.RefreshToken)
	if err != nil {
		return fmt.Errorf("refresh ChatGPT session (run `yummy login chatgpt` if this persists): %w", err)
	}
	if t.AccountID == "" {
		t.AccountID = ts.tokens.AccountID
	}
	if t.IDToken == "" {
		t.IDToken = ts.tokens.IDToken
	}
	if err := ts.store.Save(t); err != nil {
		return fmt.Errorf("save refreshed ChatGPT session: %w", err)
	}
	ts.tokens = t
	ts.lastRefresh = time.Now()
	return nil
}
