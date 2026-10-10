package themes

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"sort"

	"github.com/GarroshIcecream/yummy/internal/config"
)

//go:embed builtin/*.yaml
var builtinThemeFS embed.FS

// BuiltinThemes returns the default theme followed by the themes shipped with
// the binary (internal/themes/builtin), sorted by name.
func BuiltinThemes() []Theme {
	themes := []Theme{NewDefaultTheme()}

	entries, err := fs.ReadDir(builtinThemeFS, "builtin")
	if err != nil {
		slog.Error("Failed to read built-in themes", "error", err)
		return themes
	}
	var loaded []Theme
	for _, entry := range entries {
		data, err := builtinThemeFS.ReadFile("builtin/" + entry.Name())
		if err != nil {
			slog.Error("Failed to read built-in theme", "name", entry.Name(), "error", err)
			continue
		}
		theme, err := parseYAMLTheme(data)
		if err != nil {
			slog.Error("Failed to parse built-in theme", "name", entry.Name(), "error", err)
			continue
		}
		loaded = append(loaded, *theme)
	}
	sort.Slice(loaded, func(i, j int) bool { return loaded[i].Name < loaded[j].Name })
	return append(themes, loaded...)
}

// mergeThemes layers custom themes over the built-in ones: a custom theme with
// the same name as a built-in replaces it in place, others are appended.
func mergeThemes(builtin, custom []Theme) []Theme {
	themes := append([]Theme(nil), builtin...)
	for _, c := range custom {
		replaced := false
		for i := range themes {
			if themes[i].Name == c.Name {
				themes[i] = c
				replaced = true
				break
			}
		}
		if !replaced {
			themes = append(themes, c)
		}
	}
	return themes
}

// ThemeManager handles theme operations
type ThemeManager struct {
	currentThemeName string
	themes           []Theme
	themesDir        string
}

// ThemeInfo represents theme metadata for selection UIs.
type ThemeInfo struct {
	Name        string
	Description string
}

// NewThemeManager creates a new theme manager with a specific themes directory
func NewThemeManager(themesDir string) (*ThemeManager, error) {
	config := config.GetGlobalConfig()

	// Create themes directory if it doesn't exist
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		if err := os.MkdirAll(themesDir, 0o755); err != nil {
			slog.Error("Failed to create themes directory", "dir", themesDir, "error", err)
			return nil, err
		}
	}

	// Built-in themes, overridden/extended by custom themes from the directory
	customThemes, err := LoadThemesFromDirectory(themesDir)
	if err != nil {
		slog.Error("Failed to load custom themes", "error", err)
	}
	themes := mergeThemes(BuiltinThemes(), customThemes)

	// Find current theme
	currentThemeName := themes[0].Name
	for _, theme := range themes {
		if theme.Name == config.Theme {
			currentThemeName = theme.Name
			break
		}
	}

	return &ThemeManager{
		themes:           themes,
		themesDir:        themesDir,
		currentThemeName: currentThemeName,
	}, nil
}

// RegisterTheme registers a new theme
func (tm *ThemeManager) RegisterTheme(theme Theme) {
	tm.themes = append(tm.themes, theme)
}

// SetTheme sets the current theme
func (tm *ThemeManager) SetThemeByName(name string) error {
	for i := range tm.themes {
		if tm.themes[i].Name == name {
			tm.currentThemeName = name
			return nil
		}
	}
	return fmt.Errorf("theme %s not found", name)
}

// GetCurrentTheme returns the current theme
func (tm *ThemeManager) GetCurrentTheme() *Theme {
	for i := range tm.themes {
		if tm.themes[i].Name == tm.currentThemeName {
			return &tm.themes[i]
		}
	}
	if len(tm.themes) == 0 {
		return nil
	}
	return &tm.themes[0]
}

// GetAvailableThemes returns a list of available theme names
func (tm *ThemeManager) GetAvailableThemes() []string {
	themes := make([]string, 0, len(tm.themes))
	for _, theme := range tm.themes {
		themes = append(themes, theme.Name)
	}
	return themes
}

// GetAvailableThemeInfos returns available themes with metadata for selection UIs.
func (tm *ThemeManager) GetAvailableThemeInfos() []ThemeInfo {
	themes := make([]ThemeInfo, 0, len(tm.themes))
	for _, theme := range tm.themes {
		themes = append(themes, ThemeInfo{
			Name:        theme.Name,
			Description: theme.Description,
		})
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

	tm.themes = mergeThemes(BuiltinThemes(), customThemes)
	if err := tm.SetThemeByName(tm.currentThemeName); err != nil && len(tm.themes) > 0 {
		tm.currentThemeName = tm.themes[0].Name
	}

	return nil
}

// GetThemesDirectory returns the themes directory path
func (tm *ThemeManager) GetThemesDirectory() string {
	return tm.themesDir
}
