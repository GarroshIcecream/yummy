package chat

import (
	"strconv"
	"strings"
	"testing"

	"github.com/GarroshIcecream/yummy/internal/config"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

// TestSidebarNoBareGap guards the sidebar against the ragged-background defect:
// every rendered line must sit uniformly on the elevated surface, with no SGR
// reset immediately followed by a bare space (a cell that fell through to the
// terminal background) and no second, mismatched surface color leaking through
// from a style that baked its own background.
func TestSidebarNoBareGap(t *testing.T) {
	theme := themes.NewDefaultTheme()

	out := RenderSidebar(nil, config.NewDefaultKeyBindings().ToKeyMap().GetChatKeyMap(), &theme, 30, 20)

	// #171A21 (the old, mismatched sidebar-content surface) must never appear —
	// the whole panel sits on SurfaceBg (#1D2330). Match the bare channel triple
	// so it is caught whether emitted standalone or fused with a foreground code.
	const straySurface = "48;2;23;26;33m" // #171A21
	if strings.Contains(out, straySurface) {
		t.Errorf("sidebar renders a mismatched #171A21 surface span — should be uniform SurfaceBg")
	}

	for i, line := range strings.Split(out, "\n") {
		for _, reset := range []string{"\x1b[0m ", "\x1b[m "} {
			if strings.Contains(line, reset) {
				t.Errorf("sidebar line %d: reset followed by a bare space — background leaks (%s)", i, strconv.Quote(line))
				break
			}
		}
	}
}
