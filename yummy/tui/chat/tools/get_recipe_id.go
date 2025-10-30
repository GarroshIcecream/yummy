package tools

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/GarroshIcecream/yummy/yummy/db"
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
		FunctionName:        "getRecipeById",
		FunctionDescription: "Get a complete recipe by its ID from the cookbook (database stored recipes)",
		Cookbook:            cookbook,
	}
}

func (t *GetRecipeIdTool) Name() string {
	return t.FunctionName
}

func (t *GetRecipeIdTool) Description() string {
	return t.FunctionDescription
}

func (t *GetRecipeIdTool) Call(ctx context.Context, input string) (string, error) {
	slog.Debug("Executing tool", "tool", t.FunctionName, "input", input)
	recipeIDInt, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid recipe ID: %w", err)
	}

	recipeRaw, err := t.Cookbook.GetFullRecipe(uint(recipeIDInt))
	if err != nil {
		return fmt.Sprintf("Recipe with ID %d not found", recipeIDInt), nil
	}
	return recipeRaw.FormatRecipeMarkdown(), nil
}
