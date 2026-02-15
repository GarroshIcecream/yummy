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
		Padding(0, 1)

	t.DetailContent = lipgloss.NewStyle().
		Padding(0, 2)

	t.DetailHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4a9eff"))

	t.DetailFooter = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	t.ScrollBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB"))

	t.Loading = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		Italic(true)

	t.Instruction = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F0E68C")).
		PaddingLeft(2)

	// Status line styles
	statusBg := lipgloss.Color("#1a1a1a")

	t.Warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F0E68C")).
		Background(statusBg).
		Padding(0, 1)

	t.Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Background(statusBg).
		Padding(0, 1)

	t.Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Background(statusBg).
		Padding(0, 1)

	t.StatusLine = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Background(statusBg).
		Padding(0, 1)

	t.StatusLineLeft = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc")).
		Background(statusBg).
		Padding(0, 1)

	t.StatusLineRight = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Background(statusBg).
		Padding(0, 1)

	t.StatusLineMode = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4a9eff")).
		Bold(true).
		Padding(0, 1)

	t.StatusLineFile = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc")).
		Background(statusBg).
		Padding(0, 1)

	t.StatusLineInfo = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Background(statusBg).
		Padding(0, 1)

	t.StatusLineSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Background(statusBg)

	// List styles
	accentBlue := lipgloss.Color("#4a9eff")
	mutedGray := lipgloss.Color("#626262")
	dimGray := lipgloss.Color("#3a3a3a")

	t.ListStyles = list.Styles{
		TitleBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1),
		Spinner: lipgloss.NewStyle().
			Foreground(accentBlue),
		FilterPrompt: lipgloss.NewStyle().
			Foreground(accentBlue).
			Bold(true).
			Padding(0, 0, 0, 2),
		FilterCursor: lipgloss.NewStyle().
			Foreground(accentBlue).
			Bold(true),

		DefaultFilterCharacterMatch: lipgloss.NewStyle().
			Foreground(accentBlue).
			Underline(true),
		StatusBar: lipgloss.NewStyle().
			Foreground(mutedGray).
			Padding(0, 0, 1, 2),
		StatusEmpty: lipgloss.NewStyle().
			Foreground(mutedGray),
		StatusBarActiveFilter: lipgloss.NewStyle().
			Foreground(accentBlue),
		StatusBarFilterCount: lipgloss.NewStyle().
			Foreground(mutedGray),
		NoItems: lipgloss.NewStyle().
			Foreground(mutedGray).
			Padding(0, 0, 0, 2),
		PaginationStyle: lipgloss.NewStyle().
			PaddingLeft(2),
		HelpStyle: lipgloss.NewStyle().
			Foreground(mutedGray).
			Padding(1, 0, 0, 2),
		ActivePaginationDot: lipgloss.NewStyle().
			Foreground(accentBlue),
		InactivePaginationDot: lipgloss.NewStyle().
			Foreground(dimGray),
		ArabicPagination: lipgloss.NewStyle().
			Foreground(mutedGray),
		DividerDot: lipgloss.NewStyle().
			Foreground(dimGray),
	}

	// Delegate styles for lists
	normalTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0")).
		Bold(true).
		Padding(0, 0, 0, 1)

	normalDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#777777")).
		Padding(0, 0, 0, 1)

	selectedTitle := lipgloss.NewStyle().
		Foreground(accentBlue).
		Bold(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderTop(false).
		BorderBottom(false).
		BorderRight(false).
		BorderLeftForeground(accentBlue).
		PaddingLeft(1)

	selectedDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		PaddingLeft(2)

	dimmedTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555")).
		Padding(0, 0, 0, 1)

	dimmedDesc := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Padding(0, 0, 0, 1)

	filterMatch := lipgloss.NewStyle().
		Foreground(accentBlue).
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
	chatAccent := lipgloss.Color("#4a9eff")

	t.ChatTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 2)

	t.Chat = lipgloss.NewStyle().
		Padding(1, 2).
		Foreground(lipgloss.Color("#e0e0e0"))

	t.Sidebar = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderRight(false).
		BorderTop(false).
		BorderBottom(false).
		BorderLeftForeground(lipgloss.Color("#2a2a2a")).
		Padding(1, 2, 1, 2).
		Foreground(lipgloss.Color("#b0b0b0"))

	t.SidebarHeader = lipgloss.NewStyle().
		Foreground(chatAccent).
		Bold(true).
		MarginBottom(1)

	t.SidebarSection = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4a9eff")).
		Bold(true).
		MarginTop(1)

	t.SidebarContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	t.SidebarSuccess = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4"))

	t.SidebarError = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))

	t.UserMessage = lipgloss.NewStyle().
		Foreground(chatAccent).
		Bold(true)

	t.UserContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0")).
		Width(0)

	t.AssistantMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Bold(true)

	t.AssistantContent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0")).
		Width(0)

	t.User = lipgloss.NewStyle().
		Foreground(chatAccent).
		Bold(true).
		Width(0)

	t.Assistant = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")).
		Width(0)

	t.Spinner = lipgloss.NewStyle().
		Foreground(chatAccent)

	// Main menu styles
	menuAccent := lipgloss.Color("#4a9eff")
	menuDim := lipgloss.Color("#555555")

	t.MainMenuBorder = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333"))

	t.MainMenuContainer = lipgloss.NewStyle()

	t.MainMenuSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333"))

	t.MainMenuWelcome = lipgloss.NewStyle().
		Foreground(menuDim)

	t.MainMenuLogo = lipgloss.NewStyle().
		Foreground(menuAccent).
		Bold(true)

	t.MainMenuSubtitle = lipgloss.NewStyle().
		Foreground(menuDim)

	t.MainMenuTitleBorder = lipgloss.NewStyle().
		Align(lipgloss.Center)

	t.MainMenuSelectedArrow = lipgloss.NewStyle().
		Foreground(menuAccent).
		Bold(true)

	t.MainMenuSelectedItem = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(menuAccent).
		PaddingLeft(1)

	t.MainMenuUnselectedItem = lipgloss.NewStyle().
		PaddingLeft(3)

	t.MainMenuSelectedIcon = lipgloss.NewStyle().
		Foreground(menuAccent).
		Bold(true)

	t.MainMenuUnselectedIcon = lipgloss.NewStyle().
		Foreground(menuDim)

	t.MainMenuSelectedTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(menuAccent)

	t.MainMenuUnselectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999999"))

	t.MainMenuSelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc"))

	t.MainMenuUnselectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))

	t.MainMenuHelpHeader = lipgloss.NewStyle().
		Foreground(menuDim).
		Bold(true)

	t.MainMenuHelpContent = lipgloss.NewStyle().
		Foreground(menuDim)

	t.MainMenuHelpBorder = lipgloss.NewStyle()

	t.MainMenuHelpKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0"))

	t.MainMenuHelpDesc = lipgloss.NewStyle().
		Foreground(menuDim)

	t.MainMenuSpinner = lipgloss.NewStyle().
		Foreground(menuAccent)

	// State selector styles
	t.StateSelectorContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	t.StateSelectorDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4a9eff")).
		Padding(1, 2)

	t.StateSelectorTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF"))

	t.StateSelectorHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

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
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	t.SessionSelectorPagination = lipgloss.NewStyle().
		MarginLeft(2)

	t.SessionSelectorHelp = lipgloss.NewStyle().
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
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	t.ModelSelectorPagination = lipgloss.NewStyle().
		MarginLeft(2)

	t.ModelSelectorHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	// Theme selector styles
	t.ThemeSelectorContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	t.ThemeSelectorDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(1, 2)

	t.ThemeSelectorTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	t.ThemeSelectorPagination = lipgloss.NewStyle().
		MarginLeft(2)

	t.ThemeSelectorHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	// Legacy delegate styles kept for YAML theme compatibility
	themeSelectorSelectedTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)
	themeSelectorNormalTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
	t.ThemeSelectorDelegateStyles = list.DefaultItemStyles{
		NormalTitle:   themeSelectorNormalTitle,
		NormalDesc:    t.DelegateStyles.NormalDesc,
		SelectedTitle: themeSelectorSelectedTitle,
		SelectedDesc:  t.DelegateStyles.NormalDesc,
		DimmedTitle:   t.DelegateStyles.DimmedTitle,
		DimmedDesc:    t.DelegateStyles.DimmedDesc,
		FilterMatch:   t.DelegateStyles.FilterMatch,
	}

	// Add recipe from URL dialog styles
	t.AddRecipeFromURLContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)
	t.AddRecipeFromURLDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5C7AEA")).
		Padding(1, 3)
	t.AddRecipeFromURLTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	t.AddRecipeFromURLSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333"))
	t.AddRecipeFromURLHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))
	t.AddRecipeFromURLPrompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999999"))
	t.AddRecipeFromURLError = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))
	t.AddRecipeFromURLSpinner = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5C7AEA"))
	t.AddRecipeFromURLAccent = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5C7AEA")).
		Bold(true)
	t.AddRecipeFromURLInputBorder = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5C7AEA"))
	t.AddRecipeFromURLKeyHighlight = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5C7AEA")).
		Bold(true)

	// Recipe selector styles
	t.RecipeSelectorContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)
	t.RecipeSelectorDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(1, 2)
	t.RecipeSelectorTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	t.RecipeSelectorHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))
	t.RecipeSelectorSelected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4a4a4a")).
		Bold(true)

	// Command palette styles
	t.CommandPaletteContainer = lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)
	t.CommandPaletteDialog = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#B48EAD")).
		Padding(1, 2)
	t.CommandPaletteTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	t.CommandPaletteHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))
	t.CommandPaletteShortcut = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))
	t.CommandPaletteSelected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#B48EAD")).
		Bold(true)

	// Rating styles
	t.RatingBar = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("#999999"))
	t.RatingStarActive = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)
	t.RatingStarInactive = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))
	t.RatingDialogContainer = lipgloss.NewStyle().
		Align(lipgloss.Center)
	t.RatingDialogBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFD700")).
		Padding(0, 2)
	t.RatingDialogTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	t.RatingDialogHelp = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	// Cooking mode styles
	t.CookingStepCounter = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4a9eff")).
		Bold(true)
	t.CookingInstruction = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0")).
		Bold(true).
		Padding(1, 4)
	t.CookingNavHint = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))
	t.CookingSidebar = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		PaddingLeft(2).
		PaddingRight(1)
	t.CookingSidebarTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4a9eff")).
		Bold(true)
	t.CookingIngredient = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999999"))
	t.CookingIngredientAmount = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)
	t.CookingIngredientDetail = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#777777")).
		Italic(true)

	// Cooking chat styles
	t.CookingChatPanel = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#333333")).
		PaddingLeft(1).
		PaddingRight(1)
	t.CookingChatTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4a9eff")).
		Bold(true)

	// Cooking timer styles
	t.CookingTimerActive = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B35")).
		Bold(true)
	t.CookingTimerDone = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50C878")).
		Bold(true)
	t.CookingTimerLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)
	t.CookingTimerMessage = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999999")).
		Italic(true)
	t.CookingTimerBarFilled = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B35")).
		Background(lipgloss.Color("#4a2010"))
	t.CookingTimerBarEmpty = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555")).
		Background(lipgloss.Color("#1a1a1a"))
	t.CookingTimerBarCompleted = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50C878")).
		Background(lipgloss.Color("#1a3a25"))

	// Shared textarea styles
	t.TextareaCursorLine = lipgloss.NewStyle()
	t.TextareaBase = lipgloss.NewStyle().PaddingLeft(1)
	t.TextareaPlaceholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	t.TextareaText = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0e0e0"))
	t.TextareaPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#4a9eff")).Bold(true)
	t.TextareaEndOfBuffer = lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a1a"))

	// Shared separator styles
	t.SeparatorLine = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	t.MessageSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("#2a2a2a"))

	// Shared dialog row styles
	t.DialogSelectedRow = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4a4a4a")).
		Bold(true)
	t.DialogUnselectedRow = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc"))

	// Session selector description rows
	t.SessionSelectorSelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999999")).
		Background(lipgloss.Color("#4a4a4a"))
	t.SessionSelectorUnselectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	// Sidebar value style
	t.SidebarValue = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0e0e0"))

	// Chat empty state
	t.ChatEmptyState = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Italic(true)

	// Chat mention styles
	t.ChatMention = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#87CEEB"))
	t.ChatMentionPopupBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555"))
	t.ChatMentionPopupHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Italic(true).
		PaddingLeft(1)
	t.ChatMentionPopupItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		PaddingLeft(1).
		PaddingRight(1)
	t.ChatMentionPopupSelected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4ECDC4")). // matches AssistantMessage
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1)

	// Cooking-specific styles
	t.CookingChatUserLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4a9eff")).
		Bold(true)
	t.CookingChatAssistantLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50C878")).
		Bold(true)
	t.CookingChatEmpty = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555")).
		Italic(true)
	t.CookingNoRecipe = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555"))
	t.CookingRecipeName = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0")).
		Bold(true)
	t.CookingProgressFilled = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4a9eff"))
	t.CookingProgressUnfilled = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333"))
	t.CookingIngredientHighlight = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		Background(lipgloss.Color("#3a3000"))
	t.CookingNavArrow = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0"))
	t.CookingHelpKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0e0e0"))

	return t
}
