package tools

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/tmc/langchaingo/tools"
)

// FindByIngredientsTool answers "what can I cook with X, Y and Z?".
type FindByIngredientsTool struct {
	Cookbook *db.CookBook
}

var _ tools.Tool = &FindByIngredientsTool{}

func NewFindByIngredientsTool(cookbook *db.CookBook) *FindByIngredientsTool {
	return &FindByIngredientsTool{Cookbook: cookbook}
}

func (t *FindByIngredientsTool) Name() string { return "findRecipesByIngredients" }

func (t *FindByIngredientsTool) Description() string {
	return "Find recipes that use the ingredients the user has, ranked by how many of them each recipe uses. " +
		"Input: comma-separated ingredients, e.g. chicken, rice, garlic."
}

func (t *FindByIngredientsTool) Call(ctx context.Context, input string) (string, error) {
	slog.Debug("Executing tool", "tool", t.Name(), "input", input)
	args := parseArgs(input)
	raw := args["query"]
	for _, k := range []string{"ingredients", "ingredient"} {
		if v := args[k]; v != "" {
			raw = v
		}
	}
	var terms []string
	for _, p := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ';' }) {
		p = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(p), "and "))
		if p = unquote(p); p != "" {
			terms = append(terms, strings.ToLower(p))
		}
	}
	if len(terms) == 0 {
		return "Pass the ingredients to look for, e.g. chicken, rice, garlic.", nil
	}

	recipes, err := fullRecipes(t.Cookbook)
	if err != nil {
		return "", err
	}

	type hit struct {
		recipe  utils.RecipeRaw
		matched []string
	}
	var hits []hit
	for _, r := range recipes {
		var matched []string
		for _, term := range terms {
			if ingredientMatches(r, term) {
				matched = append(matched, term)
			}
		}
		if len(matched) > 0 {
			hits = append(hits, hit{r, matched})
		}
	}
	if len(hits) == 0 {
		return fmt.Sprintf("No recipes use any of: %s.", strings.Join(terms, ", ")), nil
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if len(hits[i].matched) != len(hits[j].matched) {
			return len(hits[i].matched) > len(hits[j].matched)
		}
		// Fewer ingredients overall = fewer things the user still needs.
		return len(hits[i].recipe.Metadata.Ingredients) < len(hits[j].recipe.Metadata.Ingredients)
	})

	var b strings.Builder
	fmt.Fprintf(&b, "%d recipe(s) use some of: %s\n", len(hits), strings.Join(terms, ", "))
	for i, h := range hits {
		if i == searchLimit {
			fmt.Fprintf(&b, "...and %d more.\n", len(hits)-searchLimit)
			break
		}
		fmt.Fprintf(&b, "%s - uses %d/%d (%s), needs %d ingredients in total\n",
			summaryLine(h.recipe), len(h.matched), len(terms), strings.Join(h.matched, ", "),
			len(h.recipe.Metadata.Ingredients))
	}
	return b.String(), nil
}
