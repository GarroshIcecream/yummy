package chat

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

// renderTitledPanel wraps body in a rounded-border panel whose top border
// embeds a left-aligned title and an optional right-aligned status. lipgloss
// v2 has no native border-title support, so the top edge is composed by hand
// and the remaining three sides are drawn by a borderless-top style.
//
// title and status are already-styled strings (the caller controls their
// colors and background so each panel can sit on its own surface). contentWidth
// is the width of the body's text area (excluding the panel's border and
// horizontal padding). The returned panel is contentWidth+4 columns wide: 1
// border + 1 pad on each side.
func renderTitledPanel(theme *themes.Theme, body, title, status string, contentWidth int, bg color.Color) string {
	const padX = 1
	border := lipgloss.RoundedBorder()
	bc := lipgloss.NewStyle().Foreground(theme.ChatPanelBorder).Background(bg)

	// The top edge spans the full panel width. Between the two corners sit
	// (padX + contentWidth + padX) columns — the same span the bordered body
	// below occupies.
	inner := contentWidth + 2*padX
	top := panelTopEdge(bc, border, title, status, inner)

	// lipgloss Width sets the total box width (border + padding included), so
	// add the two border columns and the horizontal padding to leave exactly
	// contentWidth columns for the body text.
	bodyStyle := lipgloss.NewStyle().
		Border(border).
		BorderTop(false).
		BorderForeground(theme.ChatPanelBorder).
		Background(bg).
		Padding(0, padX).
		Width(contentWidth + 2*padX + 2)

	return top + "\n" + bodyStyle.Render(body)
}

// panelTopEdge builds the titled top border line. inner is the number of
// columns between (and not including) the two rounded corners. title and status
// are already styled; everything else is painted in the border color.
func panelTopEdge(bc lipgloss.Style, border lipgloss.Border, title, status string, inner int) string {
	// No title or status: a plain rounded top edge.
	if title == "" && status == "" {
		return bc.Render(border.TopLeft + strings.Repeat(border.Top, inner) + border.TopRight)
	}

	// Fixed decoration around the title: "<corner>─ <title> " on the left.
	titleW := lipgloss.Width(title)
	leftUsed := 2 + titleW + 1 // "─ " + title + " "

	rightUsed := 1 // trailing "─" before the corner
	if status != "" {
		// " " + status + " " + trailing "─"
		rightUsed = lipgloss.Width(status) + 3
	}

	// If the title and status can't both fit, drop the status rather than
	// overflow the frame and ragged the top border.
	if status != "" && inner-leftUsed-rightUsed < 1 {
		status = ""
		rightUsed = 1
	}

	fill := max(inner-leftUsed-rightUsed, 0)

	var sb strings.Builder
	sb.WriteString(bc.Render(border.TopLeft + border.Top + " "))
	sb.WriteString(title)
	sb.WriteString(bc.Render(" " + strings.Repeat(border.Top, fill)))
	if status != "" {
		sb.WriteString(bc.Render(" "))
		sb.WriteString(status)
		sb.WriteString(bc.Render(" "))
	}
	sb.WriteString(bc.Render(border.Top + border.TopRight))
	return sb.String()
}
