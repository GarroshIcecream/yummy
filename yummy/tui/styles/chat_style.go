package styles

import "github.com/charmbracelet/lipgloss"

var (
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	// TitleStyle defines the styling for the application title
	ChatTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true).
			Margin(1, 0).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6B6B")).
			Padding(0, 2).
			Align(lipgloss.Center).
			MaxWidth(80)

	// ChatStyle defines the styling for the chat viewport
	ChatStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.ASCIIBorder()).
			BorderForeground(highlight).
			Padding(1, 2)

	// UserStyle defines the styling for user messages
	UserStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6B6B")).
			Margin(0, 0, 1, 0).
			Width(0) // Allow dynamic width based on container

	// AssistantStyle defines the styling for assistant messages
	AssistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4ECDC4")).
			Margin(0, 0, 1, 0).
			Width(0) // Allow dynamic width based on container

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4"))
)
