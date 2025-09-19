package styles

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFFFFF")).
	Background(lipgloss.Color("#FF6B6B")).
	Padding(0, 4)

var InfoStyle = func() lipgloss.Style {
	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	return lipgloss.NewStyle().
		BorderStyle(b).
		Foreground(lipgloss.Color("#87CEEB"))
}()

var ErrorStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF6B6B")).
	Padding(0, 2)

var HeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#98FB98")).
	PaddingTop(1).
	PaddingBottom(1)

var IngredientStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#DDA0DD")).
	PaddingLeft(2)

var DocStyle = lipgloss.NewStyle().
	Margin(1, 2)

var DetailContentStyle = lipgloss.NewStyle().
	Padding(1, 2).
	Margin(0, 1)

var DetailHeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFFFFF")).
	Background(lipgloss.Color("#FF6B6B")).
	Padding(0, 4).
	MarginBottom(0)

var DetailFooterStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#87CEEB")).
	Padding(0, 1).
	MarginTop(1)

var ScrollBarStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#87CEEB"))

var LoadingStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFB6B6")).
	Italic(true)

var InstructionStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#F0E68C")).
	PaddingLeft(2)

// Status line styles - Modern theme with better colors
var WarningStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#F0E68C")).
	Background(lipgloss.Color("#2a2a2a")).
	Padding(0, 1)

var SuccessStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#4ECDC4")).
	Background(lipgloss.Color("#2a2a2a")).
	Padding(0, 1)

var HelpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFB6B6")).
	Background(lipgloss.Color("#2a2a2a")).
	Padding(0, 1)

var StatusLineStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFB6B6")).
	Background(lipgloss.Color("#2a2a2a")).
	Padding(0, 1)

// Neovim/Vim-style status line components - Enhanced with theme colors
var StatusLineLeftStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFB6B6")).
	Background(lipgloss.Color("#1a1a1a")).
	Padding(0, 1)

var StatusLineRightStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFB6B6")).
	Background(lipgloss.Color("#1a1a1a")).
	Padding(0, 1)

var StatusLineModeStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#000000")).
	Background(lipgloss.Color("#FF6B6B")).
	Bold(true).
	Padding(0, 1)

var StatusLineFileStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#87CEEB")).
	Background(lipgloss.Color("#1a1a1a")).
	Padding(0, 1)

var StatusLineInfoStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#4ECDC4")).
	Background(lipgloss.Color("#1a1a1a")).
	Padding(0, 1)

var StatusLineSeparatorStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFB6B6")).
	Background(lipgloss.Color("#1a1a1a"))

var PunkyStyle = list.Styles{
	TitleBar:     lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Background(lipgloss.Color("black")).Bold(true).Padding(1, 2),
	Title:        lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Padding(1, 2),
	Spinner:      lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
	FilterPrompt: lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Padding(1, 2),
	FilterCursor: lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),

	DefaultFilterCharacterMatch: lipgloss.NewStyle().Foreground(lipgloss.Color("green")),
	StatusBar:                   lipgloss.NewStyle().Foreground(lipgloss.Color("white")).Background(lipgloss.Color("black")),
	StatusEmpty:                 lipgloss.NewStyle().Foreground(lipgloss.Color("red")),
	StatusBarActiveFilter:       lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")),
	StatusBarFilterCount:        lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")),
	NoItems:                     lipgloss.NewStyle().Foreground(lipgloss.Color("red")),
	PaginationStyle:             lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
	HelpStyle:                   lipgloss.NewStyle().Foreground(lipgloss.Color("white")).Background(lipgloss.Color("black")),
	ActivePaginationDot:         lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
	InactivePaginationDot:       lipgloss.NewStyle().Foreground(lipgloss.Color("black")),
	ArabicPagination:            lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
	DividerDot:                  lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
}
