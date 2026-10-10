package chat

import (
	"strconv"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textarea"

	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

// TestComposerFillsSurface guards the chat composer against the ragged-input
// defect: the textarea does not paint its full width on its own, so empty and
// short lines leave bare cells that fall through to the terminal background.
// The composer must be filled to the base canvas and framed on that same
// surface so it reads as one uniform flat field — with no stray #222938 (an old
// input tint), no #1D2330 (the old elevated tint), and no SGR reset immediately
// followed by a bare space.
func TestComposerFillsSurface(t *testing.T) {
	theme := themes.NewDefaultTheme()

	ta := textarea.New()
	ta.Placeholder = "Ask about cooking, recipes, ingredients..."
	ta.SetStyles(theme.TextAreaStyles())
	ta.Focus()
	ta.SetWidth(60)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.Prompt = "› "

	composer := utils.FillBackground(ta.View(), theme.BaseBg, 60)
	out := renderTitledPanel(&theme, composer, "", "", 60, theme.BaseBg)

	if strings.Contains(out, "48;2;34;41;56m") { // #222938
		t.Errorf("composer still renders the old #222938 tint — should sit on the base canvas")
	}
	if strings.Contains(out, "48;2;29;35;48m") { // #1D2330
		t.Errorf("composer still renders the old elevated tint — should be flat on the base canvas")
	}

	for i, line := range strings.Split(out, "\n") {
		for _, reset := range []string{"\x1b[0m ", "\x1b[m "} {
			if strings.Contains(line, reset) {
				t.Errorf("composer line %d: reset followed by a bare space — background leaks (%s)", i, strconv.Quote(line))
				break
			}
		}
	}
}
