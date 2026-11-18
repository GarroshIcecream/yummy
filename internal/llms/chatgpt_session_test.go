package llms

import (
	"context"
	"errors"
	"testing"
	"time"

	auth "github.com/GarroshIcecream/yummy/internal/auth/chatgpt"
)

func TestChatGPTSessionLifecycle(t *testing.T) {
	dir := t.TempDir()
	if ChatGPTSignedIn(dir) {
		t.Fatal("signed in before login")
	}
	// A provider created before sign-in shares the token source.
	p, _ := NewProvider(chatgptConfig(), dir)
	if err := SaveChatGPTSession(dir, &auth.Tokens{AccessToken: "a", RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatalf("SaveChatGPTSession: %v", err)
	}
	if !ChatGPTSignedIn(dir) {
		t.Fatal("not signed in after save")
	}
	if _, err := p.(*chatgptProvider).ts.Token(context.Background()); err != nil {
		t.Fatalf("provider token after sign-in: %v", err)
	}
	if err := ChatGPTLogout(dir); err != nil {
		t.Fatalf("ChatGPTLogout: %v", err)
	}
	if ChatGPTSignedIn(dir) {
		t.Fatal("still signed in after logout")
	}
	if _, err := p.(*chatgptProvider).ts.Token(context.Background()); !errors.Is(err, auth.ErrNotLoggedIn) {
		t.Fatalf("provider token after logout: err = %v, want ErrNotLoggedIn", err)
	}
}
