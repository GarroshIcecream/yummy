// Package overlay provides functionality for compositing TUI views on top of each other.
// This is an internal port of github.com/rmhubbert/bubbletea-overlay, using
// charm.land/lipgloss/v2 for size measurement instead of lipgloss v1.
package overlay

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Position represents a relative offset in the TUI. There are five possible values:
// Top, Right, Bottom, Left, and Center.
type Position int

const (
	Top Position = iota + 1
	Right
	Bottom
	Left
	Center
)

// Viewable is implemented by any model that can render itself as a string.
type Viewable interface {
	View() string
}

// TeaViewable is implemented by models whose View() returns tea.View (bubbletea v2).
type TeaViewable interface {
	View() tea.View
}

// teaViewAdapter wraps a TeaViewable to satisfy Viewable.
type teaViewAdapter struct{ m TeaViewable }

func (a teaViewAdapter) View() string { return a.m.View().Content }

// WrapTeaViewable wraps a TeaViewable so it can be used as Viewable.
func WrapTeaViewable(m TeaViewable) Viewable { return teaViewAdapter{m} }

// Model manages calculating and compositing the overlay UI from the background and foreground models.
type Model struct {
	Foreground Viewable
	Background Viewable
	XPosition  Position
	YPosition  Position
	XOffset    int
	YOffset    int
}

// New creates and returns a pointer to a new overlay Model.
func New(fore Viewable, back Viewable, xPos Position, yPos Position, xOff int, yOff int) *Model {
	return &Model{
		Foreground: fore,
		Background: back,
		XPosition:  xPos,
		YPosition:  yPos,
		XOffset:    xOff,
		YOffset:    yOff,
	}
}

// View applies the compositing and renders the overlaid view.
func (m *Model) View() string {
	if m.Foreground == nil && m.Background == nil {
		return ""
	}
	if m.Foreground == nil && m.Background != nil {
		return m.Background.View()
	}
	if m.Foreground != nil && m.Background == nil {
		return m.Foreground.View()
	}

	return Composite(
		m.Foreground.View(),
		m.Background.View(),
		m.XPosition,
		m.YPosition,
		m.XOffset,
		m.YOffset,
	)
}

// Composite merges and flattens the background and foreground views into a single view.
// Based on the implementation used by Superfile:
// https://github.com/yorukot/superfile/blob/main/src/pkg/string_function/overplace.go
func Composite(fg, bg string, xPos, yPos Position, xOff, yOff int) string {
	if fg == "" {
		return bg
	}
	if bg == "" {
		return fg
	}
	if strings.Count(fg, "\n") == 0 && strings.Count(bg, "\n") == 0 {
		return fg
	}

	fgWidth, fgHeight := lipgloss.Size(fg)
	bgWidth, bgHeight := lipgloss.Size(bg)

	if fgWidth >= bgWidth && fgHeight >= bgHeight {
		return fg
	}

	x, y := offsets(fg, bg, xPos, yPos, xOff, yOff)
	x = clamp(x, 0, bgWidth-fgWidth)
	y = clamp(y, 0, bgHeight-fgHeight)

	fgLines := splitLines(fg)
	bgLines := splitLines(bg)
	var sb strings.Builder

	for i, bgLine := range bgLines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		if i < y || i >= y+fgHeight {
			sb.WriteString(bgLine)
			continue
		}

		pos := 0
		if x > 0 {
			left := ansi.Truncate(bgLine, x, "")
			pos = ansi.StringWidth(left)
			sb.WriteString(left)
			if pos < x {
				sb.WriteString(whitespace(x - pos))
				pos = x
			}
		}

		fgLine := fgLines[i-y]
		sb.WriteString(fgLine)
		pos += ansi.StringWidth(fgLine)

		right := ansi.TruncateLeft(bgLine, pos, "")
		bgLineWidth := ansi.StringWidth(bgLine)
		rightWidth := ansi.StringWidth(right)
		if rightWidth <= bgLineWidth-pos {
			sb.WriteString(whitespace(bgLineWidth - rightWidth - pos))
		}
		sb.WriteString(right)
	}
	return sb.String()
}

// offsets calculates the actual vertical and horizontal offsets used to position the foreground
// view relative to the background view.
func offsets(fg, bg string, xPos, yPos Position, xOff, yOff int) (int, int) {
	var x, y int

	switch xPos {
	case Left:
		x = 0
	case Center:
		x = lipgloss.Width(bg)/2 - lipgloss.Width(fg)/2
	case Right:
		x = lipgloss.Width(bg) - lipgloss.Width(fg)
	}

	switch yPos {
	case Top:
		y = 0
	case Center:
		y = lipgloss.Height(bg)/2 - lipgloss.Height(fg)/2
	case Bottom:
		y = lipgloss.Height(bg) - lipgloss.Height(fg)
	}

	return x + xOff, y + yOff
}

// clamp clamps a value between lower and upper bounds.
func clamp(v, lower, upper int) int {
	if lower > upper {
		lower, upper = upper, lower
	}
	if v < lower {
		return lower
	}
	if v > upper {
		return upper
	}
	return v
}

// splitLines normalises non-standard newlines and splits a string into lines.
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(s, "\n")
}

// whitespace returns a string of spaces of the requested length.
func whitespace(length int) string {
	return strings.Repeat(" ", length)
}
