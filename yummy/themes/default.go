package themes

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// NewDefaultTheme creates a new default theme with all styles initialized
func NewDefaultTheme() Theme {
	t := Theme{
		Name: "default",
	}

	// Core styles
	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF6B6B")).
		Padding(0, 4)

	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	t.Info = lipgloss.NewStyle().
		BorderStyle(b).
		Foreground(lipgloss.Color("#87CEEB"))

	t.Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Padding(0, 2)

	t.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#98FB98")).
		PaddingTop(1).
		PaddingBottom(1)

	t.Ingredient = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDA0DD")).
		PaddingLeft(2)

	t.Doc = lipgloss.NewStyle().
		Margin(1, 2)

	t.DetailContent = lipgloss.NewStyle().
		Padding(1, 2)

	t.DetailHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF6B6B")).
		Padding(0, 4).
		MarginBottom(0)

	t.DetailFooter = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB")).
		Padding(0, 1).
		MarginTop(1)

	t.ScrollBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB"))

	t.Loading = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		Italic(true)

	t.Instruction = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F0E68C")).
		PaddingLeft(2)

	// Status line styles
	t.Warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F0E68C")).
		Background(lipgloss.Color("#2a2a2a")).
		Padding(0, 1)

	t.Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Background(lipgloss.Color("#2a2a2a")).
		Padding(0, 1)

	t.Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		Background(lipgloss.Color("#2a2a2a")).
		Padding(0, 1)

	t.StatusLine = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		Background(lipgloss.Color("#2a2a2a")).
		Padding(0, 1)

	t.StatusLineLeft = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1)

	t.StatusLineRight = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1)

	t.StatusLineMode = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#FF6B6B")).
		Bold(true).
		Padding(0, 1)

	t.StatusLineFile = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1)

	t.StatusLineInfo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1)

	t.StatusLineSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		Background(lipgloss.Color("#1a1a1a"))

	// List styles
	t.ListStyles = list.Styles{
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

	// Delegate styles for lists
	normalTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	normalDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0"))

	selectedColor := lipgloss.Color("#FF6B6B")
	selectedTitle := lipgloss.NewStyle().
		Foreground(selectedColor).
		BorderLeftForeground(selectedColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true)

	selectedDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6"))

	dimmedTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0"))

	dimmedDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0"))

	filterMatch := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
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
		Foreground(lipgloss.Color("#ff6b9d")).
		Bold(true).
		Margin(1, 0).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ff6b9d")).
		Padding(0, 2).
		Align(lipgloss.Center).
		Background(lipgloss.Color("#2a2a2a"))

	t.Chat = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Padding(1, 2).
		Margin(0, 1, 0, 0).
		Foreground(lipgloss.Color("#ffffff"))

	t.Sidebar = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		Padding(2, 2, 2, 2).
		Margin(1, 1, 0, 0).
		Width(28).
		Background(lipgloss.Color("#1a1a1a")).
		Foreground(lipgloss.Color("#ffffff"))

	t.SidebarHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff6b9d")).
		Bold(true).
		Align(lipgloss.Center).
		Margin(0, 0, 2, 0).
		Background(lipgloss.Color("#2a2a2a")).
		Padding(1, 2)

	t.SidebarSection = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true).
		Margin(0, 0, 1, 0).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false).
		Padding(0, 0, 1, 0)

	t.SidebarContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc")).
		Margin(0, 0, 0, 2)

	t.SidebarSuccess = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Bold(true)

	t.SidebarError = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff6b6b")).
		Bold(true)

	t.UserMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff6b9d")).
		Bold(true).
		Align(lipgloss.Left).
		Margin(0, 0, 0, 0)

	t.UserContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Align(lipgloss.Left).
		Margin(0, 0, 1, 0).
		Width(0)

	t.AssistantMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Bold(true).
		Align(lipgloss.Left).
		Margin(0, 0, 0, 0)

	t.AssistantContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Align(lipgloss.Left).
		Margin(0, 0, 1, 0).
		Width(0)

	t.User = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Margin(0, 0, 1, 0).
		Width(0)

	t.Assistant = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Margin(0, 0, 1, 0).
		Width(0)

	t.Spinner = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Bold(true)

	// Main menu styles
	t.MainMenuBorder = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB"))

	t.MainMenuContainer = lipgloss.NewStyle().
		Background(lipgloss.Color("#1A0B2E")).
		Padding(1, 2)

	t.MainMenuSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB"))

	t.MainMenuWelcome = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDA0DD")).
		Italic(true).
		Padding(1, 0)

	t.MainMenuLogo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B19CD9")).
		Bold(true)

	t.MainMenuSubtitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDA0DD")).
		Italic(true).
		Padding(1, 0)

	t.MainMenuTitleBorder = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#9370DB")).
		Align(lipgloss.Center).
		Padding(1, 2).
		Margin(1, 0)

	t.MainMenuSelectedArrow = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)

	t.MainMenuSelectedItem = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#FFD700"))

	t.MainMenuUnselectedItem = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#9370DB"))

	t.MainMenuSelectedIcon = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)

	t.MainMenuUnselectedIcon = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDA0DD"))

	t.MainMenuSelectedTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700"))

	t.MainMenuUnselectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E6E6FA"))

	t.MainMenuSelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Italic(true).
		PaddingLeft(4).
		PaddingBottom(1)

	t.MainMenuUnselectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B19CD9")).
		Italic(true).
		PaddingLeft(4).
		PaddingBottom(1)

	t.MainMenuHelpHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB")).
		Bold(true).
		PaddingBottom(1)

	t.MainMenuHelpContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDA0DD")).
		PaddingLeft(2)

	t.MainMenuHelpBorder = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#9370DB")).
		Background(lipgloss.Color("#2D1B3D")).
		Padding(1, 2).
		Margin(1, 0)

	t.MainMenuSpinner = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB"))

	// State selector styles
	t.StateSelectorContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Width(50).
		Height(15)

	t.StateSelectorDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4a9eff")).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(2, 3)

	t.StateSelectorTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4a9eff")).
		MarginBottom(2).
		Align(lipgloss.Center)

	t.StateSelectorHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Align(lipgloss.Center).
		MarginTop(2)

	t.StateSelectorItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc")).
		Padding(0, 1)

	t.StateSelectorSelectedItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#4a9eff")).
		Bold(true).
		Padding(0, 1)

	t.StateSelectorIndicator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)

	t.StateSelectorSelectedIndicator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	// Session selector styles
	t.SessionSelectorContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	t.SessionSelectorDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#04B575")).
		Padding(1, 2)

	t.SessionSelectorTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true).
		Margin(1, 0, 1, 2)

	t.SessionSelectorPagination = lipgloss.NewStyle().
		MarginLeft(2)

	t.SessionSelectorHelp = lipgloss.NewStyle().
		MarginLeft(2).
		Foreground(lipgloss.Color("#626262"))

	// Model selector styles
	t.ModelSelectorContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	t.ModelSelectorDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#04B575")).
		Padding(1, 2)

	t.ModelSelectorTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true).
		Margin(1, 0, 1, 2)

	t.ModelSelectorPagination = lipgloss.NewStyle().
		MarginLeft(2)

	t.ModelSelectorHelp = lipgloss.NewStyle().
		MarginLeft(2).
		Foreground(lipgloss.Color("#626262"))

	// Model selector delegate styles
	modelSelectorSelectedTitle := lipgloss.NewStyle().
		BorderLeftForeground(lipgloss.Color("#04B575")).
		Foreground(lipgloss.Color("#04B575")).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true)

	modelSelectorSelectedDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Italic(true)

	modelSelectorNormalTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	t.ModelSelectorDelegateStyles = list.DefaultItemStyles{
		NormalTitle:   modelSelectorNormalTitle,
		NormalDesc:    t.DelegateStyles.NormalDesc,
		SelectedTitle: modelSelectorSelectedTitle,
		SelectedDesc:  modelSelectorSelectedDesc,
		DimmedTitle:   t.DelegateStyles.DimmedTitle,
		DimmedDesc:    t.DelegateStyles.DimmedDesc,
		FilterMatch:   t.DelegateStyles.FilterMatch,
	}

	// Theme selector styles
	t.ThemeSelectorContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	t.ThemeSelectorDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(1, 2)

	t.ThemeSelectorTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true).
		Margin(1, 0, 1, 2)

	t.ThemeSelectorPagination = lipgloss.NewStyle().
		MarginLeft(2)

	t.ThemeSelectorHelp = lipgloss.NewStyle().
		MarginLeft(2).
		Foreground(lipgloss.Color("#626262"))

	// Theme selector delegate styles
	themeSelectorSelectedTitle := lipgloss.NewStyle().
		BorderLeftForeground(lipgloss.Color("#FF6B6B")).
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true)

	themeSelectorSelectedDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Italic(true)

	themeSelectorNormalTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	t.ThemeSelectorDelegateStyles = list.DefaultItemStyles{
		NormalTitle:   themeSelectorNormalTitle,
		NormalDesc:    t.DelegateStyles.NormalDesc,
		SelectedTitle: themeSelectorSelectedTitle,
		SelectedDesc:  themeSelectorSelectedDesc,
		DimmedTitle:   t.DelegateStyles.DimmedTitle,
		DimmedDesc:    t.DelegateStyles.DimmedDesc,
		FilterMatch:   t.DelegateStyles.FilterMatch,
	}

	return t
}
