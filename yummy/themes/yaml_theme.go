package themes

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// YAMLTheme represents a theme defined in YAML format
type YAMLTheme struct {
	Name        string               `yaml:"name"`
	Description string               `yaml:"description,omitempty"`
	Colors      map[string]string    `yaml:"colors"`
	Styles      map[string]YAMLStyle `yaml:"styles"`
	Lists       YAMLListStyles       `yaml:"lists,omitempty"`
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

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") || !strings.HasSuffix(entry.Name(), ".yml") {
			slog.Info("Skipping entry", "name", entry.Name())
			continue
		}

		themePath := filepath.Join(dir, entry.Name())
		theme, err := LoadThemeFromYAML(themePath)
		if err != nil {
			slog.Error("Failed to load theme", "path", themePath, "error", err)
			continue
		}

		themes = append(themes, *theme)
	}

	return themes, nil
}

// ToTheme converts a YAMLTheme to a Theme
func (yt *YAMLTheme) ToTheme() (Theme, error) {
	theme := Theme{
		Name: yt.Name,
	}

	// Helper function to resolve color references
	resolveColor := func(color string) lipgloss.Color {
		if color == "" {
			return lipgloss.Color("")
		}
		// Check if it's a color reference
		if resolved, exists := yt.Colors[color]; exists {
			return lipgloss.Color(resolved)
		}
		return lipgloss.Color(color)
	}

	// Helper function to create a style from YAMLStyle
	createStyle := func(ys YAMLStyle) lipgloss.Style {
		style := lipgloss.NewStyle()

		if ys.Foreground != "" {
			style = style.Foreground(resolveColor(ys.Foreground))
		}
		if ys.Background != "" {
			style = style.Background(resolveColor(ys.Background))
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
			style = applyPadding(style, ys.Padding)
		}

		// Handle margin
		if ys.Margin != "" {
			style = applyMargin(style, ys.Margin)
		}

		// Handle border
		if ys.Border != "" && ys.Border != "none" {
			style = applyBorder(style, ys.Border, resolveColor(ys.BorderColor))
		}

		// Handle alignment
		if ys.Align != "" {
			style = applyAlign(style, ys.Align)
		}

		if ys.Width > 0 {
			style = style.Width(ys.Width)
		}
		if ys.Height > 0 {
			style = style.Height(ys.Height)
		}

		return style
	}

	// Convert all styles
	for styleName, yamlStyle := range yt.Styles {
		style := createStyle(yamlStyle)

		// Use reflection or a switch statement to set the appropriate field
		switch styleName {
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
		}
	}

	// Handle list styles
	if yt.Lists.TitleBar != (YAMLStyle{}) {
		theme.ListStyles.TitleBar = createStyle(yt.Lists.TitleBar)
	}
	if yt.Lists.Title != (YAMLStyle{}) {
		theme.ListStyles.Title = createStyle(yt.Lists.Title)
	}
	if yt.Lists.Spinner != (YAMLStyle{}) {
		theme.ListStyles.Spinner = createStyle(yt.Lists.Spinner)
	}
	if yt.Lists.FilterPrompt != (YAMLStyle{}) {
		theme.ListStyles.FilterPrompt = createStyle(yt.Lists.FilterPrompt)
	}
	if yt.Lists.FilterCursor != (YAMLStyle{}) {
		theme.ListStyles.FilterCursor = createStyle(yt.Lists.FilterCursor)
	}
	if yt.Lists.DefaultFilterCharacterMatch != (YAMLStyle{}) {
		theme.ListStyles.DefaultFilterCharacterMatch = createStyle(yt.Lists.DefaultFilterCharacterMatch)
	}
	if yt.Lists.StatusBar != (YAMLStyle{}) {
		theme.ListStyles.StatusBar = createStyle(yt.Lists.StatusBar)
	}
	if yt.Lists.StatusEmpty != (YAMLStyle{}) {
		theme.ListStyles.StatusEmpty = createStyle(yt.Lists.StatusEmpty)
	}
	if yt.Lists.StatusBarActiveFilter != (YAMLStyle{}) {
		theme.ListStyles.StatusBarActiveFilter = createStyle(yt.Lists.StatusBarActiveFilter)
	}
	if yt.Lists.StatusBarFilterCount != (YAMLStyle{}) {
		theme.ListStyles.StatusBarFilterCount = createStyle(yt.Lists.StatusBarFilterCount)
	}
	if yt.Lists.NoItems != (YAMLStyle{}) {
		theme.ListStyles.NoItems = createStyle(yt.Lists.NoItems)
	}
	if yt.Lists.PaginationStyle != (YAMLStyle{}) {
		theme.ListStyles.PaginationStyle = createStyle(yt.Lists.PaginationStyle)
	}
	if yt.Lists.HelpStyle != (YAMLStyle{}) {
		theme.ListStyles.HelpStyle = createStyle(yt.Lists.HelpStyle)
	}
	if yt.Lists.ActivePaginationDot != (YAMLStyle{}) {
		theme.ListStyles.ActivePaginationDot = createStyle(yt.Lists.ActivePaginationDot)
	}
	if yt.Lists.InactivePaginationDot != (YAMLStyle{}) {
		theme.ListStyles.InactivePaginationDot = createStyle(yt.Lists.InactivePaginationDot)
	}
	if yt.Lists.ArabicPagination != (YAMLStyle{}) {
		theme.ListStyles.ArabicPagination = createStyle(yt.Lists.ArabicPagination)
	}
	if yt.Lists.DividerDot != (YAMLStyle{}) {
		theme.ListStyles.DividerDot = createStyle(yt.Lists.DividerDot)
	}

	// Handle delegate styles
	if yt.Lists.NormalTitle != (YAMLStyle{}) {
		theme.DelegateStyles.NormalTitle = createStyle(yt.Lists.NormalTitle)
	}
	if yt.Lists.NormalDesc != (YAMLStyle{}) {
		theme.DelegateStyles.NormalDesc = createStyle(yt.Lists.NormalDesc)
	}
	if yt.Lists.SelectedTitle != (YAMLStyle{}) {
		theme.DelegateStyles.SelectedTitle = createStyle(yt.Lists.SelectedTitle)
	}
	if yt.Lists.SelectedDesc != (YAMLStyle{}) {
		theme.DelegateStyles.SelectedDesc = createStyle(yt.Lists.SelectedDesc)
	}
	if yt.Lists.DimmedTitle != (YAMLStyle{}) {
		theme.DelegateStyles.DimmedTitle = createStyle(yt.Lists.DimmedTitle)
	}
	if yt.Lists.DimmedDesc != (YAMLStyle{}) {
		theme.DelegateStyles.DimmedDesc = createStyle(yt.Lists.DimmedDesc)
	}
	if yt.Lists.FilterMatch != (YAMLStyle{}) {
		theme.DelegateStyles.FilterMatch = createStyle(yt.Lists.FilterMatch)
	}

	return theme, nil
}

