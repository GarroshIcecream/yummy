package cmd

import (
	"fmt"
	"os"

	"github.com/GarroshIcecream/yummy/yummy/tui/detail"
	"github.com/GarroshIcecream/yummy/yummy/ui"
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
			return fmt.Errorf("recipe ID is required")
		}

		// Parse recipe ID from arguments
		var recipe_id uint
		if _, err := fmt.Sscanf(args[0], "%d", &recipe_id); err != nil {
			return fmt.Errorf("invalid recipe ID: %s", args[0])
		}

		tui, err := setupApp(cmd)
		if err != nil {
			return err
		}

		msg := tui.GetModel(ui.SessionStateDetail).(*detail.DetailModel).FetchRecipe(recipe_id)
		if msg.Err != nil {
			return msg.Err
		}

		// Export the recipe to a file
		filename := fmt.Sprintf("recipe_%d.md", recipe_id)
		content := msg.Content

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %v", filename, err)
		}

		fmt.Printf("Recipe exported to %s\n", filename)
		return nil
	},
}

func init() {
	exportCmd.Flags().BoolP("quiet", "q", false, "Hide spinner")
}
