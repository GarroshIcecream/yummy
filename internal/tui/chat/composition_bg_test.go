package chat

import (
	"strconv"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/GarroshIcecream/yummy/internal/config"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

// TestChatCompositionNoSeam reproduces View()'s sidebar branch and guards the
// inter-panel seams: the 2-column gutter and the JoinHorizontal padding below
// the shorter sidebar column must be painted with the base canvas color, not
// left as bare cells that leak the terminal background. The sidebar's own
// elevated surface must survive the fill.
func TestChatCompositionNoSeam(t *testing.T) {
	theme := themes.NewDefaultTheme()

	const contentWidth = 50
	const sidebarWidth = 30
	const total = contentWidth + 4 + 2 + sidebarWidth

	body := strings.TrimRight(strings.Repeat("conversation line\n", 6), "\n")
	chatPanel := renderTitledPanel(&theme, body, theme.ChatPanelTitle.Render("🍳 Yummy"), "", contentWidth, theme.BaseBg)
	inputPanel := renderTitledPanel(&theme, utils.FillBackground("› ", theme.BaseBg, contentWidth), "", "", contentWidth, theme.BaseBg)
	mainContent := lipgloss.JoinVertical(lipgloss.Left, chatPanel, inputPanel)

	sidebar := RenderSidebar(nil, config.NewDefaultKeyBindings().ToKeyMap().GetChatKeyMap(), &theme, sidebarWidth, 6)
	gap := lipgloss.NewStyle().MarginLeft(2).Render(sidebar)
	combined := lipgloss.JoinHorizontal(lipgloss.Top, mainContent, gap)
	out := utils.FillBackground(combined, theme.BaseBg, total)

	// The whole chat page is one flat dark surface: the base canvas must be
	// present and no elevated tint (#1D2330) may leak into the chat page.
	if !strings.Contains(out, "48;2;15;17;21m") { // #0F1115
		t.Errorf("base canvas not established across the composition")
	}
	if strings.Contains(out, "48;2;29;35;48m") { // #1D2330
		t.Errorf("elevated tint leaked into the flat chat page")
	}
	// No reset may be immediately followed by a bare space — every seam is filled.
	for i, line := range strings.Split(out, "\n") {
		for _, reset := range []string{"\x1b[0m ", "\x1b[m "} {
			if strings.Contains(line, reset) {
				t.Errorf("composition line %d: reset followed by a bare space — seam leaks (%s)", i, strconv.Quote(line))
				break
			}
		}
	}
}
