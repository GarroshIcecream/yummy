package detail

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"

	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

// bgSeq returns the truecolor background SGR lipgloss emits for a hex color.
func bgSeq(hex string) string {
	c := lipgloss.Color(hex)
	r, g, b, _ := c.RGBA()
	return "48;2;" + strconv.Itoa(int(r>>8)) + ";" + strconv.Itoa(int(g>>8)) + ";" + strconv.Itoa(int(b>>8))
}

// The cooking chat panel must sit uniformly on the flat base canvas (#0F1115),
// matching the main chat page — separated from the recipe area only by its
// border, not by color. Every cell sits on the base canvas: the base must be
// present, and neither the old mismatched frame (#171A21) nor the elevated
// tint (#1D2330) may appear.
func TestCookingChatPanelHasNoBaseBgLeak(t *testing.T) {
	theme := themes.NewDefaultTheme()

	vp := viewport.New(viewport.WithWidth(40), viewport.WithHeight(10))
	vp.Style = lipgloss.NewStyle().Background(theme.BaseBg)

	ta := textarea.New()
	ta.SetStyles(theme.TextAreaStyles())
	ta.Prompt = "› "
	ta.ShowLineNumbers = false

	m := &CookingModel{
		theme:        &theme,
		chatViewport: vp,
		chatTextarea: ta,
		chatHistory: []chatEntry{
			{role: "user", content: "whats up"},
			{role: "assistant", content: "Hello! Ready to cook?"},
		},
		chatWaiting: true, // also exercises the "Thinking..." spinner path
	}

	out := m.renderChatPanel(44, 24)

	baseBg := bgSeq("#0F1115")     // must appear (the uniform surface)
	strayBg := bgSeq("#171A21")    // must NOT appear (old mismatched frame)
	elevatedBg := bgSeq("#1D2330") // must NOT appear (old elevated tint)

	if !strings.Contains(out, baseBg) {
		_ = os.WriteFile("/tmp/cooking_panel.txt", []byte(strconv.Quote(out)), 0o644)
		t.Errorf("cooking panel never establishes the base canvas (%s); see /tmp/cooking_panel.txt", baseBg)
	}
	if strings.Contains(out, strayBg) {
		t.Errorf("cooking panel contains the old mismatched surface (%s) — should be flat base canvas", strayBg)
	}
	if strings.Contains(out, elevatedBg) {
		t.Errorf("cooking panel contains the elevated tint (%s) — chat panel should be flat on the base canvas", elevatedBg)
	}
}
