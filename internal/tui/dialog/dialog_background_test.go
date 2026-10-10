package dialog

import (
	"strconv"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"

	common "github.com/GarroshIcecream/yummy/internal/models/common"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

// assertNoBareGap fails if any rendered line contains an SGR reset that is
// immediately followed by a space — i.e. a region that fell through to the
// terminal background instead of the dialog surface. This is the exact defect
// that produced the ragged header/row bands.
func assertNoBareGap(t *testing.T, name, out string) {
	t.Helper()
	for i, line := range strings.Split(out, "\n") {
		for _, reset := range []string{"\x1b[0m ", "\x1b[m "} {
			if strings.Contains(line, reset) {
				t.Errorf("%s line %d: reset followed by a bare space — background leaks (%q)", name, i, strconv.Quote(line))
				break
			}
		}
	}
}

func TestStateSelectorNoBareGap(t *testing.T) {
	theme := themes.NewDefaultTheme()
	m := &StateSelectorDialogCmp{
		states: []common.SessionState{
			common.SessionStateMainMenu,
			common.SessionStateList,
			common.SessionStateDetail,
			common.SessionStateEdit,
			common.SessionStateChat,
		},
		selectedIndex: 0,
		theme:         &theme,
		width:         70,
		height:        20,
	}
	assertNoBareGap(t, "state_selector", m.View().Content)
}

func TestAddRecipeFromURLNoBareGap(t *testing.T) {
	theme := themes.NewDefaultTheme()
	ti := textinput.New()
	ti.SetStyles(theme.DialogSearchStyles)
	m := &AddRecipeFromURLDialogCmp{
		urlInput: ti,
		theme:    &theme,
		width:    60,
		height:   14,
	}
	assertNoBareGap(t, "add_recipe_from_url", m.View().Content)
}

func TestScraperUpdateNoBareGap(t *testing.T) {
	theme := themes.NewDefaultTheme()
	m := &ScraperUpdateDialogCmp{
		theme:     &theme,
		width:     60,
		height:    12,
		phase:     scraperPhaseReady,
		installed: "1.0.0",
		latest:    "1.1.0",
	}
	assertNoBareGap(t, "scraper_update", m.View().Content)
}

func TestRatingNoBareGap(t *testing.T) {
	theme := themes.NewDefaultTheme()
	m := &RatingDialogCmp{theme: &theme, cursor: 3}
	assertNoBareGap(t, "rating", m.View().Content)
}

func TestCommandPaletteNoBareGap(t *testing.T) {
	theme := themes.NewDefaultTheme()
	ti := textinput.New()
	ti.SetStyles(theme.DialogSearchStyles)
	ti.SetWidth(60)

	items := []CommandItem{
		{Name: "Switch View", Shortcut: "ctrl+s"},
		{Name: "Change Theme", Shortcut: "ctrl+t"},
		{Name: "Check recipe-scrapers Update"},
	}
	m := &CommandPaletteDialogCmp{
		allItems:    items,
		filtered:    items,
		searchInput: ti,
		theme:       &theme,
		width:       70,
		height:      20,
	}
	assertNoBareGap(t, "command_palette", m.View().Content)
}
