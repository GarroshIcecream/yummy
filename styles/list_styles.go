package styles

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func ApplyDelegateStyles(d list.DefaultDelegate) list.DefaultDelegate {
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2)

	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Padding(0, 2)

	selectedColor := lipgloss.Color("#FF6B6B")
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(selectedColor).
		BorderLeftForeground(selectedColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		Padding(0, 2)

	d.Styles.SelectedDesc = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFB6B6"))

	return d
}

func ApplyListStyles(l list.Model) list.Model {

	// Make the title pop with a subtle glow effect
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	// Style pagination with dots
	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		PaddingLeft(2)

	// Style the help text to be more subtle
	l.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Padding(1, 0, 0, 2)

	// Make the filter prompt stand out
	l.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)

	// Style the filter cursor
	l.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF6B6B"))

	// Style the "no items" message
	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Italic(true)

	// Style the pagination dots
	l.Styles.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		SetString("●")

	l.Styles.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		SetString("○")

	l.Styles.DividerDot = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		SetString(" • ")

	// Style the status bar
	l.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Padding(0, 0, 1, 2)

	// Style the status bar when filtering
	l.Styles.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)

	// Style the filter count
	l.Styles.StatusBarFilterCount = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6"))

	return l
}
