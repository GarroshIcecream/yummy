package styles

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().
			BorderStyle(b).
			Padding(0, 1).
			Foreground(lipgloss.Color("#FFA07A"))
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return lipgloss.NewStyle().
			BorderStyle(b).
			Foreground(lipgloss.Color("#87CEEB"))
	}()

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#98FB98")).
			PaddingTop(1).
			PaddingBottom(1)

	ingredientStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DDA0DD")).
			PaddingLeft(2)

	instructionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F0E68C")).
				PaddingLeft(2)
)
