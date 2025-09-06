package cmd

import (
	"fmt"

	tui "github.com/GarroshIcecream/yummy/yummy/tui"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [recipe_id]",
	Short: "Export a recipe to a file",
	Long: `Export a recipe to a file.
		The recipe can be provided as an argument.`,
	Example: `
		# Export a recipe to a file
		yummy export 123
  	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		recipe_id, _ := cmd.Flags().GetString("recipe_id")

		app, err := tui.New(cmd.Context(), recipe_id)
		if err != nil {
			return err
		}
		defer app.Shutdown()

		if recipe_id == "" {
			return fmt.Errorf("no recipe id provided")
		}

		return app.RunNonInteractive(cmd.Context(), recipe_id, quiet)
	},
}

func init() {
	exportCmd.Flags().BoolP("quiet", "q", false, "Hide spinner")
}
