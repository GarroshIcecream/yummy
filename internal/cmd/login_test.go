package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/GarroshIcecream/yummy/internal/auth/chatgpt"
)

func withDataDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	prev := dataDirFunc
	dataDirFunc = func() (string, error) { return dir, nil }
	t.Cleanup(func() { dataDirFunc = prev })
	return dir
}

func TestLogoutRemovesTokens(t *testing.T) {
	dir := withDataDir(t)
	store := chatgpt.NewStore(dir)
	if err := store.Save(&chatgpt.Tokens{AccessToken: "a", RefreshToken: "r"}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	logoutCmd.SetOut(&out)
	if err := logoutCmd.RunE(logoutCmd, []string{"chatgpt"}); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := store.Load(); !errors.Is(err, chatgpt.ErrNotLoggedIn) {
		t.Fatalf("Load after logout: err = %v, want ErrNotLoggedIn", err)
	}
}

func TestLoginRejectsUnknownProvider(t *testing.T) {
	withDataDir(t)
	if err := loginCmd.RunE(loginCmd, []string{"foo"}); err == nil {
		t.Fatal("login foo: want error")
	}
	if err := logoutCmd.RunE(logoutCmd, []string{"foo"}); err == nil {
		t.Fatal("logout foo: want error")
	}
}
