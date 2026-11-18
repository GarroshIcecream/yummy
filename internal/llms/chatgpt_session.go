package llms

import (
	"context"
	"errors"

	auth "github.com/GarroshIcecream/yummy/internal/auth/chatgpt"
)

// ChatGPTSignedIn reports whether a ChatGPT session is stored in dataDir.
func ChatGPTSignedIn(dataDir string) bool {
	_, err := auth.NewStore(dataDir).Load()
	return err == nil
}

// ChatGPTLogin runs the browser sign-in and stores the session. openBrowser
// receives the authorization URL. It returns when ctx is done.
func ChatGPTLogin(ctx context.Context, dataDir string, openBrowser func(url string) error) (*auth.Tokens, error) {
	tokens, err := auth.NewClient().Login(ctx, openBrowser)
	if err != nil {
		return nil, err
	}
	if err := SaveChatGPTSession(dataDir, tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

// SaveChatGPTSession stores a session and makes running providers use it.
func SaveChatGPTSession(dataDir string, tokens *auth.Tokens) error {
	if tokens == nil {
		return errors.New("no ChatGPT session to save")
	}
	if err := auth.NewStore(dataDir).Save(tokens); err != nil {
		return err
	}
	sharedTokenSource(dataDir).Invalidate()
	return nil
}

// ChatGPTLogout removes the stored session and forgets it in memory.
func ChatGPTLogout(dataDir string) error {
	if err := auth.NewStore(dataDir).Delete(); err != nil {
		return err
	}
	sharedTokenSource(dataDir).Invalidate()
	return nil
}
