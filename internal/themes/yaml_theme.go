package themes

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"image/color"

	lipgloss "charm.land/lipgloss/v2"
	"gopkg.in/yaml.v3"
)

// YAMLTheme represents a theme defined in YAML format
type YAMLTheme struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Colors      map[string]string `yaml:"colors"`
	// Palette sets the semantic colors every style is derived from (see
	// Palette). Values are hex colors or names from Colors. Unset entries fall
	// back to the default palette. Styles below override individual styles on
	// top of the palette-derived ones.
	Palette map[string]string    `yaml:"palette,omitempty"`
	Styles  map[string]YAMLStyle `yaml:"styles"`
	Lists   YAMLListStyles       `yaml:"lists,omitempty"`
}

// YAMLStyle represents a single style configuration
type YAMLStyle struct {
	Foreground    string `yaml:"foreground,omitempty"`
	Background    string `yaml:"background,omitempty"`
	Bold          bool   `yaml:"bold,omitempty"`
	Italic        bool   `yaml:"italic,omitempty"`
	Underline     bool   `yaml:"underline,omitempty"`
	Strikethrough bool   `yaml:"strikethrough,omitempty"`
	Padding       string `yaml:"padding,omitempty"` // "top,right,bottom,left" or "all" or "horizontal,vertical"
	Margin        string `yaml:"margin,omitempty"`  // "top,right,bottom,left" or "all" or "horizontal,vertical"
	Border        string `yaml:"border,omitempty"`  // "normal", "rounded", "double", "thick", "none"
	BorderColor   string `yaml:"border_color,omitempty"`
	Align         string `yaml:"align,omitempty"` // "left", "center", "right"
	Width         int    `yaml:"width,omitempty"`
	Height        int    `yaml:"height,omitempty"`
}

// YAMLListStyles represents list-specific styles
type YAMLListStyles struct {
	TitleBar                    YAMLStyle `yaml:"title_bar,omitempty"`
	Title                       YAMLStyle `yaml:"title,omitempty"`
	Spinner                     YAMLStyle `yaml:"spinner,omitempty"`
	FilterPrompt                YAMLStyle `yaml:"filter_prompt,omitempty"`
	FilterCursor                YAMLStyle `yaml:"filter_cursor,omitempty"`
	DefaultFilterCharacterMatch YAMLStyle `yaml:"default_filter_character_match,omitempty"`
	StatusBar                   YAMLStyle `yaml:"status_bar,omitempty"`
	StatusEmpty                 YAMLStyle `yaml:"status_empty,omitempty"`
	StatusBarActiveFilter       YAMLStyle `yaml:"status_bar_active_filter,omitempty"`
	StatusBarFilterCount        YAMLStyle `yaml:"status_bar_filter_count,omitempty"`
	NoItems                     YAMLStyle `yaml:"no_items,omitempty"`
	PaginationStyle             YAMLStyle `yaml:"pagination_style,omitempty"`
	HelpStyle                   YAMLStyle `yaml:"help_style,omitempty"`
	ActivePaginationDot         YAMLStyle `yaml:"active_pagination_dot,omitempty"`
	InactivePaginationDot       YAMLStyle `yaml:"inactive_pagination_dot,omitempty"`
	ArabicPagination            YAMLStyle `yaml:"arabic_pagination,omitempty"`
	DividerDot                  YAMLStyle `yaml:"divider_dot,omitempty"`
	// Delegate styles
	NormalTitle   YAMLStyle `yaml:"normal_title,omitempty"`
	NormalDesc    YAMLStyle `yaml:"normal_desc,omitempty"`
	SelectedTitle YAMLStyle `yaml:"selected_title,omitempty"`
	SelectedDesc  YAMLStyle `yaml:"selected_desc,omitempty"`
	DimmedTitle   YAMLStyle `yaml:"dimmed_title,omitempty"`
	DimmedDesc    YAMLStyle `yaml:"dimmed_desc,omitempty"`
	FilterMatch   YAMLStyle `yaml:"filter_match,omitempty"`
}

// LoadThemeFromYAML loads a theme from a YAML file
func LoadThemeFromYAML(filename string) (*Theme, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseYAMLTheme(data)
}

func parseYAMLTheme(data []byte) (*Theme, error) {
	var yamlTheme YAMLTheme
	if err := yaml.Unmarshal(data, &yamlTheme); err != nil {
		return nil, err
	}

	theme, err := yamlTheme.ToTheme()
	if err != nil {
		return nil, err
	}
	return &theme, nil
}

