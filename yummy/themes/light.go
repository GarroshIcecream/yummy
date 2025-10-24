package themes

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// NewLightTheme creates a new light theme with a clean, bright look
func NewLightTheme() *Theme {
	t := &Theme{
		Name: "light",
	}

	// Core styles - Light theme with clean colors
	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#2563EB")). // Blue background
		Padding(0, 4)

	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	t.Info = lipgloss.NewStyle().
		BorderStyle(b).
		Foreground(lipgloss.Color("#059669")) // Green

	t.Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DC2626")). // Red
		Padding(0, 2)

	t.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#059669")). // Green
		PaddingTop(1).
		PaddingBottom(1)

	t.Ingredient = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")). // Purple
		PaddingLeft(2)

	t.Doc = lipgloss.NewStyle().
		Margin(1, 2)

	t.DetailContent = lipgloss.NewStyle().
		Padding(1, 2).
		Margin(0, 1)

	t.DetailHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#2563EB")).
		Padding(0, 4).
		MarginBottom(0)

	t.DetailFooter = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669")).
		Padding(0, 1).
		MarginTop(1)

	t.ScrollBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669"))

	t.Loading = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D97706")). // Orange
		Italic(true)

	t.Instruction = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D97706")). // Orange
		PaddingLeft(2)

	// Status line styles
	t.Warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D97706")).
		Background(lipgloss.Color("#F3F4F6")). // Light gray
		Padding(0, 1)

	t.Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669")).
		Background(lipgloss.Color("#F3F4F6")).
		Padding(0, 1)

	t.Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")). // Gray
		Background(lipgloss.Color("#F3F4F6")).
		Padding(0, 1)

	t.StatusLine = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Background(lipgloss.Color("#F3F4F6")).
		Padding(0, 1)

	t.StatusLineLeft = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Background(lipgloss.Color("#E5E7EB")). // Slightly darker gray
		Padding(0, 1)

	t.StatusLineRight = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Background(lipgloss.Color("#E5E7EB")).
		Padding(0, 1)

	t.StatusLineMode = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#2563EB")).
		Bold(true).
		Padding(0, 1)

	t.StatusLineFile = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669")).
		Background(lipgloss.Color("#E5E7EB")).
		Padding(0, 1)

	t.StatusLineInfo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Background(lipgloss.Color("#E5E7EB")).
		Padding(0, 1)

	t.StatusLineSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Background(lipgloss.Color("#E5E7EB"))

	// List styles
	t.ListStyles = list.Styles{
		TitleBar:     lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")).Background(lipgloss.Color("#F3F4F6")).Bold(true).Padding(1, 2),
		Title:        lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")).Bold(true).Padding(1, 2),
		Spinner:      lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")),
		FilterPrompt: lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")).Padding(1, 2),
		FilterCursor: lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")).Bold(true),

		DefaultFilterCharacterMatch: lipgloss.NewStyle().Foreground(lipgloss.Color("#059669")),
		StatusBar:                   lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Background(lipgloss.Color("#F3F4F6")),
		StatusEmpty:                 lipgloss.NewStyle().Foreground(lipgloss.Color("#DC2626")),
		StatusBarActiveFilter:       lipgloss.NewStyle().Foreground(lipgloss.Color("#D97706")),
		StatusBarFilterCount:        lipgloss.NewStyle().Foreground(lipgloss.Color("#D97706")),
		NoItems:                     lipgloss.NewStyle().Foreground(lipgloss.Color("#DC2626")),
		PaginationStyle:             lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")),
		HelpStyle:                   lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Background(lipgloss.Color("#F3F4F6")),
		ActivePaginationDot:         lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")),
		InactivePaginationDot:       lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")),
		ArabicPagination:            lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")),
		DividerDot:                  lipgloss.NewStyle().Foreground(lipgloss.Color("#2563EB")),
	}

	// Delegate styles for lists
	normalTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1F2937")).
		Padding(0, 2)

	normalDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Padding(0, 2)

	selectedColor := lipgloss.Color("#2563EB")
	selectedTitle := lipgloss.NewStyle().
		Foreground(selectedColor).
		BorderLeftForeground(selectedColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		Padding(0, 2)

	selectedDesc := selectedTitle.
		Foreground(lipgloss.Color("#059669"))

	dimmedTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Padding(0, 2)

	dimmedDesc := dimmedTitle.
		Foreground(lipgloss.Color("#6B7280"))

	filterMatch := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2563EB")).
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
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Margin(1, 0).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(0, 2).
		Align(lipgloss.Center).
		Background(lipgloss.Color("#F3F4F6"))

	t.Chat = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#D1D5DB")).
		Padding(1, 2).
		Margin(0, 1, 0, 0).
		Foreground(lipgloss.Color("#1F2937"))

	t.Sidebar = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#D1D5DB")).
		Padding(2, 2, 2, 2).
		Margin(1, 1, 0, 0).
		Width(28).
		Background(lipgloss.Color("#F9FAFB")).
		Foreground(lipgloss.Color("#1F2937"))

	t.SidebarHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Align(lipgloss.Center).
		Margin(0, 0, 2, 0).
		Background(lipgloss.Color("#F3F4F6")).
		Padding(1, 2)

	t.SidebarSection = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1F2937")).
		Bold(true).
		Margin(0, 0, 1, 0).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#D1D5DB")).
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false).
		Padding(0, 0, 1, 0)

	t.SidebarContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Margin(0, 0, 0, 2)

	t.SidebarSuccess = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669")).
		Bold(true)

	t.SidebarError = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DC2626")).
		Bold(true)

	t.UserMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Align(lipgloss.Left).
		Margin(0, 0, 0, 0)

	t.UserContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1F2937")).
		Align(lipgloss.Left).
		Margin(0, 0, 1, 0).
		Width(0)

	t.AssistantMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669")).
		Bold(true).
		Align(lipgloss.Left).
		Margin(0, 0, 0, 0)

	t.AssistantContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1F2937")).
		Align(lipgloss.Left).
		Margin(0, 0, 1, 0).
		Width(0)

	t.User = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DC2626")).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#DC2626")).
		Margin(0, 0, 1, 0).
		Width(0)

	t.Assistant = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669")).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#059669")).
		Margin(0, 0, 1, 0).
		Width(0)

	t.Spinner = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#059669")).
		Bold(true)

	// Main menu styles
	t.MainMenuBorder = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2563EB"))

	t.MainMenuContainer = lipgloss.NewStyle().
		Background(lipgloss.Color("#F9FAFB")).
		Padding(1, 2)

	t.MainMenuSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2563EB"))

	t.MainMenuWelcome = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true).
		Padding(1, 0)

	t.MainMenuLogo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true)

	t.MainMenuSubtitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true).
		Padding(1, 0)

	t.MainMenuTitleBorder = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#2563EB")).
		Align(lipgloss.Center).
		Padding(1, 2).
		Margin(1, 0)

	t.MainMenuSelectedArrow = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D97706")).
		Bold(true)

	t.MainMenuSelectedItem = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#D97706"))

	t.MainMenuUnselectedItem = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#2563EB"))

	t.MainMenuSelectedIcon = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D97706")).
		Bold(true)

	t.MainMenuUnselectedIcon = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	t.MainMenuSelectedTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#D97706"))

	t.MainMenuUnselectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1F2937"))

	t.MainMenuSelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D97706")).
		Italic(true).
		PaddingLeft(4).
		PaddingBottom(1)

	t.MainMenuUnselectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true).
		PaddingLeft(4).
		PaddingBottom(1)

	t.MainMenuHelpHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2563EB")).
		Bold(true).
		PaddingBottom(1)

	t.MainMenuHelpContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		PaddingLeft(2)

	t.MainMenuHelpBorder = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#2563EB")).
		Background(lipgloss.Color("#F3F4F6")).
		Padding(1, 2).
		Margin(1, 0)

	t.MainMenuSpinner = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2563EB"))

	// State selector styles
	t.StateSelectorContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Width(50).
		Height(15).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#2563EB")).
		Background(lipgloss.Color("#F9FAFB")).
		Padding(2, 3)

	t.StateSelectorDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#059669")).
		Padding(2, 3)

	t.StateSelectorTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#059669")).
		MarginBottom(2).
		Align(lipgloss.Center)

	t.StateSelectorHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Italic(true).
		Align(lipgloss.Center).
		MarginTop(2)

	t.StateSelectorItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Padding(0, 2).
		Margin(0, 0, 1, 0)

	t.StateSelectorSelectedItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#2563EB")).
		Bold(true).
		Padding(0, 2).
		Margin(0, 0, 1, 0)

	t.StateSelectorIndicator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D97706")).
		Bold(true)

	t.StateSelectorSelectedIndicator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	return t
}
