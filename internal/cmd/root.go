package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/fang/v2"
	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	log "github.com/GarroshIcecream/yummy/internal/log"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	tui "github.com/GarroshIcecream/yummy/internal/tui"
	"github.com/GarroshIcecream/yummy/internal/tui/chat"
	"github.com/GarroshIcecream/yummy/internal/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")

	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
}

var rootCmd = &cobra.Command{
	Use:   "yummy",
	Short: "Terminal cookbook manager and recipe assistant",
	Long: `Yummy is a terminal-first cookbook manager with recipe import/export,
URL scraping, themes, and an Ollama-powered cooking assistant.`,
	Example: `
# Run in interactive mode
yummy

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

	// Set global config
	config.SetGlobalConfig(cfg)

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

	chatConfig := config.GetChatConfig()
	_, err = chat.GetOllamaServiceStatus(chatConfig.DefaultModel)
	if err != nil {
		slog.Error("Failed to get ollama service status", "error", err)
		return nil, fmt.Errorf("failed to get ollama service status: %v", err)
	}

	tuiInstance, err := tui.New(cookbook, sessionLog, themeManager, ctx)
	if err != nil {
		slog.Error("Failed to create tui instance", "error", err)
		return nil, fmt.Errorf("failed to create TUI instance: %v", err)
	}

	return tuiInstance, nil
}