// LoadThemesFromDirectory loads all YAML themes from a directory
func LoadThemesFromDirectory(dir string) ([]Theme, error) {
	var themes []Theme
	seenNames := make(map[string]string)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml")) {
			slog.Info("Skipping entry", "name", entry.Name())
			continue
		}

		themePath := filepath.Join(dir, entry.Name())
		theme, err := LoadThemeFromYAML(themePath)
		if err != nil {
			slog.Error("Failed to load theme", "path", themePath, "error", err)
			continue
		}
		if prevPath, exists := seenNames[theme.Name]; exists {
			slog.Error("Skipping duplicate theme name", "name", theme.Name, "path", themePath, "existing_path", prevPath)
			continue
		}
		seenNames[theme.Name] = themePath

		themes = append(themes, *theme)
	}

	return themes, nil
}

// ToTheme converts a YAMLTheme to a Theme
func (yt *YAMLTheme) ToTheme() (Theme, error) {
	if err := validateYAMLTheme(yt); err != nil {
		return Theme{}, err
	}

	// Helper function to resolve color references, returning a plain string
	resolveColor := func(color string) string {
		if color == "" {
			return ""
		}
		// Check if it's a color reference
		if resolved, exists := yt.Colors[color]; exists {
			return resolved
		}
		return color
	}

	// Derive every style from the palette (defaults for unset entries) so any
	// style the theme does not override still matches the theme's colors.
	palette := DefaultPalette()
	if err := palette.apply(yt.Palette, resolveColor); err != nil {
		return Theme{}, err
	}
	theme := NewThemeFromPalette(yt.Name, yt.Description, palette)

	// Helper function to create a v2 lipgloss style from YAMLStyle
	createStyle := func(ys YAMLStyle) (lipgloss.Style, error) {
		style := lipgloss.NewStyle()

		if ys.Foreground != "" {
			style = style.Foreground(lipgloss.Color(resolveColor(ys.Foreground)))
		}
		if ys.Background != "" {
			style = style.Background(lipgloss.Color(resolveColor(ys.Background)))
		}
		if ys.Bold {
			style = style.Bold(true)
		}
		if ys.Italic {
			style = style.Italic(true)
		}
		if ys.Underline {
			style = style.Underline(true)
		}
		if ys.Strikethrough {
			style = style.Strikethrough(true)
		}

		// Handle padding
		if ys.Padding != "" {
			var err error
			style, err = applyPadding(style, ys.Padding)
			if err != nil {
				return style, err
			}
		}

		// Handle margin
		if ys.Margin != "" {
			var err error
			style, err = applyMargin(style, ys.Margin)
			if err != nil {
				return style, err
			}
		}

		// Handle border
		if ys.Border != "" && ys.Border != "none" {
			var err error
			style, err = applyBorder(style, ys.Border, lipgloss.Color(resolveColor(ys.BorderColor)))
			if err != nil {
				return style, err
			}
		}

		// Handle alignment
		if ys.Align != "" {
			var err error
			style, err = applyAlign(style, ys.Align)
			if err != nil {
				return style, err
			}
		}

		if ys.Width > 0 {
			style = style.Width(ys.Width)
		}
		if ys.Height > 0 {
			style = style.Height(ys.Height)
		}

		return style, nil
	}

	// Convert all styles
	for styleName, yamlStyle := range yt.Styles {
		style, err := createStyle(yamlStyle)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid style %q: %w", styleName, err)
		}

		// Use reflection or a switch statement to set the appropriate field
		switch styleName {
		case "app_background":
			theme.AppBackground = style
			if bg := style.GetBackground(); isSet(bg) {
				theme.BaseBg = bg
			}
		case "title":
			theme.Title = style
		case "info":
			theme.Info = style
		case "error":
			theme.Error = style
		case "header":
			theme.Header = style
		case "ingredient":
			theme.Ingredient = style
		case "doc":
			theme.Doc = style
		case "detail_content":
			theme.DetailContent = style
		case "detail_header":
			theme.DetailHeader = style
		case "detail_footer":
			theme.DetailFooter = style
		case "scroll_bar":
			theme.ScrollBar = style
		case "loading":
			theme.Loading = style
		case "instruction":
			theme.Instruction = style
		case "warning":
			theme.Warning = style
		case "success":
			theme.Success = style
		case "help":
			theme.Help = style
		case "status_line":
			theme.StatusLine = style
		case "status_line_left":
			theme.StatusLineLeft = style
		case "status_line_right":
			theme.StatusLineRight = style
		case "status_line_mode":
			theme.StatusLineMode = style
		case "status_line_file":
			theme.StatusLineFile = style
		case "status_line_info":
			theme.StatusLineInfo = style
		case "status_line_separator":
			theme.StatusLineSeparator = style
		case "chat_title":
			theme.ChatTitle = style
		case "chat":
			theme.Chat = style
		case "sidebar":
			theme.Sidebar = style
		case "sidebar_header":
			theme.SidebarHeader = style
		case "sidebar_section":
			theme.SidebarSection = style
		case "sidebar_content":
			theme.SidebarContent = style
		case "sidebar_success":
			theme.SidebarSuccess = style
		case "sidebar_error":
			theme.SidebarError = style
		case "user_message":
			theme.UserMessage = style
		case "user_content":
			theme.UserContent = style
		case "assistant_message":
			theme.AssistantMessage = style
		case "assistant_content":
			theme.AssistantContent = style
		case "user":
			theme.User = style
		case "assistant":
			theme.Assistant = style
		case "spinner":
			theme.Spinner = style
		case "main_menu_border":
			theme.MainMenuBorder = style
		case "main_menu_container":
			theme.MainMenuContainer = style
		case "main_menu_separator":
			theme.MainMenuSeparator = style
		case "main_menu_welcome":
			theme.MainMenuWelcome = style
		case "main_menu_logo":
			theme.MainMenuLogo = style
		case "main_menu_subtitle":
			theme.MainMenuSubtitle = style
		case "main_menu_title_border":
			theme.MainMenuTitleBorder = style
		case "main_menu_selected_arrow":
			theme.MainMenuSelectedArrow = style
		case "main_menu_selected_item":
			theme.MainMenuSelectedItem = style
		case "main_menu_unselected_item":
			theme.MainMenuUnselectedItem = style
		case "main_menu_selected_icon":
			theme.MainMenuSelectedIcon = style
		case "main_menu_unselected_icon":
			theme.MainMenuUnselectedIcon = style
		case "main_menu_selected_title":
			theme.MainMenuSelectedTitle = style
		case "main_menu_unselected_title":
			theme.MainMenuUnselectedTitle = style
		case "main_menu_selected_desc":
			theme.MainMenuSelectedDesc = style
		case "main_menu_unselected_desc":
			theme.MainMenuUnselectedDesc = style
		case "main_menu_help_header":
			theme.MainMenuHelpHeader = style
		case "main_menu_help_content":
			theme.MainMenuHelpContent = style
		case "main_menu_help_border":
			theme.MainMenuHelpBorder = style
		case "main_menu_help_key":
			theme.MainMenuHelpKey = style
		case "main_menu_help_desc":
			theme.MainMenuHelpDesc = style
		case "main_menu_spinner":
			theme.MainMenuSpinner = style
		case "state_selector_container":
			theme.StateSelectorContainer = style
		case "state_selector_dialog":
			theme.StateSelectorDialog = style
		case "state_selector_title":
			theme.StateSelectorTitle = style
		case "state_selector_help":
			theme.StateSelectorHelp = style
		case "state_selector_item":
			theme.StateSelectorItem = style
		case "state_selector_selected_item":
			theme.StateSelectorSelectedItem = style
		case "state_selector_indicator":
			theme.StateSelectorIndicator = style
		case "state_selector_selected_indicator":
			theme.StateSelectorSelectedIndicator = style
		case "session_selector_container":
			theme.SessionSelectorContainer = style
		case "session_selector_dialog":
			theme.SessionSelectorDialog = style
		case "session_selector_title":
			theme.SessionSelectorTitle = style
		case "session_selector_pagination":
			theme.SessionSelectorPagination = style
		case "session_selector_help":
			theme.SessionSelectorHelp = style
		case "model_selector_container":
			theme.ModelSelectorContainer = style
		case "model_selector_dialog":
			theme.ModelSelectorDialog = style
		case "model_selector_title":
			theme.ModelSelectorTitle = style
		case "model_selector_pagination":
			theme.ModelSelectorPagination = style
		case "model_selector_help":
			theme.ModelSelectorHelp = style
		case "theme_selector_container":
			theme.ThemeSelectorContainer = style
		case "theme_selector_dialog":
			theme.ThemeSelectorDialog = style
		case "theme_selector_title":
			theme.ThemeSelectorTitle = style
		case "theme_selector_pagination":
			theme.ThemeSelectorPagination = style
		case "theme_selector_help":
			theme.ThemeSelectorHelp = style
		case "add_recipe_from_url_container":
			theme.AddRecipeFromURLContainer = style
		case "add_recipe_from_url_dialog":
			theme.AddRecipeFromURLDialog = style
		case "add_recipe_from_url_title":
			theme.AddRecipeFromURLTitle = style
		case "add_recipe_from_url_help":
			theme.AddRecipeFromURLHelp = style
		case "add_recipe_from_url_prompt":
			theme.AddRecipeFromURLPrompt = style
		case "add_recipe_from_url_error":
			theme.AddRecipeFromURLError = style
		case "add_recipe_from_url_separator":
			theme.AddRecipeFromURLSeparator = style
		case "add_recipe_from_url_spinner":
			theme.AddRecipeFromURLSpinner = style
		case "add_recipe_from_url_accent":
			theme.AddRecipeFromURLAccent = style
		case "add_recipe_from_url_input_border":
			theme.AddRecipeFromURLInputBorder = style
		case "add_recipe_from_url_key_highlight":
			theme.AddRecipeFromURLKeyHighlight = style
		case "recipe_selector_container":
			theme.RecipeSelectorContainer = style
		case "recipe_selector_dialog":
			theme.RecipeSelectorDialog = style
		case "recipe_selector_title":
			theme.RecipeSelectorTitle = style
		case "recipe_selector_help":
			theme.RecipeSelectorHelp = style
		case "recipe_selector_selected":
			theme.RecipeSelectorSelected = style
		case "command_palette_container":
			theme.CommandPaletteContainer = style
		case "command_palette_dialog":
			theme.CommandPaletteDialog = style
		case "command_palette_title":
			theme.CommandPaletteTitle = style
		case "command_palette_help":
			theme.CommandPaletteHelp = style
		case "command_palette_shortcut":
			theme.CommandPaletteShortcut = style
		case "command_palette_selected":
			theme.CommandPaletteSelected = style
		case "rating_bar":
			theme.RatingBar = style
		case "rating_star_active":
			theme.RatingStarActive = style
		case "rating_star_inactive":
			theme.RatingStarInactive = style
		case "rating_dialog_container":
			theme.RatingDialogContainer = style
		case "rating_dialog_box":
			theme.RatingDialogBox = style
		case "rating_dialog_title":
			theme.RatingDialogTitle = style
		case "rating_dialog_help":
			theme.RatingDialogHelp = style
		case "cooking_step_counter":
			theme.CookingStepCounter = style
		case "cooking_instruction":
			theme.CookingInstruction = style
		case "cooking_nav_hint":
			theme.CookingNavHint = style
		case "cooking_sidebar":
			theme.CookingSidebar = style
		case "cooking_sidebar_title":
			theme.CookingSidebarTitle = style
		case "cooking_ingredient":
			theme.CookingIngredient = style
		case "cooking_ingredient_amount":
			theme.CookingIngredientAmount = style
		case "cooking_ingredient_detail":
			theme.CookingIngredientDetail = style
		case "cooking_chat_panel":
			theme.CookingChatPanel = style
		case "cooking_chat_title":
			theme.CookingChatTitle = style
		case "cooking_timer_active":
			theme.CookingTimerActive = style
		case "cooking_timer_done":
			theme.CookingTimerDone = style
		case "cooking_timer_label":
			theme.CookingTimerLabel = style
		case "cooking_timer_message":
			theme.CookingTimerMessage = style
		case "cooking_timer_bar_filled":
			theme.CookingTimerBarFilled = style
		case "cooking_timer_bar_empty":
			theme.CookingTimerBarEmpty = style
		case "cooking_timer_bar_completed":
			theme.CookingTimerBarCompleted = style
		case "textarea_cursor_line":
			theme.TextareaCursorLine = style
		case "textarea_base":
			theme.TextareaBase = style
		case "textarea_placeholder":
			theme.TextareaPlaceholder = style
		case "textarea_text":
			theme.TextareaText = style
		case "textarea_prompt":
			theme.TextareaPrompt = style
		case "textarea_end_of_buffer":
			theme.TextareaEndOfBuffer = style
		case "separator_line":
			theme.SeparatorLine = style
		case "message_separator":
			theme.MessageSeparator = style
		case "dialog_selected_row":
			theme.DialogSelectedRow = style
		case "dialog_unselected_row":
			theme.DialogUnselectedRow = style
		case "session_selector_selected_desc":
			theme.SessionSelectorSelectedDesc = style
		case "session_selector_unselected_desc":
			theme.SessionSelectorUnselectedDesc = style
		case "sidebar_value":
			theme.SidebarValue = style
		case "chat_empty_state":
			theme.ChatEmptyState = style
		case "tool_call_label":
			theme.ToolCallLabel = style
		case "tool_call_input":
			theme.ToolCallInput = style
		case "chat_mention":
			theme.ChatMention = style
		case "chat_mention_popup_border":
			theme.ChatMentionPopupBorder = style
		case "chat_mention_popup_header":
			theme.ChatMentionPopupHeader = style
		case "chat_mention_popup_item":
			theme.ChatMentionPopupItem = style
		case "chat_mention_popup_selected":
			theme.ChatMentionPopupSelected = style
		case "cooking_chat_user_label":
			theme.CookingChatUserLabel = style
		case "cooking_chat_assistant_label":
			theme.CookingChatAssistantLabel = style
		case "cooking_chat_empty":
			theme.CookingChatEmpty = style
		case "cooking_no_recipe":
			theme.CookingNoRecipe = style
		case "cooking_recipe_name":
			theme.CookingRecipeName = style
		case "cooking_progress_filled":
			theme.CookingProgressFilled = style
		case "cooking_progress_unfilled":
			theme.CookingProgressUnfilled = style
		case "cooking_ingredient_highlight":
			theme.CookingIngredientHighlight = style
		case "cooking_nav_arrow":
			theme.CookingNavArrow = style
		case "cooking_help_key":
			theme.CookingHelpKey = style
		}
	}

	// Handle list styles
	if yt.Lists.TitleBar != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.TitleBar)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "title_bar", err)
		}
		theme.ListStyles.TitleBar = style
	}
	if yt.Lists.Title != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.Title)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "title", err)
		}
		theme.ListStyles.Title = style
	}
	if yt.Lists.Spinner != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.Spinner)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "spinner", err)
		}
		theme.ListStyles.Spinner = style
	}
	if yt.Lists.FilterPrompt != (YAMLStyle{}) {
		// FilterPrompt maps to Filter.Focused.Prompt and Filter.Blurred.Prompt in bubbles v2
		promptStyle, err := createStyle(yt.Lists.FilterPrompt)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "filter_prompt", err)
		}
		theme.ListStyles.Filter.Focused.Prompt = promptStyle
		theme.ListStyles.Filter.Blurred.Prompt = promptStyle
	}
	if yt.Lists.FilterCursor != (YAMLStyle{}) {
		// FilterCursor maps to Filter.Cursor.Color in bubbles v2
		if yt.Lists.FilterCursor.Foreground != "" {
			theme.ListStyles.Filter.Cursor.Color = lipgloss.Color(resolveColor(yt.Lists.FilterCursor.Foreground))
		}
	}
	if yt.Lists.DefaultFilterCharacterMatch != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.DefaultFilterCharacterMatch)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "default_filter_character_match", err)
		}
		theme.ListStyles.DefaultFilterCharacterMatch = style
	}
	if yt.Lists.StatusBar != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.StatusBar)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "status_bar", err)
		}
		theme.ListStyles.StatusBar = style
	}
	if yt.Lists.StatusEmpty != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.StatusEmpty)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "status_empty", err)
		}
		theme.ListStyles.StatusEmpty = style
	}
	if yt.Lists.StatusBarActiveFilter != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.StatusBarActiveFilter)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "status_bar_active_filter", err)
		}
		theme.ListStyles.StatusBarActiveFilter = style
	}
	if yt.Lists.StatusBarFilterCount != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.StatusBarFilterCount)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "status_bar_filter_count", err)
		}
		theme.ListStyles.StatusBarFilterCount = style
	}
	if yt.Lists.NoItems != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.NoItems)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "no_items", err)
		}
		theme.ListStyles.NoItems = style
	}
	if yt.Lists.PaginationStyle != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.PaginationStyle)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "pagination_style", err)
		}
		theme.ListStyles.PaginationStyle = style
	}
	if yt.Lists.HelpStyle != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.HelpStyle)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "help_style", err)
		}
		theme.ListStyles.HelpStyle = style
	}
	if yt.Lists.ActivePaginationDot != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.ActivePaginationDot)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "active_pagination_dot", err)
		}
		theme.ListStyles.ActivePaginationDot = style
	}
	if yt.Lists.InactivePaginationDot != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.InactivePaginationDot)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "inactive_pagination_dot", err)
		}
		theme.ListStyles.InactivePaginationDot = style
	}
	if yt.Lists.ArabicPagination != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.ArabicPagination)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "arabic_pagination", err)
		}
		theme.ListStyles.ArabicPagination = style
	}
	if yt.Lists.DividerDot != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.DividerDot)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "divider_dot", err)
		}
		theme.ListStyles.DividerDot = style
	}

	// Handle delegate styles
	if yt.Lists.NormalTitle != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.NormalTitle)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "normal_title", err)
		}
		theme.DelegateStyles.NormalTitle = style
	}
	if yt.Lists.NormalDesc != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.NormalDesc)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "normal_desc", err)
		}
		theme.DelegateStyles.NormalDesc = style
	}
	if yt.Lists.SelectedTitle != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.SelectedTitle)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "selected_title", err)
		}
		theme.DelegateStyles.SelectedTitle = style
	}
	if yt.Lists.SelectedDesc != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.SelectedDesc)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "selected_desc", err)
		}
		theme.DelegateStyles.SelectedDesc = style
	}
	if yt.Lists.DimmedTitle != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.DimmedTitle)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "dimmed_title", err)
		}
		theme.DelegateStyles.DimmedTitle = style
	}
	if yt.Lists.DimmedDesc != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.DimmedDesc)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "dimmed_desc", err)
		}
		theme.DelegateStyles.DimmedDesc = style
	}
	if yt.Lists.FilterMatch != (YAMLStyle{}) {
		style, err := createStyle(yt.Lists.FilterMatch)
		if err != nil {
			return Theme{}, fmt.Errorf("invalid list style %q: %w", "filter_match", err)
		}
		theme.DelegateStyles.FilterMatch = style
	}

	return theme, nil
}

