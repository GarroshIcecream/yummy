package tools

import (
	"context"
	"fmt"
	"log/slog"
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

func NewGetRecipeNameTool(cookbook *db.CookBook) *GetRecipeNameTool {
	return &GetRecipeNameTool{
		FunctionName:        "searchRecipeByName",
		FunctionDescription: "Search for recipes by name (case-insensitive partial match)",
		Cookbook:            cookbook,
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
	allRecipes, err := t.Cookbook.AllRecipes()
	if err != nil {
		return "", fmt.Errorf("failed to fetch recipes: %w", err)
	}

	// Filter recipes by name (case-insensitive partial match)
	var matches []utils.RecipeRaw
	searchName := strings.ToLower(input)

	for _, r := range allRecipes {
		if strings.Contains(strings.ToLower(r.RecipeName), searchName) {
			matches = append(matches, r)
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No recipes found matching '%s'", input), nil
	}

	// Format results
	var result strings.Builder
	if len(matches) == 1 {
		result.WriteString(fmt.Sprintf("Found 1 recipe matching '%s':\n\n", input))
	} else {
		result.WriteString(fmt.Sprintf("Found %d recipes matching '%s':\n\n", len(matches), input))
	}

	for i, match := range matches {
		result.WriteString(fmt.Sprintf("%d. **%s** (ID: %d)\n", i+1, match.RecipeName, match.RecipeID))
		if match.RecipeDescription != "" {
			result.WriteString(fmt.Sprintf("   Description: %s\n", match.RecipeDescription))
		}
		if match.Metadata.Author != "" {
			result.WriteString(fmt.Sprintf("   Author: %s\n", match.Metadata.Author))
		}
		if len(match.Metadata.Categories) > 0 {
			result.WriteString(fmt.Sprintf("   Categories: %s\n", strings.Join(match.Metadata.Categories, ", ")))
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}
