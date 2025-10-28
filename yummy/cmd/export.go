package cmd

import (
	"fmt"
	"log/slog"
	"os"

	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	recipes "github.com/GarroshIcecream/yummy/yummy/recipe"
	"github.com/GarroshIcecream/yummy/yummy/tui/detail"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [recipe_id]",
	Short: "Export a recipe to a file",
	Long:  `Export a recipe to a file. The recipe can be provided as an argument.`,
	Example: `
		# Export a recipe to a file
		yummy export 123
  	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			slog.Error("Recipe ID is required", "args", args)
			return fmt.Errorf("recipe ID is required")
		}

		// Parse recipe ID from arguments
		var recipe_id uint
		if _, err := fmt.Sscanf(args[0], "%d", &recipe_id); err != nil {
			slog.Error("Invalid recipe ID", "recipeID", args[0], "error", err)
			return fmt.Errorf("invalid recipe ID: %s", args[0])
		}

		tui, err := setupApp(cmd)
		if err != nil {
			slog.Error("Failed to create TUI instance", "error", err)
			return err
		}

		msg := tui.GetModel(consts.SessionStateDetail).(*detail.DetailModel).FetchRecipe(recipe_id)
		if msg.Recipe == nil {
			slog.Error("Failed to fetch recipe", "recipeID", recipe_id)
			return fmt.Errorf("failed to fetch recipe")
		}

		// Export the recipe to a file
		filename := fmt.Sprintf("recipe_%d.md", recipe_id)
		content := recipes.FormatRecipeContent(msg.Recipe)

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			slog.Error("Failed to write file", "filename", filename, "error", err)
			return fmt.Errorf("failed to write file %s: %v", filename, err)
		}

		slog.Info("Recipe exported successfully", "filename", filename)
		fmt.Printf("Recipe exported to %s\n", filename)
		return nil
	},
}