// Helper functions for style properties
func applyPadding(style lipgloss.Style, padding string) (lipgloss.Style, error) {
	parts := strings.Split(padding, ",")
	if len(parts) == 1 {
		// Single value for all sides
		value, err := parseInt(parts[0])
		if err != nil {
			return style, err
		}
		return style.Padding(value), nil
	} else if len(parts) == 2 {
		// Vertical, horizontal (CSS shorthand order)
		vertical, err := parseInt(parts[0])
		if err != nil {
			return style, err
		}
		horizontal, err := parseInt(parts[1])
		if err != nil {
			return style, err
		}
		return style.Padding(vertical, horizontal), nil
	} else if len(parts) == 4 {
		// Top, right, bottom, left
		top, err := parseInt(parts[0])
		if err != nil {
			return style, err
		}
		right, err := parseInt(parts[1])
		if err != nil {
			return style, err
		}
		bottom, err := parseInt(parts[2])
		if err != nil {
			return style, err
		}
		left, err := parseInt(parts[3])
		if err != nil {
			return style, err
		}
		return style.Padding(top, right, bottom, left), nil
	}
	return style, fmt.Errorf("padding must have 1, 2, or 4 comma-separated values")
}

func applyMargin(style lipgloss.Style, margin string) (lipgloss.Style, error) {
	parts := strings.Split(margin, ",")
	if len(parts) == 1 {
		// Single value for all sides
		value, err := parseInt(parts[0])
		if err != nil {
			return style, err
		}
		return style.Margin(value), nil
	} else if len(parts) == 2 {
		// Vertical, horizontal (CSS shorthand order)
		vertical, err := parseInt(parts[0])
		if err != nil {
			return style, err
		}
		horizontal, err := parseInt(parts[1])
		if err != nil {
			return style, err
		}
		return style.Margin(vertical, horizontal), nil
	} else if len(parts) == 4 {
		// Top, right, bottom, left
		top, err := parseInt(parts[0])
		if err != nil {
			return style, err
		}
		right, err := parseInt(parts[1])
		if err != nil {
			return style, err
		}
		bottom, err := parseInt(parts[2])
		if err != nil {
			return style, err
		}
		left, err := parseInt(parts[3])
		if err != nil {
			return style, err
		}
		return style.Margin(top, right, bottom, left), nil
	}
	return style, fmt.Errorf("margin must have 1, 2, or 4 comma-separated values")
}

