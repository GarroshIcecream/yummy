package themes

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// NewDarkTheme creates a new dark theme with a sleek, modern look
func NewDarkTheme() *Theme {
	t := &Theme{
		Name: "dark",
	}

	// Core styles - Dark theme with subtle accents
	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#2D3748")).
		Padding(0, 4)

	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	t.Info = lipgloss.NewStyle().
		BorderStyle(b).
		Foreground(lipgloss.Color("#68D391"))

	t.Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F56565")).
		Padding(0, 2)

	t.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#81E6D9")).
		PaddingTop(1).
		PaddingBottom(1)

	t.Ingredient = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D6BCFA")).
		PaddingLeft(2)

	t.Doc = lipgloss.NewStyle().
		Margin(1, 2)

	t.DetailContent = lipgloss.NewStyle().
		Padding(1, 2).
		Margin(0, 1)

	t.DetailHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#2D3748")).
		Padding(0, 4).
		MarginBottom(0)

	t.DetailFooter = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#68D391")).
		Padding(0, 1).
		MarginTop(1)

	t.ScrollBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#68D391"))

	t.Loading = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F6AD55")).
		Italic(true)

	t.Instruction = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F6E05E")).
		PaddingLeft(2)

	// Status line styles
	t.Warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F6E05E")).
		Background(lipgloss.Color("#1A202C")).
		Padding(0, 1)

	t.Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#68D391")).
		Background(lipgloss.Color("#1A202C")).
		Padding(0, 1)

	t.Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Background(lipgloss.Color("#1A202C")).
		Padding(0, 1)

	t.StatusLine = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Background(lipgloss.Color("#1A202C")).
		Padding(0, 1)

	t.StatusLineLeft = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Background(lipgloss.Color("#171923")).
		Padding(0, 1)

	t.StatusLineRight = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Background(lipgloss.Color("#171923")).
		Padding(0, 1)

	t.StatusLineMode = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#2D3748")).
		Bold(true).
		Padding(0, 1)

	t.StatusLineFile = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#68D391")).
		Background(lipgloss.Color("#171923")).
		Padding(0, 1)

	t.StatusLineInfo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#81E6D9")).
		Background(lipgloss.Color("#171923")).
		Padding(0, 1)

	t.StatusLineSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Background(lipgloss.Color("#171923"))

	// List styles
	t.ListStyles = list.Styles{
		TitleBar:     lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")).Background(lipgloss.Color("#1A202C")).Bold(true).Padding(1, 2),
		Title:        lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")).Bold(true).Padding(1, 2),
		Spinner:      lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")),
		FilterPrompt: lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")).Padding(1, 2),
		FilterCursor: lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")).Bold(true),

		DefaultFilterCharacterMatch: lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")),
		StatusBar:                   lipgloss.NewStyle().Foreground(lipgloss.Color("#A0AEC0")).Background(lipgloss.Color("#1A202C")),
		StatusEmpty:                 lipgloss.NewStyle().Foreground(lipgloss.Color("#F56565")),
		StatusBarActiveFilter:       lipgloss.NewStyle().Foreground(lipgloss.Color("#F6E05E")),
		StatusBarFilterCount:        lipgloss.NewStyle().Foreground(lipgloss.Color("#F6E05E")),
		NoItems:                     lipgloss.NewStyle().Foreground(lipgloss.Color("#F56565")),
		PaginationStyle:             lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")),
		HelpStyle:                   lipgloss.NewStyle().Foreground(lipgloss.Color("#A0AEC0")).Background(lipgloss.Color("#1A202C")),
		ActivePaginationDot:         lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")),
		InactivePaginationDot:       lipgloss.NewStyle().Foreground(lipgloss.Color("#4A5568")),
		ArabicPagination:            lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")),
		DividerDot:                  lipgloss.NewStyle().Foreground(lipgloss.Color("#68D391")),
	}

	// Delegate styles for lists
	normalTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0")).
		Padding(0, 2)

	normalDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Padding(0, 2)

	selectedColor := lipgloss.Color("#68D391")
	selectedTitle := lipgloss.NewStyle().
		Foreground(selectedColor).
		BorderLeftForeground(selectedColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		Padding(0, 2)

	selectedDesc := selectedTitle.
		Foreground(lipgloss.Color("#81E6D9"))

	dimmedTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Padding(0, 2)

	dimmedDesc := dimmedTitle.
		Foreground(lipgloss.Color("#A0AEC0"))

	filterMatch := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#68D391")).
		Underline(true)

	t.DelegateStyles = list.DefaultItemStyles{
		NormalTitle:   normalTitle,
		NormalDesc:    normalDesc,
		SelectedTitle: selectedTitle,
		SelectedDesc:  selectedDesc,
		DimmedTitle:   dimmedTitle,
		DimmedDesc:    dimmedDesc,
		FilterMatch:   filterMatch,
	}

	// Chat styles
	t.ChatTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D6BCFA")).
		Bold(true).
		Margin(1, 0).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#D6BCFA")).
		Padding(0, 2).
		Align(lipgloss.Center).
		Background(lipgloss.Color("#1A202C"))

	t.Chat = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4A5568")).
		Padding(1, 2).
		Margin(0, 1, 0, 0).
		Foreground(lipgloss.Color("#E2E8F0"))

	t.Sidebar = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#4A5568")).
		Padding(2, 2, 2, 2).
		Margin(1, 1, 0, 0).
		Width(28).
		Background(lipgloss.Color("#171923")).
		Foreground(lipgloss.Color("#E2E8F0"))

	t.SidebarHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D6BCFA")).
		Bold(true).
		Align(lipgloss.Center).
		Margin(0, 0, 2, 0).
		Background(lipgloss.Color("#1A202C")).
		Padding(1, 2)

	t.SidebarSection = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0")).
		Bold(true).
		Margin(0, 0, 1, 0).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#4A5568")).
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false).
		Padding(0, 0, 1, 0)

	t.SidebarContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Margin(0, 0, 0, 2)

	t.SidebarSuccess = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#68D391")).
		Bold(true)

	t.SidebarError = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F56565")).
		Bold(true)

	t.UserMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D6BCFA")).
		Bold(true).
		Align(lipgloss.Left).
		Margin(0, 0, 0, 0)

	t.UserContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0")).
		Align(lipgloss.Left).
		Margin(0, 0, 1, 0).
		Width(0)

	t.AssistantMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#81E6D9")).
		Bold(true).
		Align(lipgloss.Left).
		Margin(0, 0, 0, 0)

	t.AssistantContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0")).
		Align(lipgloss.Left).
		Margin(0, 0, 1, 0).
		Width(0)

	t.User = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F56565")).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#F56565")).
		Margin(0, 0, 1, 0).
		Width(0)

	t.Assistant = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#81E6D9")).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#81E6D9")).
		Margin(0, 0, 1, 0).
		Width(0)

	t.Spinner = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#81E6D9")).
		Bold(true)

	// Main menu styles
	t.MainMenuBorder = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#68D391"))

	t.MainMenuContainer = lipgloss.NewStyle().
		Background(lipgloss.Color("#171923")).
		Padding(1, 2)

	t.MainMenuSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#68D391"))

	t.MainMenuWelcome = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Italic(true).
		Padding(1, 0)

	t.MainMenuLogo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D6BCFA")).
		Bold(true)

	t.MainMenuSubtitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Italic(true).
		Padding(1, 0)

	t.MainMenuTitleBorder = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#68D391")).
		Align(lipgloss.Center).
		Padding(1, 2).
		Margin(1, 0)

	t.MainMenuSelectedArrow = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F6E05E")).
		Bold(true)

	t.MainMenuSelectedItem = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#F6E05E"))

	t.MainMenuUnselectedItem = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#68D391"))

	t.MainMenuSelectedIcon = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F6E05E")).
		Bold(true)

	t.MainMenuUnselectedIcon = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0"))

	t.MainMenuSelectedTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F6E05E"))

	t.MainMenuUnselectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E2E8F0"))

	t.MainMenuSelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F6AD55")).
		Italic(true).
		PaddingLeft(4).
		PaddingBottom(1)

	t.MainMenuUnselectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Italic(true).
		PaddingLeft(4).
		PaddingBottom(1)

	t.MainMenuHelpHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#68D391")).
		Bold(true).
		PaddingBottom(1)

	t.MainMenuHelpContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		PaddingLeft(2)

	t.MainMenuHelpBorder = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#68D391")).
		Background(lipgloss.Color("#1A202C")).
		Padding(1, 2).
		Margin(1, 0)

	t.MainMenuSpinner = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#68D391"))

	// State selector styles
	t.StateSelectorContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Width(50).
		Height(15).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#68D391")).
		Background(lipgloss.Color("#171923")).
		Padding(2, 3)

	t.StateSelectorDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#81E6D9")).
		Padding(2, 3)

	t.StateSelectorTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#81E6D9")).
		MarginBottom(2).
		Align(lipgloss.Center)

	t.StateSelectorHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#718096")).
		Italic(true).
		Align(lipgloss.Center).
		MarginTop(2)

	t.StateSelectorItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0AEC0")).
		Padding(0, 2).
		Margin(0, 0, 1, 0)

	t.StateSelectorSelectedItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#68D391")).
		Bold(true).
		Padding(0, 2).
		Margin(0, 0, 1, 0)

	t.StateSelectorIndicator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F6E05E")).
		Bold(true)

	t.StateSelectorSelectedIndicator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	return t
}
