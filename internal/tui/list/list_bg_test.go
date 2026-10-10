package list

import (
	"strconv"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"

	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

// TestListViewSitsOnCanvas guards the recipe list against rendering with no
// background: it must sit uniformly on the base canvas (#0F1115), with no
// SGR reset immediately followed by a bare space (the 1-cell gaps the bubbles
// list leaves around the selected-item bar, title, and help line).
func TestListViewSitsOnCanvas(t *testing.T) {
	theme := themes.NewDefaultTheme()

	d := list.NewDefaultDelegate()
	d.Styles = theme.DelegateStyles
	items := []list.Item{
		utils.RecipeRaw{RecipeID: 1, RecipeName: "Carbonara"},
		utils.RecipeRaw{RecipeID: 2, RecipeName: "Tiramisu"},
	}
	l := list.New(items, d, 40, 12)
	l.Styles = theme.ListStyles
	l.Title = "Recipes"

	m := &ListModel{RecipeList: l, theme: &theme, width: 40, height: 12}
	m.SetSize(40, 12)

	out := m.View()

	if !strings.Contains(out, "48;2;15;17;21") { // #0F1115
		t.Errorf("list view does not establish the base canvas background")
	}
	for i, line := range strings.Split(out, "\n") {
		for _, reset := range []string{"\x1b[0m ", "\x1b[m "} {
			if strings.Contains(line, reset) {
				t.Errorf("list view line %d: reset followed by a bare space (%s)", i, strconv.Quote(line))
				break
			}
		}
	}
}
