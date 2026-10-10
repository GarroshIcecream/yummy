package detail

import (
	"strconv"
	"strings"
	"testing"

	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

// TestDetailViewSitsOnCanvas guards the recipe-detail view against rendering
// with no background — it must sit on the app's base canvas (#0F1115) so it
// doesn't leak the terminal background, with no reset-followed-by-bare-space.
func TestDetailViewSitsOnCanvas(t *testing.T) {
	theme := themes.NewDefaultTheme()
	m := &DetailModel{
		theme:           &theme,
		width:           44,
		height:          10,
		Recipe:          &utils.RecipeRaw{RecipeName: "Test"},
		renderedContent: "# Pasta\n\nBoil water.\nAdd salt.\nServe.",
	}

	out := m.View()

	if !strings.Contains(out, "48;2;15;17;21m") { // #0F1115
		t.Errorf("detail view does not establish the base canvas background")
	}
	for i, line := range strings.Split(out, "\n") {
		for _, reset := range []string{"\x1b[0m ", "\x1b[m "} {
			if strings.Contains(line, reset) {
				t.Errorf("detail view line %d: reset followed by a bare space (%s)", i, strconv.Quote(line))
				break
			}
		}
	}
}
