package themes

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a complete theme with all styling components
type Theme struct {
	Name string

	// Core styles
	Title         lipgloss.Style
	Info          lipgloss.Style
	Error         lipgloss.Style
	Header        lipgloss.Style
	Ingredient    lipgloss.Style
	Doc           lipgloss.Style
	DetailContent lipgloss.Style
	DetailHeader  lipgloss.Style
	DetailFooter  lipgloss.Style
	ScrollBar     lipgloss.Style
	Loading       lipgloss.Style
	Instruction   lipgloss.Style

	// Status line styles
	Warning             lipgloss.Style
	Success             lipgloss.Style
	Help                lipgloss.Style
	StatusLine          lipgloss.Style
	StatusLineLeft      lipgloss.Style
	StatusLineRight     lipgloss.Style
	StatusLineMode      lipgloss.Style
	StatusLineFile      lipgloss.Style
	StatusLineInfo      lipgloss.Style
	StatusLineSeparator lipgloss.Style

	// List styles
	ListStyles     list.Styles
	DelegateStyles list.DefaultItemStyles

	// Chat styles
	ChatTitle        lipgloss.Style
	Chat             lipgloss.Style
	Sidebar          lipgloss.Style
	SidebarHeader    lipgloss.Style
	SidebarSection   lipgloss.Style
	SidebarContent   lipgloss.Style
	SidebarSuccess   lipgloss.Style
	SidebarError     lipgloss.Style
	UserMessage      lipgloss.Style
	UserContent      lipgloss.Style
	AssistantMessage lipgloss.Style
	AssistantContent lipgloss.Style
	User             lipgloss.Style
	Assistant        lipgloss.Style
	Spinner          lipgloss.Style

	// Main menu styles
	MainMenuBorder          lipgloss.Style
	MainMenuContainer       lipgloss.Style
	MainMenuSeparator       lipgloss.Style
	MainMenuWelcome         lipgloss.Style
	MainMenuLogo            lipgloss.Style
	MainMenuSubtitle        lipgloss.Style
	MainMenuTitleBorder     lipgloss.Style
	MainMenuSelectedArrow   lipgloss.Style
	MainMenuSelectedItem    lipgloss.Style
	MainMenuUnselectedItem  lipgloss.Style
	MainMenuSelectedIcon    lipgloss.Style
	MainMenuUnselectedIcon  lipgloss.Style
	MainMenuSelectedTitle   lipgloss.Style
	MainMenuUnselectedTitle lipgloss.Style
	MainMenuSelectedDesc    lipgloss.Style
	MainMenuUnselectedDesc  lipgloss.Style
	MainMenuHelpHeader      lipgloss.Style
	MainMenuHelpContent     lipgloss.Style
	MainMenuHelpBorder      lipgloss.Style
	MainMenuHelpKey         lipgloss.Style
	MainMenuHelpDesc        lipgloss.Style
	MainMenuSpinner         lipgloss.Style

	// State selector styles
	StateSelectorContainer         lipgloss.Style
	StateSelectorDialog            lipgloss.Style
	StateSelectorTitle             lipgloss.Style
	StateSelectorHelp              lipgloss.Style
	StateSelectorItem              lipgloss.Style
	StateSelectorSelectedItem      lipgloss.Style
	StateSelectorIndicator         lipgloss.Style
	StateSelectorSelectedIndicator lipgloss.Style

	// Session selector styles
	SessionSelectorContainer  lipgloss.Style
	SessionSelectorDialog     lipgloss.Style
	SessionSelectorTitle      lipgloss.Style
	SessionSelectorPagination lipgloss.Style
	SessionSelectorHelp       lipgloss.Style

	// Model selector styles
	ModelSelectorContainer      lipgloss.Style
	ModelSelectorDialog         lipgloss.Style
	ModelSelectorTitle          lipgloss.Style
	ModelSelectorPagination     lipgloss.Style
	ModelSelectorHelp           lipgloss.Style
	ModelSelectorDelegateStyles list.DefaultItemStyles

	// Theme selector styles
	ThemeSelectorContainer      lipgloss.Style
	ThemeSelectorDialog         lipgloss.Style
	ThemeSelectorTitle          lipgloss.Style
	ThemeSelectorPagination     lipgloss.Style
	ThemeSelectorHelp           lipgloss.Style
	ThemeSelectorDelegateStyles list.DefaultItemStyles

	// Add recipe from URL dialog styles
	AddRecipeFromURLContainer    lipgloss.Style
	AddRecipeFromURLDialog       lipgloss.Style
	AddRecipeFromURLTitle        lipgloss.Style
	AddRecipeFromURLHelp         lipgloss.Style
	AddRecipeFromURLPrompt       lipgloss.Style
	AddRecipeFromURLError        lipgloss.Style
	AddRecipeFromURLSeparator    lipgloss.Style
	AddRecipeFromURLSpinner      lipgloss.Style
	AddRecipeFromURLAccent       lipgloss.Style
	AddRecipeFromURLInputBorder  lipgloss.Style
	AddRecipeFromURLKeyHighlight lipgloss.Style

	// Recipe selector styles
	RecipeSelectorContainer lipgloss.Style
	RecipeSelectorDialog    lipgloss.Style
	RecipeSelectorTitle     lipgloss.Style
	RecipeSelectorHelp      lipgloss.Style
	RecipeSelectorSelected  lipgloss.Style

	// Command palette styles
	CommandPaletteContainer lipgloss.Style
	CommandPaletteDialog    lipgloss.Style
	CommandPaletteTitle     lipgloss.Style
	CommandPaletteHelp      lipgloss.Style
	CommandPaletteShortcut  lipgloss.Style
	CommandPaletteSelected  lipgloss.Style

	// Rating styles
	RatingBar             lipgloss.Style
	RatingStarActive      lipgloss.Style
	RatingStarInactive    lipgloss.Style
	RatingDialogContainer lipgloss.Style
	RatingDialogBox       lipgloss.Style
	RatingDialogTitle     lipgloss.Style
	RatingDialogHelp      lipgloss.Style

	// Cooking mode styles
	CookingStepCounter      lipgloss.Style
	CookingInstruction      lipgloss.Style
	CookingNavHint          lipgloss.Style
	CookingSidebar          lipgloss.Style
	CookingSidebarTitle     lipgloss.Style
	CookingIngredient       lipgloss.Style
	CookingIngredientAmount lipgloss.Style
	CookingIngredientDetail lipgloss.Style

	// Cooking chat styles
	CookingChatPanel lipgloss.Style
	CookingChatTitle lipgloss.Style

	// Cooking timer styles
	CookingTimerActive       lipgloss.Style
	CookingTimerDone         lipgloss.Style
	CookingTimerLabel        lipgloss.Style
	CookingTimerMessage      lipgloss.Style
	CookingTimerBarFilled    lipgloss.Style
	CookingTimerBarEmpty     lipgloss.Style
	CookingTimerBarCompleted lipgloss.Style

	// Shared textarea styles (chat + cooking)
	TextareaCursorLine  lipgloss.Style
	TextareaBase        lipgloss.Style
	TextareaPlaceholder lipgloss.Style
	TextareaText        lipgloss.Style
	TextareaPrompt      lipgloss.Style
	TextareaEndOfBuffer lipgloss.Style

	// Shared separator styles
	SeparatorLine    lipgloss.Style
	MessageSeparator lipgloss.Style

	// Shared dialog row styles
	DialogSelectedRow   lipgloss.Style
	DialogUnselectedRow lipgloss.Style

	// Session selector description rows
	SessionSelectorSelectedDesc   lipgloss.Style
	SessionSelectorUnselectedDesc lipgloss.Style

	// Sidebar value style
	SidebarValue lipgloss.Style

	// Chat empty state
	ChatEmptyState lipgloss.Style

	// Chat mention (@[Recipe]) styles
	ChatMention              lipgloss.Style // highlighted recipe mentions in messages
	ChatMentionPopupBorder   lipgloss.Style // popup container border
	ChatMentionPopupHeader   lipgloss.Style // "Recipes" header
	ChatMentionPopupItem     lipgloss.Style // unselected item
	ChatMentionPopupSelected lipgloss.Style // selected item

	// Cooking-specific styles
	CookingChatUserLabel       lipgloss.Style
	CookingChatAssistantLabel  lipgloss.Style
	CookingChatEmpty           lipgloss.Style
	CookingNoRecipe            lipgloss.Style
	CookingRecipeName          lipgloss.Style
	CookingProgressFilled      lipgloss.Style
	CookingProgressUnfilled    lipgloss.Style
	CookingIngredientHighlight lipgloss.Style
	CookingNavArrow            lipgloss.Style
	CookingHelpKey             lipgloss.Style
}

