package themes

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

// Theme represents a complete theme with all styling components
type Theme struct {
	Name        string
	Description string

	// Palette holds the semantic colors the theme's styles were derived from.
	// Renderers that need raw colors (markdown, spinners) read them here.
	Palette Palette

	// BaseBg is the base background color for the theme, exposed so that
	// components (viewport, glamour, etc.) can use it directly.
	BaseBg color.Color

	// Core styles
	AppBackground lipgloss.Style
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
	ChatPanelTitle   lipgloss.Style // brand/title text shown in the chat panel's top border
	ChatPanelBorder  color.Color    // border color shared by the chat + input + sidebar panels
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

	// Confirmation dialog: destructive title and buttons (focused / not)
	DialogDangerTitle         lipgloss.Style
	DialogDangerButton        lipgloss.Style
	DialogDangerButtonBlurred lipgloss.Style
	DialogButton              lipgloss.Style
	DialogButtonFocused       lipgloss.Style

	// Shared search-input styles for popover dialogs (command palette,
	// model/session/theme/recipe selectors). Carries the dialog's elevated
	// background so the input fills its full width uniformly.
	DialogSearchStyles textinput.Styles

	// SurfaceBg is the background color of elevated surfaces (popover dialogs).
	// Exposed so views can fill header gaps and spacer lines uniformly via
	// SurfaceFill instead of re-deriving the color.
	SurfaceBg color.Color

	// Session selector description rows
	SessionSelectorSelectedDesc   lipgloss.Style
	SessionSelectorUnselectedDesc lipgloss.Style

	// Sidebar value style
	SidebarValue lipgloss.Style

	// Chat empty state
	ChatEmptyState      lipgloss.Style
	ChatEmptyStateTitle lipgloss.Style

	// Tool call display styles
	ToolCallLabel lipgloss.Style
	ToolCallInput lipgloss.Style

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

// SurfaceFill returns a style whose only attribute is the elevated-surface
// background. Use it to paint header gaps, spacer lines, and any padding that
// must blend with a popover dialog's surface instead of falling through to the
// terminal background.
func (t *Theme) SurfaceFill() lipgloss.Style {
	return lipgloss.NewStyle().Background(t.SurfaceBg)
}

// TextAreaStyles assembles a textarea.Styles from the theme's shared Textarea*
// fields so every multi-line input (chat composer, cooking-mode chat) renders
// with the same fully-filled background. Focused and blurred states share the
// styling; the cursor inherits the prompt's accent color.
func (t *Theme) TextAreaStyles() textarea.Styles {
	s := textarea.DefaultStyles(true)
	for _, state := range []*textarea.StyleState{&s.Focused, &s.Blurred} {
		state.Base = t.TextareaBase
		state.Text = t.TextareaText
		state.Placeholder = t.TextareaPlaceholder
		state.Prompt = t.TextareaPrompt
		state.CursorLine = t.TextareaCursorLine
		state.EndOfBuffer = t.TextareaEndOfBuffer
	}
	if fg := t.TextareaPrompt.GetForeground(); isSet(fg) {
		s.Cursor.Color = fg
	}
	return s
}

// isSet reports whether a style color getter returned a real color: unset
// colors come back as lipgloss.NoColor, not nil.
func isSet(c color.Color) bool {
	if c == nil {
		return false
	}
	_, none := c.(lipgloss.NoColor)
	return !none
}

// ComposerPromptWidth is the display width of ComposerPrompt.
const ComposerPromptWidth = 3

// ComposerPrompt is the prompt for multi-line composers (chat, cooking chat):
// " > " on the first row only, so an empty composer reads as one field. The
// leading space is part of the prompt rather than textarea Base padding,
// because the textarea applies Base padding to the prompt again when it
// renders the placeholder, which shifted the prompt a column to the right
// until the user started typing.
func ComposerPrompt(info textarea.PromptInfo) string {
	if info.LineNumber == 0 {
		return " > "
	}
	return "   "
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
