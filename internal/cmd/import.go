package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/spf13/cobra"
)

func init() {
	importCmd.Flags().StringP("name", "n", "", "Custom name for the imported recipe")
}

var importCmd = &cobra.Command{
	Use:   "import [file_path]",
	Short: "Import a recipe from a markdown or JSON file",
	Long: `Import a recipe from a file. Supports both markdown (.md) and JSON (.json) formats.
The markdown format should match the export format used by yummy export command.`,
	Example: `
		# Import a recipe from markdown file
		yummy import recipe.md

		# Import a recipe from JSON file
		yummy import recipe.json

		# Import with custom name
		yummy import recipe.md --name "My Custom Recipe"
  	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		customName, _ := cmd.Flags().GetString("name")

		if len(args) == 0 {
			slog.Error("File path is required", "args", args)
			return fmt.Errorf("file path is required")
		}

		filePath := args[0]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			slog.Error("File does not exist", "path", filePath, "error", err)
			return fmt.Errorf("file does not exist: %s", filePath)
		}

		// Determine file format
		ext := strings.ToLower(filepath.Ext(filePath))
		var recipeRaw *utils.RecipeRaw
		var err error

		switch ext {
		case ".md":
			recipeRaw, err = utils.ParseMarkdownRecipe(filePath, customName)
		case ".json":
			recipeRaw, err = utils.ParseJSONRecipe(filePath, customName)
		default:
			slog.Error("Unsupported file format", "format", ext)
			return fmt.Errorf("unsupported file format: %s. Supported formats: .md, .json", ext)
		}

		if err != nil {
			slog.Error("Failed to parse recipe", "filePath", filePath, "error", err)
			return fmt.Errorf("failed to parse recipe: %v", err)
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

		// Save the recipe
		recipeID, err := cookbook.SaveScrapedRecipe(recipeRaw)
		if err != nil {
			slog.Error("Failed to save recipe", "error", err)
			return fmt.Errorf("failed to save recipe: %v", err)
		}

		slog.Info("Recipe imported successfully", "filePath", filePath, "id", recipeID)
		fmt.Printf("✅ Recipe imported successfully! ID: %d\n", recipeID)
		return nil
	},
}
