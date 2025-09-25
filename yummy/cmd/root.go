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
	"github.com/GarroshIcecream/yummy/yummy/tui/chat"
	"github.com/GarroshIcecream/yummy/yummy/version"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug")
	rootCmd.Flags().BoolP("help", "h", false, "Help")

	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
}

var rootCmd = &cobra.Command{
	Use:   "yummy",
	Short: "Yummy - Terminal-based cookbook manager and recipe assistant",
	Long: `🍳 Yummy — Your Command-Line Recipe Companion

A fast, delightful command-line application for managing recipes. Built with care and powered by Bubble Tea, 
Yummy brings a beautiful terminal-first experience to every home cook, developer, and recipe curator.

🚀 Core Features:
• Recipe Management: Add, edit, and organize recipes with ingredient lists, measures, instructions, and metadata
• Powerful Search: Quick search and categorization to find the recipe you need
• Export Options: Export collections to JSON or CSV for sharing or migration
• Clean TUI: Navigable interface with list/detail views, editable forms, and status indicators
• Developer Friendly: Small codebase with clear package boundaries — ideal for contributors and experimentation

Perfect for developers who can't cook but can definitely write code. Features TUI, JSON export, and zero kitchen fires! 🔥

Cook boldly. Ship deliciousness.`,
	Example: `
# Run in interactive mode
yummy

# Print version
yummy -v

# Run with debug logging
yummy -d
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
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
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
	ctx := cmd.Context()

	cwd, err := ResolveCwd(cmd)
	if err != nil {
		return nil, err
	}

	//debug, _ := cmd.Flags().GetBool("debug")
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

	// check ollama status
	ollamaStatus := chat.GetOllamaServiceStatus()
	if len(ollamaStatus.Errors) > 0 {
		slog.Error("Ollama service status", "errors", ollamaStatus.Errors)
		return nil, fmt.Errorf("ollama service status: %v", ollamaStatus.Errors)
	}

	tuiInstance, err := tui.New(conn, ctx)
	if err != nil {
		slog.Error("Failed to create tui instance", "error", err)
		return nil, err
	}

	return tuiInstance, nil
}
