package themes

import (
	"image/color"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	lipgloss "charm.land/lipgloss/v2"
)

// NewDefaultTheme creates the built-in default theme.
func NewDefaultTheme() Theme {
	return NewThemeFromPalette(
		"default",
		"Built-in default theme with bright accents on a dark surface",
		DefaultPalette(),
	)
}

// NewThemeFromPalette derives every style in the app from a palette. All
// views share the same small vocabulary so the UI stays consistent under any
// theme:
//
//   - everything sits on p.Base; the status bar and the selected main-menu
//     item are raised onto p.Surface; popover dialogs float on p.Overlay;
//   - every dialog has the same rounded p.Primary border, bold title, subtle
//     help text and p.Selection-highlighted row;
//   - text uses three levels (Text / Muted / Subtle), accents are Primary
//     (focus, titles), Secondary (assistant, links) and Accent (amounts,
//     stars, timers).
func NewThemeFromPalette(name, description string, p Palette) Theme {
	t := Theme{
		Name:        name,
		Description: description,
		Palette:     p,
	}

	base := p.Base
	surface := p.Surface
	overlay := p.Overlay
	selection := p.Selection
	border := p.Border
	text := p.Text
	muted := p.Muted
	subtle := p.Subtle
	primary := p.Primary
	secondary := p.Secondary
	accent := p.Accent

	t.BaseBg = base
	t.SurfaceBg = overlay

	fg := func(c color.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(c) }

	// Core styles
	t.AppBackground = lipgloss.NewStyle().Background(base).Foreground(text)

	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(p.OnPrimary).
		Background(primary).
		Padding(0, 4)

	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	t.Info = lipgloss.NewStyle().
		BorderStyle(b).
		BorderForeground(border).
		Foreground(secondary)

	t.Error = fg(p.Error).Padding(0, 2)
	t.Header = fg(primary).Bold(true).PaddingTop(1).PaddingBottom(1)
	t.Ingredient = fg(text).PaddingLeft(2)

	t.Doc = lipgloss.NewStyle().
		Foreground(text).
		Background(base).
		Padding(0, 1)

	t.DetailContent = lipgloss.NewStyle().
		Foreground(text).
		Background(base).
		Padding(0, 2)

	t.DetailHeader = fg(primary).Bold(true)
	t.DetailFooter = fg(subtle)
	t.ScrollBar = fg(primary)
	t.Loading = fg(muted).Italic(true)
	t.Instruction = fg(accent).PaddingLeft(2)

	// Inline status colors. Foreground only so they blend into whatever
	// surface they are rendered on.
	t.Warning = fg(p.Warning)
	t.Success = fg(p.Success)
	t.Help = fg(subtle)

	// Status line: one raised bar across the bottom of the screen.
	t.StatusLine = fg(subtle).Background(surface).Padding(0, 1)
	t.StatusLineLeft = fg(text).Background(surface).Padding(0, 1)
	t.StatusLineRight = fg(muted).Background(surface).Padding(0, 1)
	t.StatusLineMode = lipgloss.NewStyle().
		Foreground(p.OnPrimary).
		Background(primary).
		Bold(true).
		Padding(0, 1)
	t.StatusLineFile = fg(text).Background(surface).Padding(0, 1)
	t.StatusLineInfo = fg(muted).Background(surface).Padding(0, 1)
	t.StatusLineSeparator = fg(border).Background(surface)

	// List styles
	filterStyles := textinput.DefaultStyles(true)
	filterPrompt := fg(primary).Background(base).Bold(true)
	filterStyles.Focused.Prompt = filterPrompt
	filterStyles.Blurred.Prompt = filterPrompt
	filterStyles.Cursor.Color = primary
	filterStyles.Focused.Text = fg(text).Background(base)
	filterStyles.Blurred.Text = fg(text).Background(base)
	filterStyles.Focused.Placeholder = fg(subtle).Background(base)
	filterStyles.Blurred.Placeholder = fg(subtle).Background(base)
	filterStyles.Focused.Suggestion = fg(subtle).Background(base)
	filterStyles.Blurred.Suggestion = fg(subtle).Background(base)

	t.ListStyles = list.Styles{
		TitleBar: lipgloss.NewStyle().Padding(1, 0, 1, 2),
		Title:    fg(primary).Bold(true),
		Spinner:  fg(primary),
		Filter:   filterStyles,

		DefaultFilterCharacterMatch: fg(primary).Underline(true),
		StatusBar:                   fg(subtle).Padding(0, 0, 1, 2),
		StatusEmpty:                 fg(subtle),
		StatusBarActiveFilter:       fg(primary),
		StatusBarFilterCount:        fg(subtle),
		NoItems:                     fg(subtle).Padding(0, 0, 0, 2),
		PaginationStyle:             lipgloss.NewStyle().PaddingLeft(2),
		HelpStyle:                   fg(subtle).Padding(1, 0, 0, 2),
		ActivePaginationDot:         fg(primary).SetString("■"),
		InactivePaginationDot:       fg(border).SetString("□"),
		ArabicPagination:            fg(subtle),
		DividerDot:                  fg(border).SetString(" | "),
	}

	// Delegate styles for lists. Every row starts its text in the same column:
	// unselected rows pad 2, the selected row draws a 1-cell accent bar plus 1
	// cell of padding. The bubbles delegate truncates using the normal style's
	// padding, so keeping the totals equal also keeps long descriptions from
	// wrapping on the selected row.
	accentBar := func(s lipgloss.Style) lipgloss.Style {
		return s.
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderTop(false).
			BorderBottom(false).
			BorderRight(false).
			BorderLeftForeground(primary).
			PaddingLeft(1)
	}

	t.DelegateStyles = list.DefaultItemStyles{
		NormalTitle:   fg(text).Bold(true).PaddingLeft(2),
		NormalDesc:    fg(muted).PaddingLeft(2),
		SelectedTitle: accentBar(fg(primary).Bold(true)),
		SelectedDesc:  accentBar(fg(muted)),
		DimmedTitle:   fg(subtle).PaddingLeft(2),
		DimmedDesc:    fg(border).PaddingLeft(2),
		FilterMatch:   fg(primary).Underline(true),
	}

	// Chat styles
	t.ChatTitle = fg(text).Bold(true).Padding(0, 2)
	t.Chat = lipgloss.NewStyle().Padding(1, 2).Foreground(text).Background(base)
	t.ChatPanelTitle = fg(text).Background(base).Bold(true)
	t.ChatPanelBorder = border

	t.Sidebar = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderRight(false).
		BorderTop(false).
		BorderBottom(false).
		BorderLeftForeground(border).
		Padding(1, 2, 1, 2).
		Foreground(muted).
		Background(base)

	t.SidebarHeader = fg(primary).Bold(true).MarginBottom(1)
	t.SidebarSection = fg(primary).Bold(true).MarginTop(1)
	t.SidebarContent = fg(subtle).Background(base)
	t.SidebarSuccess = fg(secondary).Background(base)
	t.SidebarError = fg(p.Error).Background(base)
	t.SidebarValue = fg(text).Background(base)

	t.UserMessage = fg(primary).Background(base).Bold(true)
	t.UserContent = fg(text).Background(base)
	t.AssistantMessage = fg(secondary).Background(base).Bold(true)
	t.AssistantContent = fg(text).Background(base)
	t.User = fg(primary).Background(base).Bold(true)
	t.Assistant = fg(secondary).Background(base)
	t.Spinner = fg(primary).Background(base)

	// Main menu styles
	t.MainMenuBorder = fg(border).Background(base)
	t.MainMenuContainer = lipgloss.NewStyle().Background(base).Foreground(text)
	t.MainMenuSeparator = fg(border).Background(base)
	t.MainMenuWelcome = fg(subtle).Background(base)
	t.MainMenuLogo = fg(primary).Background(base).Bold(true)
	t.MainMenuSubtitle = fg(subtle).Background(base)
	t.MainMenuTitleBorder = lipgloss.NewStyle().Align(lipgloss.Center)
	t.MainMenuSelectedArrow = fg(primary).Bold(true)

	// The selected item draws a 1-cell accent bar + 1 cell padding; unselected
	// items pad 2 so every title starts in the same column.
	t.MainMenuSelectedItem = lipgloss.NewStyle().
		Background(surface).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(primary).
		BorderBackground(base).
		PaddingLeft(1)
	t.MainMenuUnselectedItem = lipgloss.NewStyle().Background(base).PaddingLeft(2)

	t.MainMenuSelectedIcon = fg(primary).Background(surface).Bold(true)
	t.MainMenuUnselectedIcon = fg(subtle).Background(base)
	t.MainMenuSelectedTitle = fg(primary).Background(surface).Bold(true)
	t.MainMenuUnselectedTitle = fg(text).Background(base)
	t.MainMenuSelectedDesc = fg(muted).Background(surface)
	t.MainMenuUnselectedDesc = fg(subtle).Background(base)
	t.MainMenuHelpHeader = fg(subtle).Bold(true)
	t.MainMenuHelpContent = fg(subtle).Background(base)
	t.MainMenuHelpBorder = lipgloss.NewStyle().Background(base)
	t.MainMenuHelpKey = fg(muted).Background(base).Bold(true)
	t.MainMenuHelpDesc = fg(subtle).Background(base)
	t.MainMenuSpinner = fg(primary).Background(base)

	// ── Popover dialogs ────────────────────────────────────────────────────
	// One shared look: overlay surface, rounded primary border, bold title,
	// subtle help, and a Selection-highlighted row.
	container := lipgloss.NewStyle().
		Background(base).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)
	dialogBox := lipgloss.NewStyle().
		Background(overlay).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primary).
		BorderBackground(base).
		Padding(1, 2)
	dialogTitle := fg(text).Background(overlay).Bold(true)
	dialogHelp := fg(subtle).Background(overlay)
	selectedRow := fg(primary).Background(selection).Bold(true)
	unselectedRow := fg(text).Background(overlay)

	t.DialogSelectedRow = selectedRow
	t.DialogUnselectedRow = unselectedRow

	button := lipgloss.NewStyle().Padding(0, 2).Bold(true)
	t.DialogDangerTitle = fg(p.Error).Background(overlay).Bold(true)
	t.DialogDangerButton = button.Foreground(p.OnPrimary).Background(p.Error)
	t.DialogDangerButtonBlurred = button.Foreground(p.Error).Background(selection)
	t.DialogButton = button.Foreground(muted).Background(selection)
	t.DialogButtonFocused = button.Foreground(p.OnPrimary).Background(primary)

	t.StateSelectorContainer = container
	t.StateSelectorDialog = dialogBox
	t.StateSelectorTitle = dialogTitle
	t.StateSelectorHelp = dialogHelp
	t.StateSelectorItem = unselectedRow.Padding(0, 1)
	t.StateSelectorSelectedItem = selectedRow.Padding(0, 1)
	t.StateSelectorIndicator = fg(subtle).Background(overlay)
	t.StateSelectorSelectedIndicator = fg(primary).Background(selection).Bold(true)

	t.SessionSelectorContainer = container
	t.SessionSelectorDialog = dialogBox
	t.SessionSelectorTitle = dialogTitle
	t.SessionSelectorPagination = lipgloss.NewStyle().PaddingLeft(2)
	t.SessionSelectorHelp = dialogHelp
	t.SessionSelectorSelectedDesc = fg(muted).Background(selection)
	t.SessionSelectorUnselectedDesc = fg(subtle).Background(overlay)

	t.ModelSelectorContainer = container
	t.ModelSelectorDialog = dialogBox
	t.ModelSelectorTitle = dialogTitle
	t.ModelSelectorPagination = lipgloss.NewStyle().PaddingLeft(2)
	t.ModelSelectorHelp = dialogHelp

	t.ThemeSelectorContainer = container
	t.ThemeSelectorDialog = dialogBox
	t.ThemeSelectorTitle = dialogTitle
	t.ThemeSelectorPagination = lipgloss.NewStyle().PaddingLeft(2)
	t.ThemeSelectorHelp = dialogHelp

	dialogDelegate := list.DefaultItemStyles{
		NormalTitle:   unselectedRow,
		NormalDesc:    fg(subtle).Background(overlay),
		SelectedTitle: selectedRow,
		SelectedDesc:  fg(muted).Background(selection),
		DimmedTitle:   fg(subtle).Background(overlay),
		DimmedDesc:    fg(border).Background(overlay),
		FilterMatch:   fg(primary).Underline(true),
	}
	t.ThemeSelectorDelegateStyles = dialogDelegate
	t.ModelSelectorDelegateStyles = dialogDelegate

	t.AddRecipeFromURLContainer = container
	t.AddRecipeFromURLDialog = dialogBox
	t.AddRecipeFromURLTitle = dialogTitle
	t.AddRecipeFromURLSeparator = fg(border).Background(overlay)
	t.AddRecipeFromURLHelp = dialogHelp
	t.AddRecipeFromURLPrompt = fg(muted).Background(overlay)
	t.AddRecipeFromURLError = fg(p.Error).Background(overlay)
	t.AddRecipeFromURLSpinner = fg(primary).Background(overlay)
	t.AddRecipeFromURLAccent = fg(primary).Background(overlay).Bold(true)
	t.AddRecipeFromURLInputBorder = fg(primary).Background(overlay)
	t.AddRecipeFromURLKeyHighlight = fg(muted).Background(overlay).Bold(true)

	t.RecipeSelectorContainer = container
	t.RecipeSelectorDialog = dialogBox
	t.RecipeSelectorTitle = dialogTitle
	t.RecipeSelectorHelp = dialogHelp
	t.RecipeSelectorSelected = selectedRow

	t.CommandPaletteContainer = container
	t.CommandPaletteDialog = dialogBox
	t.CommandPaletteTitle = dialogTitle
	t.CommandPaletteHelp = dialogHelp
	t.CommandPaletteShortcut = fg(subtle).Background(overlay)
	t.CommandPaletteSelected = selectedRow

	// Rating
	t.RatingBar = fg(muted).PaddingLeft(2)
	t.RatingStarActive = fg(accent).Background(overlay).Bold(true)
	t.RatingStarInactive = fg(subtle).Background(overlay)
	t.RatingDialogContainer = lipgloss.NewStyle().Background(base).Align(lipgloss.Center)
	t.RatingDialogBox = dialogBox.Padding(1, 3)
	t.RatingDialogTitle = dialogTitle
	t.RatingDialogHelp = dialogHelp

	// Shared search-input styles for popover dialogs. The textinput fills its
	// Width() using the Text/Placeholder style's background, so setting the
	// overlay background here makes the whole search line uniform.
	searchStyles := textinput.DefaultStyles(true)
	searchPrompt := fg(primary).Background(overlay).Bold(true)
	searchStyles.Focused.Prompt = searchPrompt
	searchStyles.Blurred.Prompt = searchPrompt
	searchStyles.Focused.Text = fg(text).Background(overlay)
	searchStyles.Blurred.Text = fg(text).Background(overlay)
	searchStyles.Focused.Placeholder = fg(subtle).Background(overlay)
	searchStyles.Blurred.Placeholder = fg(subtle).Background(overlay)
	searchStyles.Focused.Suggestion = fg(subtle).Background(overlay)
	searchStyles.Blurred.Suggestion = fg(subtle).Background(overlay)
	searchStyles.Cursor.Color = primary
	t.DialogSearchStyles = searchStyles

	// Cooking mode
	t.CookingStepCounter = fg(primary).Bold(true)
	t.CookingInstruction = fg(text).Background(base).Bold(true).Padding(1, 4)
	t.CookingNavHint = fg(subtle)
	t.CookingSidebar = lipgloss.NewStyle().
		Background(base).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(border).
		BorderBackground(base).
		PaddingLeft(2).
		PaddingRight(1)
	t.CookingSidebarTitle = fg(primary).Bold(true)
	t.CookingIngredient = fg(text)
	t.CookingIngredientAmount = fg(accent).Bold(true)
	t.CookingIngredientDetail = fg(subtle).Italic(true)

	t.CookingChatPanel = lipgloss.NewStyle().
		Background(base).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(border).
		BorderBackground(base).
		PaddingLeft(1).
		PaddingRight(1)
	t.CookingChatTitle = fg(primary).Bold(true)

	t.CookingTimerActive = fg(p.Warning).Bold(true)
	t.CookingTimerDone = fg(p.Success).Bold(true)
	t.CookingTimerLabel = fg(accent).Bold(true)
	t.CookingTimerMessage = fg(muted).Italic(true)
	t.CookingTimerBarFilled = fg(p.Warning)
	t.CookingTimerBarEmpty = fg(border)
	t.CookingTimerBarCompleted = fg(p.Success)

	t.CookingChatUserLabel = fg(primary).Bold(true)
	t.CookingChatAssistantLabel = fg(secondary).Bold(true)
	t.CookingChatEmpty = fg(subtle).Italic(true)
	t.CookingNoRecipe = fg(subtle)
	t.CookingRecipeName = fg(text).Bold(true)
	t.CookingProgressFilled = fg(primary)
	t.CookingProgressUnfilled = fg(border)
	t.CookingIngredientHighlight = fg(accent).Bold(true)
	t.CookingNavArrow = fg(text)
	t.CookingHelpKey = fg(muted).Bold(true)

	// Shared textarea styles. The chat page is one flat surface — the composer
	// sits on the same canvas as the conversation, separated only by its panel
	// border (popover dialogs keep their own overlay surface).
	t.TextareaCursorLine = lipgloss.NewStyle().Background(base)
	t.TextareaBase = lipgloss.NewStyle().Background(base)
	t.TextareaPlaceholder = fg(subtle).Background(base)
	t.TextareaText = fg(text).Background(base)
	t.TextareaPrompt = fg(primary).Background(base).Bold(true)
	t.TextareaEndOfBuffer = fg(base).Background(base)

	// Shared separators
	t.SeparatorLine = fg(border).Background(base)
	t.MessageSeparator = fg(border).Background(base)

	// Chat empty state
	t.ChatEmptyState = fg(subtle).Background(base).Italic(true)
	t.ChatEmptyStateTitle = fg(text).Background(base).Bold(true)

	// Tool call display
	t.ToolCallLabel = fg(accent).Background(base).Bold(true)
	t.ToolCallInput = fg(muted).Background(base).Italic(true)

	// Recipe mentions render as a tinted chip so @Recipe references stand out
	// from body text.
	t.ChatMention = fg(primary).Background(selection).Bold(true)
	t.ChatMentionPopupBorder = lipgloss.NewStyle().
		Background(overlay).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primary).
		BorderBackground(base)
	t.ChatMentionPopupHeader = fg(subtle).Background(overlay).Italic(true).PaddingLeft(1)
	t.ChatMentionPopupItem = fg(muted).Background(overlay).PaddingLeft(1).PaddingRight(1)
	t.ChatMentionPopupSelected = fg(primary).Background(selection).Bold(true).PaddingLeft(1).PaddingRight(1)

	return t
}