func applyBorder(style lipgloss.Style, borderType string, borderColor color.Color) (lipgloss.Style, error) {
	switch borderType {
	case "normal":
		return style.Border(lipgloss.NormalBorder()).BorderForeground(borderColor), nil
	case "rounded":
		return style.Border(lipgloss.RoundedBorder()).BorderForeground(borderColor), nil
	case "double":
		return style.Border(lipgloss.DoubleBorder()).BorderForeground(borderColor), nil
	case "thick":
		return style.Border(lipgloss.ThickBorder()).BorderForeground(borderColor), nil
	}
	return style, fmt.Errorf("unsupported border type %q", borderType)
}

func applyAlign(style lipgloss.Style, align string) (lipgloss.Style, error) {
	switch align {
	case "left":
		return style.Align(lipgloss.Left), nil
	case "center":
		return style.Align(lipgloss.Center), nil
	case "right":
		return style.Align(lipgloss.Right), nil
	}
	return style, fmt.Errorf("unsupported align value %q", align)
}

func parseInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("value cannot be empty")
	}
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q", s)
	}
	return result, nil
}

func validateYAMLTheme(yt *YAMLTheme) error {
	if strings.TrimSpace(yt.Name) == "" {
		return errors.New("theme name is required")
	}

	for styleName, style := range yt.Styles {
		if !isSupportedThemeStyle(styleName) {
			return fmt.Errorf("unknown style key %q", styleName)
		}
		if err := validateYAMLStyle(styleName, style); err != nil {
			return err
		}
	}

	listStyles := map[string]YAMLStyle{
		"title_bar":                      yt.Lists.TitleBar,
		"title":                          yt.Lists.Title,
		"spinner":                        yt.Lists.Spinner,
		"filter_prompt":                  yt.Lists.FilterPrompt,
		"filter_cursor":                  yt.Lists.FilterCursor,
		"default_filter_character_match": yt.Lists.DefaultFilterCharacterMatch,
		"status_bar":                     yt.Lists.StatusBar,
		"status_empty":                   yt.Lists.StatusEmpty,
		"status_bar_active_filter":       yt.Lists.StatusBarActiveFilter,
		"status_bar_filter_count":        yt.Lists.StatusBarFilterCount,
		"no_items":                       yt.Lists.NoItems,
		"pagination_style":               yt.Lists.PaginationStyle,
		"help_style":                     yt.Lists.HelpStyle,
		"active_pagination_dot":          yt.Lists.ActivePaginationDot,
		"inactive_pagination_dot":        yt.Lists.InactivePaginationDot,
		"arabic_pagination":              yt.Lists.ArabicPagination,
		"divider_dot":                    yt.Lists.DividerDot,
		"normal_title":                   yt.Lists.NormalTitle,
		"normal_desc":                    yt.Lists.NormalDesc,
		"selected_title":                 yt.Lists.SelectedTitle,
		"selected_desc":                  yt.Lists.SelectedDesc,
		"dimmed_title":                   yt.Lists.DimmedTitle,
		"dimmed_desc":                    yt.Lists.DimmedDesc,
		"filter_match":                   yt.Lists.FilterMatch,
	}

	for styleName, style := range listStyles {
		if style == (YAMLStyle{}) {
			continue
		}
		if err := validateYAMLStyle(styleName, style); err != nil {
			return err
		}
	}

	return nil
}

