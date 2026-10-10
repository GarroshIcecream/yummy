package chat

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/textarea"
	"github.com/charmbracelet/x/ansi"

	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

// TestComposerPromptDoesNotShiftWhenTyping guards the bug where the prompt sat
// one column further right while the placeholder showed and jumped left on the
// first keystroke (textarea Base padding was applied twice to the prompt in
// placeholder mode).
func TestComposerPromptDoesNotShiftWhenTyping(t *testing.T) {
	theme := themes.NewDefaultTheme()
	ta := textarea.New()
	ta.Placeholder = "Ask about cooking"
	ta.SetStyles(theme.TextAreaStyles())
	ta.ShowLineNumbers = false
	ta.SetWidth(40)
	ta.SetHeight(3)
	ta.SetPromptFunc(themes.ComposerPromptWidth, themes.ComposerPrompt)
	ta.Focus()

	promptCol := func() int {
		first, _, _ := strings.Cut(ansi.Strip(ta.View()), "\n")
		return strings.Index(first, ">")
	}
	empty := promptCol()
	ta.InsertString("@")
	if typed := promptCol(); typed != empty {
		t.Fatalf("prompt moved from column %d to %d after typing", empty, typed)
	}
}
