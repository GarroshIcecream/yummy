package tools

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
)

type GetRecipeNameTool struct {
	FunctionName        string            `json:"name"`
	FunctionDescription string            `json:"description"`
	CallbackHandler     callbacks.Handler `json:"callback_handler"`
	Cookbook            *db.CookBook      `json:"cookbook"`
}

var _ tools.Tool = &GetRecipeNameTool{}

const searchLimit = 10

func NewGetRecipeNameTool(cookbook *db.CookBook) *GetRecipeNameTool {
	return &GetRecipeNameTool{
		FunctionName: "searchRecipeByName",
		FunctionDescription: "Search the cookbook by text. Matches recipe names first, then descriptions, " +
			"categories and authors (case-insensitive; every word must match). " +
			`Input: the search text, e.g. pizza dough. Returns recipe IDs for getRecipeById.`,
		Cookbook: cookbook,
	}
}

func (t *GetRecipeNameTool) Name() string {
	return t.FunctionName
}

func (t *GetRecipeNameTool) Description() string {
	return t.FunctionDescription
}

func (t *GetRecipeNameTool) Call(ctx context.Context, input string) (string, error) {
	slog.Debug("Executing tool", "tool", t.FunctionName, "input", input)
	args := parseArgs(input)
	query := args["query"]
	if query == "" {
		query = args["name"]
	}
	if query == "" {
		return "Pass the text to search for, e.g. pizza dough. To browse everything use listRecipes.", nil
	}

	recipes, err := t.Cookbook.AllRecipes()
	if err != nil {
		return "", fmt.Errorf("failed to fetch recipes: %w", err)
	}

	// Score: 3 = every word in the name, 1 = every word somewhere else.
	words := strings.Fields(strings.ToLower(query))
	type scored struct {
		recipe utils.RecipeRaw
		score  int
	}
	var matches []scored
	for _, r := range recipes {
		other := strings.Join(append([]string{r.RecipeDescription, r.Metadata.Author}, r.Metadata.Categories...), " ")
		inName, inAny := true, true
		for _, w := range words {
			name := containsFold(r.RecipeName, w)
			inName = inName && name
			inAny = inAny && (name || containsFold(other, w))
		}
		switch {
		case inName:
			matches = append(matches, scored{r, 3})
		case inAny:
			matches = append(matches, scored{r, 1})
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No recipes match %q. Try fewer words, findRecipesByIngredients, or listRecipes to browse.", query), nil
	}
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].score > matches[j].score })

	var b strings.Builder
	fmt.Fprintf(&b, "%d recipe(s) match %q", len(matches), query)
	if len(matches) > searchLimit {
		fmt.Fprintf(&b, " (showing the best %d)", searchLimit)
	}
	b.WriteString(":\n")
	for i, m := range matches {
		if i == searchLimit {
			break
		}
		b.WriteString(summaryLine(m.recipe) + "\n")
	}
	b.WriteString("\nUse getRecipeById with an ID for the full recipe.")
	return b.String(), nil
}