func validateYAMLStyle(name string, style YAMLStyle) error {
	if style.Padding != "" {
		if _, err := applyPadding(lipgloss.NewStyle(), style.Padding); err != nil {
			return fmt.Errorf("invalid padding for %q: %w", name, err)
		}
	}
	if style.Margin != "" {
		if _, err := applyMargin(lipgloss.NewStyle(), style.Margin); err != nil {
			return fmt.Errorf("invalid margin for %q: %w", name, err)
		}
	}
	if style.Border != "" && style.Border != "none" {
		if _, err := applyBorder(lipgloss.NewStyle(), style.Border, lipgloss.Color(style.BorderColor)); err != nil {
			return fmt.Errorf("invalid border for %q: %w", name, err)
		}
	}
	if style.Align != "" {
		if _, err := applyAlign(lipgloss.NewStyle(), style.Align); err != nil {
			return fmt.Errorf("invalid alignment for %q: %w", name, err)
		}
	}
	return nil
}

func isSupportedThemeStyle(name string) bool {
	_, ok := supportedThemeStyles[name]
	return ok
}

var supportedThemeStyles = map[string]struct{}{
	"app_background":                    {},
	"title":                             {},
	"info":                              {},
	"error":                             {},
	"header":                            {},
	"ingredient":                        {},
	"doc":                               {},
	"detail_content":                    {},
	"detail_header":                     {},
	"detail_footer":                     {},
	"scroll_bar":                        {},
	"loading":                           {},
	"instruction":                       {},
	"warning":                           {},
	"success":                           {},
	"help":                              {},
	"status_line":                       {},
	"status_line_left":                  {},
	"status_line_right":                 {},
	"status_line_mode":                  {},
	"status_line_file":                  {},
	"status_line_info":                  {},
	"status_line_separator":             {},
	"chat_title":                        {},
	"chat":                              {},
	"sidebar":                           {},
	"sidebar_header":                    {},
	"sidebar_section":                   {},
	"sidebar_content":                   {},
	"sidebar_success":                   {},
	"sidebar_error":                     {},
	"user_message":                      {},
	"user_content":                      {},
	"assistant_message":                 {},
	"assistant_content":                 {},
	"user":                              {},
	"assistant":                         {},
	"spinner":                           {},
	"main_menu_border":                  {},
	"main_menu_container":               {},
	"main_menu_separator":               {},
	"main_menu_welcome":                 {},
	"main_menu_logo":                    {},
	"main_menu_subtitle":                {},
	"main_menu_title_border":            {},
	"main_menu_selected_arrow":          {},
	"main_menu_selected_item":           {},
	"main_menu_unselected_item":         {},
	"main_menu_selected_icon":           {},
	"main_menu_unselected_icon":         {},
	"main_menu_selected_title":          {},
	"main_menu_unselected_title":        {},
	"main_menu_selected_desc":           {},
	"main_menu_unselected_desc":         {},
	"main_menu_help_header":             {},
	"main_menu_help_content":            {},
	"main_menu_help_border":             {},
	"main_menu_help_key":                {},
	"main_menu_help_desc":               {},
	"main_menu_spinner":                 {},
	"state_selector_container":          {},
	"state_selector_dialog":             {},
	"state_selector_title":              {},
	"state_selector_help":               {},
	"state_selector_item":               {},
	"state_selector_selected_item":      {},
	"state_selector_indicator":          {},
	"state_selector_selected_indicator": {},
	"session_selector_container":        {},
	"session_selector_dialog":           {},
	"session_selector_title":            {},
	"session_selector_pagination":       {},
	"session_selector_help":             {},
	"model_selector_container":          {},
	"model_selector_dialog":             {},
	"model_selector_title":              {},
	"model_selector_pagination":         {},
	"model_selector_help":               {},
	"theme_selector_container":          {},
	"theme_selector_dialog":             {},
	"theme_selector_title":              {},
	"theme_selector_pagination":         {},
	"theme_selector_help":               {},
	"add_recipe_from_url_container":     {},
	"add_recipe_from_url_dialog":        {},
	"add_recipe_from_url_title":         {},
	"add_recipe_from_url_help":          {},
	"add_recipe_from_url_prompt":        {},
	"add_recipe_from_url_error":         {},
	"add_recipe_from_url_separator":     {},
	"add_recipe_from_url_spinner":       {},
	"add_recipe_from_url_accent":        {},
	"add_recipe_from_url_input_border":  {},
	"add_recipe_from_url_key_highlight": {},
	"recipe_selector_container":         {},
	"recipe_selector_dialog":            {},
	"recipe_selector_title":             {},
	"recipe_selector_help":              {},
	"recipe_selector_selected":          {},
	"command_palette_container":         {},
	"command_palette_dialog":            {},
	"command_palette_title":             {},
	"command_palette_help":              {},
	"command_palette_shortcut":          {},
	"command_palette_selected":          {},
	"rating_bar":                        {},
	"rating_star_active":                {},
	"rating_star_inactive":              {},
	"rating_dialog_container":           {},
	"rating_dialog_box":                 {},
	"rating_dialog_title":               {},
	"rating_dialog_help":                {},
	"cooking_step_counter":              {},
	"cooking_instruction":               {},
	"cooking_nav_hint":                  {},
	"cooking_sidebar":                   {},
	"cooking_sidebar_title":             {},
	"cooking_ingredient":                {},
	"cooking_ingredient_amount":         {},
	"cooking_ingredient_detail":         {},
	"cooking_chat_panel":                {},
	"cooking_chat_title":                {},
	"cooking_timer_active":              {},
	"cooking_timer_done":                {},
	"cooking_timer_label":               {},
	"cooking_timer_message":             {},
	"cooking_timer_bar_filled":          {},
	"cooking_timer_bar_empty":           {},
	"cooking_timer_bar_completed":       {},
	"textarea_cursor_line":              {},
	"textarea_base":                     {},
	"textarea_placeholder":              {},
	"textarea_text":                     {},
	"textarea_prompt":                   {},
	"textarea_end_of_buffer":            {},
	"separator_line":                    {},
	"message_separator":                 {},
	"dialog_selected_row":               {},
	"dialog_unselected_row":             {},
	"session_selector_selected_desc":    {},
	"session_selector_unselected_desc":  {},
	"sidebar_value":                     {},
	"chat_empty_state":                  {},
	"tool_call_label":                   {},
	"tool_call_input":                   {},
	"chat_mention":                      {},
	"chat_mention_popup_border":         {},
	"chat_mention_popup_header":         {},
	"chat_mention_popup_item":           {},
	"chat_mention_popup_selected":       {},
	"cooking_chat_user_label":           {},
	"cooking_chat_assistant_label":      {},
	"cooking_chat_empty":                {},
	"cooking_no_recipe":                 {},
	"cooking_recipe_name":               {},
	"cooking_progress_filled":           {},
	"cooking_progress_unfilled":         {},
	"cooking_ingredient_highlight":      {},
	"cooking_nav_arrow":                 {},
	"cooking_help_key":                  {},
}
