package styles

import "github.com/charmbracelet/lipgloss"

// State Selector Dialog Styles - Following the app's design system
var (
	// Container and layout styles
	StateSelectorContainerStyle = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Width(50).
		Height(15).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#9370DB")).
		Background(lipgloss.Color("#1A0B2E")).
		Padding(2, 3)

	StateSelectorDialogStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4a9eff")).
		Padding(2, 3)
		
	// Title styling
	StateSelectorTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4a9eff")).
		MarginBottom(2).
		Align(lipgloss.Center)
	
	StateSelectorHelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Align(lipgloss.Center).
		MarginTop(2)

	// State item styles
	StateSelectorItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDA0DD")).
		Padding(0, 2).
		Margin(0, 0, 1, 0)

	StateSelectorSelectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#9370DB")).
		Bold(true).
		Padding(0, 2).
		Margin(0, 0, 1, 0)

	// Indicator styles
	StateSelectorIndicatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)

	StateSelectorSelectedIndicatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
)

// GetStateItemStyle returns the appropriate style for a state item
func GetStateItemStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return StateSelectorSelectedItemStyle
	}
	return StateSelectorItemStyle
}

// GetIndicatorStyle returns the appropriate style for the indicator
func GetIndicatorStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return StateSelectorSelectedIndicatorStyle
	}
	return StateSelectorIndicatorStyle
}
