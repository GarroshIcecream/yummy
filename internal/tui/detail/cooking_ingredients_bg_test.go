package detail

import (
	"strconv"
	"strings"
	"testing"

	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

// TestCookingIngredientsSidebarUniformSurface guards the ingredients sidebar
// against the two-tone class: every cell must sit on the flat base canvas
// (#0F1115), matching the rest of the app's inline chrome. Neither the old
// mismatched frame (#171A21) nor the elevated tint (#1D2330) may appear.
func TestCookingIngredientsSidebarUniformSurface(t *testing.T) {
	theme := themes.NewDefaultTheme()

	m := &CookingModel{
		theme: &theme,
		Recipe: &utils.RecipeRaw{
			RecipeName: "Test",
			Metadata: utils.RecipeMetadata{
				Ingredients: []utils.Ingredient{
					{Amount: "200", Unit: "g", Name: "flour", Group: "Dough"},
					{Amount: "2", Name: "eggs", Details: "room temp"},
				},
			},
		},
	}

	out := m.renderIngredientsSidebar(30, 20)

	baseBg := bgSeq("#0F1115")     // must appear (the uniform surface)
	strayBg := bgSeq("#171A21")    // must NOT appear
	elevatedBg := bgSeq("#1D2330") // must NOT appear

	if !strings.Contains(out, baseBg) {
		t.Errorf("ingredients sidebar never establishes the base canvas (%s)", baseBg)
	}
	if strings.Contains(out, strayBg) {
		t.Errorf("ingredients sidebar contains the old mismatched surface (%s)", strayBg)
	}
	if strings.Contains(out, elevatedBg) {
		t.Errorf("ingredients sidebar contains the elevated tint (%s) — should be flat base canvas", elevatedBg)
	}
	for i, line := range strings.Split(out, "\n") {
		for _, reset := range []string{"\x1b[0m ", "\x1b[m "} {
			if strings.Contains(line, reset) {
				t.Errorf("ingredients sidebar line %d: reset followed by a bare space (%s)", i, strconv.Quote(line))
				break
			}
		}
	}
}
