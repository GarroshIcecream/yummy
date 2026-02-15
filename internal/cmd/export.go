package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
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

		// Resolve user directory for data storage
		datadir, err := resolveUserDir()
		if err != nil {
			return fmt.Errorf("failed to resolve user directory: %v", err)
		}

		// Load configuration
		cfg, err := config.LoadConfig(datadir)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %v", err)
		}

		cookbook, err := db.NewCookBook(datadir, &cfg.Database)
		if err != nil {
			slog.Error("Failed to initialize cookbook", "error", err)
			return fmt.Errorf("failed to initialize cookbook: %v", err)
		}

		recipe, err := cookbook.GetFullRecipe(recipe_id)
		if err != nil {
			slog.Error("Failed to fetch recipe", "recipeID", recipe_id, "error", err)
			return fmt.Errorf("failed to fetch recipe: %v", err)
		}

		// Export the recipe to a file
		normalizedName := strings.ToLower(strings.ReplaceAll(recipe.RecipeName, " ", "_"))
		filename := fmt.Sprintf("%s.md", normalizedName)
		content := recipe.FormatRecipeMarkdown()

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			slog.Error("Failed to write file", "filename", filename, "error", err)
			return fmt.Errorf("failed to write file %s: %v", filename, err)
		}

		slog.Info("Recipe exported successfully", "filename", filename)
		fmt.Printf("✅ Recipe exported to %s\n", filename)
		return nil
	},
}
