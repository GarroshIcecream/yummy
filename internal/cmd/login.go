package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GarroshIcecream/yummy/internal/auth/chatgpt"
	"github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/spf13/cobra"
)

// dataDirFunc resolves the Yummy data directory; tests override it.
var dataDirFunc = resolveUserDir

// loginTimeout bounds how long login waits for the browser sign-in.
const loginTimeout = 5 * time.Minute

// normalizeLoginProvider maps accepted provider names to "chatgpt".
func normalizeLoginProvider(name string) (string, error) {
	switch strings.ToLower(name) {
	case "chatgpt", "openai":
		return "chatgpt", nil
	default:
		return "", fmt.Errorf("unknown provider %q (supported: chatgpt)", name)
	}
}

var loginCmd = &cobra.Command{
	Use:       "login [chatgpt]",
	Short:     "Sign in to an LLM provider",
	Long:      "Sign in with your ChatGPT Plus/Pro account so the cooking assistant can use your subscription instead of a local Ollama model.",
	Example:   "  yummy login chatgpt",
	ValidArgs: []string{"chatgpt", "openai"},
	Args:      cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "chatgpt"
		if len(args) > 0 {
			name = args[0]
		}
		if _, err := normalizeLoginProvider(name); err != nil {
			return err
		}
		datadir, err := dataDirFunc()
		if err != nil {
			return fmt.Errorf("failed to resolve user directory: %w", err)
		}

		parent := cmd.Context()
		if parent == nil {
			parent = context.Background()
		}
		ctx, cancel := context.WithTimeout(parent, loginTimeout)
		defer cancel()

		out := cmd.OutOrStdout()
		tokens, err := chatgpt.NewClient().Login(ctx, func(authURL string) error {
			_, _ = fmt.Fprintf(out, "Opening your browser to sign in to ChatGPT.\nIf it doesn't open, visit:\n\n  %s\n\n", authURL)
			return utils.OpenBrowser(authURL)
		})
		if err != nil {
			return err
		}
		if err := chatgpt.NewStore(datadir).Save(tokens); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "Logged in to ChatGPT (account %s).\n", tokens.AccountID)
		_, _ = fmt.Fprintln(out, `To use it, set "chat": {"provider": "chatgpt"} in ~/.yummy/config.json.`)
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:       "logout [chatgpt]",
	Short:     "Sign out of an LLM provider",
	Example:   "  yummy logout chatgpt",
	ValidArgs: []string{"chatgpt", "openai"},
	Args:      cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "chatgpt"
		if len(args) > 0 {
			name = args[0]
		}
		if _, err := normalizeLoginProvider(name); err != nil {
			return err
		}
		datadir, err := dataDirFunc()
		if err != nil {
			return fmt.Errorf("failed to resolve user directory: %w", err)
		}
		if err := chatgpt.NewStore(datadir).Delete(); err != nil {
			return fmt.Errorf("remove ChatGPT session: %w", err)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Logged out of ChatGPT.")
		return nil
	},
}
