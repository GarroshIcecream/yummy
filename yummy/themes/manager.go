package themes

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/GarroshIcecream/yummy/yummy/config"
)

// ThemeManager handles theme operations
type ThemeManager struct {
	currentTheme *Theme
	themes       []Theme
	themesDir    string
}

// NewThemeManagerWithDir creates a new theme manager with a specific themes directory
func NewThemeManager(themesDir string) (*ThemeManager, error) {
	config := config.GetGlobalConfig()

	// Create themes directory if it doesn't exist
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		if err := os.MkdirAll(themesDir, 0755); err != nil {
			slog.Error("Failed to create themes directory", "dir", themesDir, "error", err)
			return nil, err
		}
	}

	// Start with default theme
	themes := []Theme{NewDefaultTheme()}

	// Load custom themes from directory
	customThemes, err := LoadThemesFromDirectory(themesDir)
	if err != nil {
		slog.Error("Failed to load custom themes", "error", err)
	} else {
		themes = append(themes, customThemes...)
	}

	// Find current theme
	currentTheme := themes[0]
	for _, theme := range themes {
		if theme.Name == config.Theme {
			currentTheme = theme
			break
		}
	}

	return &ThemeManager{
		themes:       themes,
		themesDir:    themesDir,
		currentTheme: &currentTheme,
	}, nil
}

// RegisterTheme registers a new theme
func (tm *ThemeManager) RegisterTheme(theme Theme) {
	tm.themes = append(tm.themes, theme)
}

// SetTheme sets the current theme
func (tm *ThemeManager) SetThemeByName(name string) error {
	for _, theme := range tm.themes {
		if theme.Name == name {
			tm.currentTheme = &theme
			return nil
		}
	}
	return fmt.Errorf("theme %s not found", name)
}

// GetCurrentTheme returns the current theme
func (tm *ThemeManager) GetCurrentTheme() *Theme {
	return tm.currentTheme
}

// GetAvailableThemes returns a list of available theme names
func (tm *ThemeManager) GetAvailableThemes() []string {
	themes := make([]string, 0, len(tm.themes))
	for _, theme := range tm.themes {
		themes = append(themes, theme.Name)
	}
	return themes
}

// ReloadThemes reloads themes from the themes directory
func (tm *ThemeManager) ReloadThemes() error {
	if tm.themesDir == "" {
		return nil
	}

	customThemes, err := LoadThemesFromDirectory(tm.themesDir)
	if err != nil {
		return fmt.Errorf("failed to reload themes: %v", err)
	}

	themes := make([]Theme, 0)
	themes = append(themes, NewDefaultTheme())
	themes = append(themes, customThemes...)
	tm.themes = themes

	return nil
}

// GetThemesDirectory returns the themes directory path
func (tm *ThemeManager) GetThemesDirectory() string {
	return tm.themesDir
}
