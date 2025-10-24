package themes

import "fmt"

// ThemeManager handles theme operations
type ThemeManager struct {
	currentTheme *Theme
	themes       []*Theme
}

// NewThemeManager creates a new theme manager
func NewThemeManager() *ThemeManager {
	themes := make([]*Theme, 0)
	themes = append(themes, NewDefaultTheme())
	themes = append(themes, NewDarkTheme())
	themes = append(themes, NewLightTheme())

	return &ThemeManager{
		themes:       themes,
		currentTheme: themes[0],
	}
}

// RegisterTheme registers a new theme
func (tm *ThemeManager) RegisterTheme(theme *Theme) {
	tm.themes = append(tm.themes, theme)
}

// SetTheme sets the current theme
func (tm *ThemeManager) SetThemeByName(name string) error {
	for _, theme := range tm.themes {
		if theme.Name == name {
			tm.currentTheme = theme
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
