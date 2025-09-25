package styles

import "github.com/charmbracelet/lipgloss"

var (
	// TitleStyle defines the styling for the application title
	ChatTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff6b9d")).
		Bold(true).
		Margin(1, 0).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ff6b9d")).
		Padding(0, 2).
		Align(lipgloss.Center).
		Background(lipgloss.Color("#2a2a2a"))

	// ChatStyle defines the styling for the chat viewport
	ChatStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Padding(1, 2).
		Margin(0, 1, 0, 0).
		Foreground(lipgloss.Color("#ffffff"))

	// Sidebar styles - Modern dark theme
	SidebarStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Padding(2, 2, 2, 2).
		Margin(1, 1, 0, 0).
		Width(28).
		Background(lipgloss.Color("#1a1a1a")).
		Foreground(lipgloss.Color("#ffffff"))

	SidebarHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff6b9d")).
		Bold(true).
		Align(lipgloss.Center).
		Margin(0, 0, 2, 0).
		Background(lipgloss.Color("#2a2a2a")).
		Padding(1, 2)

	SidebarSectionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true).
		Margin(0, 0, 1, 0).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false).
		Padding(0, 0, 1, 0)

	SidebarContentStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc")).
		Margin(0, 0, 0, 2)

	SidebarSuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Bold(true)

	SidebarErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff6b6b")).
		Bold(true)

	// User message styles - Clean text without backgrounds
	UserMessageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff6b9d")).
		Bold(true).
		Align(lipgloss.Left).
		Margin(0, 0, 0, 0)

	UserContentStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Align(lipgloss.Left).
		Margin(0, 0, 1, 0).
		Width(0) // Allow dynamic width

	// Assistant message styles - Clean text without backgrounds
	AssistantMessageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Bold(true).
		Align(lipgloss.Left).
		Margin(0, 0, 0, 0)

	AssistantContentStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Align(lipgloss.Left).
		Margin(0, 0, 1, 0).
		Width(0) // Allow dynamic width

	// Legacy styles for backward compatibility
	UserStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Margin(0, 0, 1, 0).
		Width(0) // Allow dynamic width based on container

	AssistantStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Margin(0, 0, 1, 0).
		Width(0) // Allow dynamic width based on container

	SpinnerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Bold(true)
)
