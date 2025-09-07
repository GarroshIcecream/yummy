package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	log "github.com/GarroshIcecream/yummy/yummy/log"
	tui "github.com/GarroshIcecream/yummy/yummy/tui"
	"github.com/GarroshIcecream/yummy/yummy/version"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().StringP("cwd", "c", "", "Current working directory")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug")

	rootCmd.Flags().BoolP("help", "h", false, "Help")

	rootCmd.AddCommand(exportCmd)
}

var rootCmd = &cobra.Command{
	Use:   "yummy",
	Short: "Terminal-based cookbook manager and recipe assistant",
	Long:  `Yummy is cool`,
	Example: `
# Run in interactive mode
yummy

# Run with debug logging
yummy -d

# Run with debug logging in a specific directory
yummy -d -c /path/to/project

# Print version
yummy -v
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := setupApp(cmd)
		if err != nil {
			return err
		}

		program := tea.NewProgram(
			app,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
			tea.WithMouseAllMotion(),
			tea.WithContext(cmd.Context()),
		)

		if _, err := program.Run(); err != nil {
			slog.Error("TUI run error", "error", err)
			return fmt.Errorf("TUI error: %v", err)
		}
		return nil
	},
}

func Execute() {
	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(version.Version),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}

func ResolveCwd(cmd *cobra.Command) (string, error) {
	cwd, _ := cmd.Flags().GetString("cwd")
	if cwd != "" {
		err := os.Chdir(cwd)
		if err != nil {
			return "", fmt.Errorf("failed to change directory: %v", err)
		}
		return cwd, nil
	}

	cwd, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}
	return cwd, nil
}

func setupApp(cmd *cobra.Command) (*tui.Manager, error) {
	// debug, _ := cmd.Flags().GetBool("debug")
	ctx := cmd.Context()

	cwd, err := ResolveCwd(cmd)
	if err != nil {
		return nil, err
	}

	// cfg, err := config.Init(cwd, debug)
	// if err != nil {
	// 	return nil, err
	// }

	datadir := filepath.Join(cwd, ".yummy")
	if err := os.MkdirAll(datadir, 0755); err != nil {
		slog.Error("Error creating database directory", "error", err)
	}

	// setup log
	log.Setup(datadir, true)
	
	conn, err := db.NewCookBook(datadir)
	if err != nil {
		return nil, err
	}

	tuiInstance, err := tui.New(conn, ctx)
	if err != nil {
		slog.Error("Failed to create tui instance", "error", err)
		return nil, err
	}

	return tuiInstance, nil
}