// Helper functions for style properties
func applyPadding(style lipgloss.Style, padding string) lipgloss.Style {
	parts := strings.Split(padding, ",")
	if len(parts) == 1 {
		// Single value for all sides
		return style.Padding(parseInt(parts[0]))
	} else if len(parts) == 2 {
		// Horizontal, vertical
		return style.Padding(parseInt(parts[1]), parseInt(parts[0]))
	} else if len(parts) == 4 {
		// Top, right, bottom, left
		return style.Padding(parseInt(parts[0]), parseInt(parts[1]), parseInt(parts[2]), parseInt(parts[3]))
	}
	return style
}

func applyMargin(style lipgloss.Style, margin string) lipgloss.Style {
	parts := strings.Split(margin, ",")
	if len(parts) == 1 {
		// Single value for all sides
		return style.Margin(parseInt(parts[0]))
	} else if len(parts) == 2 {
		// Horizontal, vertical
		return style.Margin(parseInt(parts[1]), parseInt(parts[0]))
	} else if len(parts) == 4 {
		// Top, right, bottom, left
		return style.Margin(parseInt(parts[0]), parseInt(parts[1]), parseInt(parts[2]), parseInt(parts[3]))
	}
	return style
}

func applyBorder(style lipgloss.Style, borderType string, borderColor lipgloss.Color) lipgloss.Style {
	switch borderType {
	case "normal":
		return style.Border(lipgloss.NormalBorder()).BorderForeground(borderColor)
	case "rounded":
		return style.Border(lipgloss.RoundedBorder()).BorderForeground(borderColor)
	case "double":
		return style.Border(lipgloss.DoubleBorder()).BorderForeground(borderColor)
	case "thick":
		return style.Border(lipgloss.ThickBorder()).BorderForeground(borderColor)
	}
	return style
}

func applyAlign(style lipgloss.Style, align string) lipgloss.Style {
	switch align {
	case "left":
		return style.Align(lipgloss.Left)
	case "center":
		return style.Align(lipgloss.Center)
	case "right":
		return style.Align(lipgloss.Right)
	}
	return style
}

func parseInt(s string) int {
	// Simple integer parsing - in a real implementation you might want more robust parsing
	if s == "" {
		return 0
	}
	// This is a simplified version - you'd want proper error handling
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		// If parsing fails, return 0 as default
		return 0
	}
	return result
}
