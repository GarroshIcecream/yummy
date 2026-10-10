package tools

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"

	"github.com/GarroshIcecream/yummy/internal/db"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
)

type GetRecipeIdTool struct {
	FunctionName        string            `json:"name"`
	FunctionDescription string            `json:"description"`
	CallbackHandler     callbacks.Handler `json:"callback_handler"`
	Cookbook            *db.CookBook      `json:"cookbook"`
}

var _ tools.Tool = &GetRecipeIdTool{}

func NewGetRecipeIdTool(cookbook *db.CookBook) *GetRecipeIdTool {
	return &GetRecipeIdTool{
		FunctionName: "getRecipeById",
		FunctionDescription: "Get the complete recipe (ingredients, steps, times, rating, source) for a recipe ID. " +
			`Input: the numeric ID, e.g. 12. Find IDs with searchRecipeByName, listRecipes or findRecipesByIngredients.`,
		Cookbook: cookbook,
	}
}

func (t *GetRecipeIdTool) Name() string {
	return t.FunctionName
}

func (t *GetRecipeIdTool) Description() string {
	return t.FunctionDescription
}

// firstNumber pulls the ID out of inputs like "12", "#12", `{"id": 12}` or
// "ID: 12".
var firstNumber = regexp.MustCompile(`\d+`)

func (t *GetRecipeIdTool) Call(ctx context.Context, input string) (string, error) {
	slog.Debug("Executing tool", "tool", t.FunctionName, "input", input)
	match := firstNumber.FindString(input)
	if match == "" {
		return fmt.Sprintf("%q is not a recipe ID. Pass the numeric ID, e.g. 12.", input), nil
	}
	recipeID, err := strconv.Atoi(match)
	if err != nil {
		return fmt.Sprintf("%q is not a recipe ID. Pass the numeric ID, e.g. 12.", input), nil
	}

	recipe, err := t.Cookbook.GetFullRecipe(uint(recipeID))
	if err != nil || recipe == nil {
		return fmt.Sprintf("No recipe with ID %d. Use searchRecipeByName or listRecipes to find valid IDs.", recipeID), nil
	}
	return fmt.Sprintf("[ID %d]\n%s", recipe.RecipeID, recipe.FormatRecipeMarkdown()), nil
}
