package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	log "github.com/GarroshIcecream/yummy/yummy/log"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	tui "github.com/GarroshIcecream/yummy/yummy/tui"
	"github.com/GarroshIcecream/yummy/yummy/tui/chat"
	"github.com/GarroshIcecream/yummy/yummy/version"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")
	rootCmd.Flags().StringP("theme", "t", "default", "Theme to use (default, dark, light)")

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

# Run with debug logging
yummy -d

# Run with a theme
yummy -t dark
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

func resolveUserDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}

	datadir := filepath.Join(homeDir, ".yummy")
	if err := os.MkdirAll(datadir, 0755); err != nil {
		return "", fmt.Errorf("failed to create Yummy data directory: %v", err)
	}

	return datadir, nil
}

func setupApp(cmd *cobra.Command) (*tui.Manager, error) {
	ctx := cmd.Context()

	// Resolve user directory for data storage
	datadir, err := resolveUserDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user directory: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig(datadir)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	// Override config with command line flags if provided
	theme, err := cmd.Flags().GetString("theme")
	if err != nil {
		return nil, err
	} else {
		cfg.Theme = theme
	}

	// Setup logging first before any other operations
	debug, _ := cmd.Flags().GetBool("debug")
	log.Setup(datadir, debug)

	themesDir := filepath.Join(datadir, "themes")
	themeManager, err := themes.NewThemeManager(themesDir)
	if err != nil {
		slog.Error("Failed to create theme manager", "error", err)
		return nil, fmt.Errorf("failed to create theme manager: %v", err)
	}

	if err := themeManager.SetThemeByName(cfg.Theme); err != nil {
		slog.Error("Failed to set theme", "theme", cfg.Theme, "error", err)
		return nil, fmt.Errorf("failed to set theme '%s': %v", cfg.Theme, err)
	}

	cookbook, err := db.NewCookBook(datadir, &cfg.Database)
	if err != nil {
		slog.Error("Failed to initialize cookbook", "error", err)
		return nil, fmt.Errorf("failed to initialize cookbook: %v", err)
	}

	sessionLog, err := db.NewSessionLog(datadir, &cfg.Database)
	if err != nil {
		slog.Error("Failed to initialize session log", "error", err)
		return nil, fmt.Errorf("failed to initialize session log: %v", err)
	}

	_, err = chat.GetOllamaServiceStatus(cfg.Chat.DefaultModel)
	if err != nil {
		slog.Error("Failed to get ollama service status", "error", err)
		return nil, fmt.Errorf("failed to get ollama service status: %v", err)
	}

	tuiInstance, err := tui.New(cookbook, sessionLog, themeManager, cfg, ctx)
	if err != nil {
		slog.Error("Failed to create tui instance", "error", err)
		return nil, fmt.Errorf("failed to create TUI instance: %v", err)
	}

	return tuiInstance, nil
}
