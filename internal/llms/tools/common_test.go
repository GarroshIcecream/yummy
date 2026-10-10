package tools

import (
	"reflect"
	"testing"
	"time"

	"github.com/GarroshIcecream/yummy/internal/utils"
)

func TestParseArgsAcceptsJSONPairsAndBareValues(t *testing.T) {
	tests := []struct {
		in   string
		want map[string]string
	}{
		{`{"category": "Dessert", "page": 2}`, map[string]string{"category": "Dessert", "page": "2"}},
		{`{"ingredients": ["chicken", "rice"]}`, map[string]string{"ingredients": "chicken,rice"}},
		{`category=dessert; page=2`, map[string]string{"category": "dessert", "page": "2"}},
		{`ingredient=chicken, rice; sort=quickest`, map[string]string{"ingredient": "chicken,rice", "sort": "quickest"}},
		{`"pizza dough"`, map[string]string{"query": "pizza dough"}},
		{"", map[string]string{}},
	}
	for _, tt := range tests {
		if got := parseArgs(tt.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseArgs(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestPaginateClampsPage(t *testing.T) {
	start, end, pages := paginate(23, 3, 10)
	if start != 20 || end != 23 || pages != 3 {
		t.Fatalf("page 3 of 23 items: got %d-%d of %d pages", start, end, pages)
	}
	start, end, _ = paginate(23, 9, 10)
	if start != 20 || end != 23 {
		t.Fatalf("out-of-range page should clamp to the last page, got %d-%d", start, end)
	}
}

func TestListFiltersAndSorting(t *testing.T) {
	quick := utils.RecipeRaw{RecipeName: "Quick", Metadata: utils.RecipeMetadata{
		TotalTime:   20 * time.Minute,
		Categories:  []string{"Dessert"},
		Ingredients: []utils.Ingredient{{Name: "dark chocolate", BaseName: "chocolate"}},
	}}
	slow := utils.RecipeRaw{RecipeName: "Slow", Metadata: utils.RecipeMetadata{
		TotalTime:  2 * time.Hour,
		Categories: []string{"Main Dish"},
	}}
	unknown := utils.RecipeRaw{RecipeName: "Unknown"}

	if !hasCategory(quick, "desserts") || hasCategory(slow, "dessert") {
		t.Error("category filter should match case-insensitively and ignore a plural s")
	}
	if !allIngredients(quick, []string{"chocolate"}) || anyIngredient(slow, []string{"chocolate"}) {
		t.Error("ingredient filter should match ingredient names")
	}

	recipes := []utils.RecipeRaw{unknown, slow, quick}
	sortRecipes(recipes, "quickest")
	if recipes[0].RecipeName != "Quick" || recipes[2].RecipeName != "Unknown" {
		t.Errorf("quickest should order by total time with unknown times last, got %v", []string{recipes[0].RecipeName, recipes[1].RecipeName, recipes[2].RecipeName})
	}
}