// Helper methods for main menu rendering
func (t *Theme) GetMainMenuBorderTop(width int) string {
	return ""
}

func (t *Theme) GetMainMenuBorderBottom(width int) string {
	return ""
}

func (t *Theme) GetMainMenuSeparator(width int) string {
	sep := strings.Repeat("─", width)
	return t.MainMenuSeparator.Render(sep)
}

func (t *Theme) RenderMainMenuItem(title, description string, isSelected bool, width int) string {
	var titleStyle, descStyle lipgloss.Style
	if isSelected {
		titleStyle = t.MainMenuSelectedTitle
		descStyle = t.MainMenuSelectedDesc
	} else {
		titleStyle = t.MainMenuUnselectedTitle
		descStyle = t.MainMenuUnselectedDesc
	}

	titleLine := titleStyle.Render(title)
	descLine := descStyle.Render(description)
	content := titleLine + "\n" + descLine

	if isSelected {
		return t.MainMenuSelectedItem.Width(width).Render(content)
	}
	return t.MainMenuUnselectedItem.Width(width).Render(content)
}

func (t *Theme) GetStateItemStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return t.StateSelectorSelectedItem
	}
	return t.StateSelectorItem
}

func (t *Theme) GetIndicatorStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return t.StateSelectorSelectedIndicator
	}
	return t.StateSelectorIndicator
}
