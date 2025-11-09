package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var (
	globalConfig   *Config
	globalConfigMu sync.RWMutex
)

// Config represents the application configuration
type Config struct {
	// UI Settings
	Theme string `json:"theme"`

	// Chat Settings
	Chat ChatConfig `json:"chat"`

	// Database Settings
	Database DatabaseConfig `json:"database"`

	// Keymap Settings (can be customized)
	Keymap KeymapConfig `json:"keymap"`

	// General Settings
	General GeneralConfig `json:"general"`

	// State Selector Dialog Settings
	StateSelectorDialog StateSelectorDialogConfig `json:"state_selector_dialog"`

	// Status Line Settings
	StatusLine StatusLineConfig `json:"status_line"`

	// Main Menu Settings
	MainMenu MainMenuConfig `json:"main_menu"`

	// Detail View Settings
	Detail DetailConfig `json:"detail"`

	// List View Settings
	List ListConfig `json:"list"`

	// Session Selector Dialog Settings
	SessionSelectorDialog SessionSelectorDialogConfig `json:"session_selector_dialog"`
}

// NewDefaultConfig returns the default configuration
func NewDefaultConfig() *Config {
	return &Config{
		Theme:                 "default",
		StateSelectorDialog:   NewDefaultStateSelectorDialogConfig(),
		SessionSelectorDialog: NewDefaultSessionSelectorDialogConfig(),
		Chat:                  NewDefaultChatConfig(),
		Database:              NewDefaultDatabaseConfig(),
		Keymap:                NewDefaultKeyBindings(),
		StatusLine:            NewDefaultStatusLineConfig(),
		MainMenu:              NewDefaultMainMenuConfig(),
		Detail:                NewDefaultDetailConfig(),
		List:                  NewDefaultListConfig(),
		General:               NewDefaultGeneralConfig(),
	}
}

// DetailConfig contains detail view settings
type DetailConfig struct {
	ViewportHeight            int    `json:"viewport_height"`
	ViewportWidth             int    `json:"viewport_width"`
	ScrollSpeed               int    `json:"scroll_speed"`
	MoveSpeed                 int    `json:"move_speed"`
	NoRecipeSelectedMessage   string `json:"no_recipe_selected_message"`
	NoContentAvailableMessage string `json:"no_content_available_message"`
}

func NewDefaultDetailConfig() DetailConfig {
	return DetailConfig{
		ViewportHeight:            30,
		ViewportWidth:             80,
		ScrollSpeed:               3,
		MoveSpeed:                 1,
		NoRecipeSelectedMessage:   "📖 There is no recipe selected. Please select a recipe from the Recipe List",
		NoContentAvailableMessage: "📝 No content available",
	}
}

// MainMenuConfig contains main menu settings
type MainMenuConfig struct {
	ContentWidth         int    `json:"content_width"`
	MenuItemWidth        int    `json:"menu_item_width"`
	MainMenuContentWidth int    `json:"main_menu_content_width"`
	MainMenuWelcomeText  string `json:"main_menu_welcome_text"`
	MainMenuSubtitleText string `json:"main_menu_subtitle_text"`
	MainMenuHelpText     string `json:"main_menu_help_text"`
}

func NewDefaultMainMenuConfig() MainMenuConfig {
	return MainMenuConfig{
		ContentWidth:         0,
		MenuItemWidth:        60,
		MainMenuContentWidth: 80,
		MainMenuWelcomeText:  "🌟 Welcome to your culinary journey! Choose an option below to get started:",
		MainMenuSubtitleText: "🍳 Your Personal Culinary Companion 🍳",
		MainMenuHelpText:     "🎮 Navigation Controls",
	}
}

// StateSelectorDialogConfig contains state selector dialog settings
type StateSelectorDialogConfig struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

func NewDefaultSessionSelectorDialogConfig() SessionSelectorDialogConfig {
	return SessionSelectorDialogConfig{
		Height: 30,
		Width:  80,
	}
}

// SessionSelectorDialogConfig contains session selector dialog settings
type SessionSelectorDialogConfig struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

func NewDefaultStateSelectorDialogConfig() StateSelectorDialogConfig {
	return StateSelectorDialogConfig{
		Height: 30,
		Width:  80,
	}
}

// ChatConfig contains chat-related settings
type ChatConfig struct {
	DefaultModel             string  `json:"default_model"`
	Temperature              float64 `json:"temperature"`
	MaxTokens                int     `json:"max_tokens"`
	SystemPrompt             string  `json:"system_prompt"`
	TextAreaPlaceholder      string  `json:"text_area_placeholder"`
	TextAreaMaxChar          int     `json:"text_area_max_char"`
	UserName                 string  `json:"user_name"`
	AssistantName            string  `json:"assistant_name"`
	AssistantAvatar          string  `json:"assistant_avatar"`
	UserAvatar               string  `json:"user_avatar"`
	AssistantThinkingMessage string  `json:"assistant_thinking_message"`

	// UI Layout constants
	UILayout UILayoutConfig `json:"ui_layout"`
}

func NewDefaultChatConfig() ChatConfig {
	return ChatConfig{
		DefaultModel: "gemma3:4b",
		Temperature:  0.9,
		MaxTokens:    1000,
		SystemPrompt: `You are a helpful cooking assistant specialized in recipes, ingredients, and culinary knowledge. You have access to a personal cookbook database and can help users with various cooking-related tasks.

		Your capabilities include:
		- Finding recipes by name or ID
		- Listing all available recipes
		- Providing cooking advice and ingredient information
		- Helping with meal planning and recipe suggestions
		- Answering questions about cooking techniques and food preparation

		Available tools:
		- searchRecipeByName: Search for recipes by name (case-insensitive)
		- getRecipeById: Get a specific recipe by its unique ID
		- listAllRecipes: List all recipes in the cookbook

		Guidelines for responses:
		- Always format your responses using markdown for better readability
		- Use headers, lists, and emphasis where appropriate
		- Be helpful and encouraging when providing cooking advice
		- If you need to search for recipes, use the available tools
		- Provide detailed information about ingredients, cooking methods, and serving suggestions
		- If a recipe isn't found, suggest similar alternatives or offer to help with general cooking questions

		Remember: You are a cooking expert, so provide accurate, helpful information and be enthusiastic about food and cooking!`,

		TextAreaPlaceholder:      "Ask anything about cooking, recipes, ingredients, or anything else you want to know about food... 🍳 ",
		TextAreaMaxChar:          400,
		UserName:                 "User",
		AssistantName:            "Assistant",
		AssistantAvatar:          "🤖",
		UserAvatar:               "👤",
		AssistantThinkingMessage: "Thinking...",
		UILayout:                 NewDefaultUILayoutConfig(),
	}
}

// UILayoutConfig contains UI layout and sizing constants
type UILayoutConfig struct {
	// Padding and margins
	ContentPadding  int `json:"content_padding"`
	MarkdownPadding int `json:"markdown_padding"`

	// Minimum dimensions
	MinContentWidth             int `json:"min_content_width"`
	MinMarkdownWidth            int `json:"min_markdown_width"`
	MinViewportHeight           int `json:"min_viewport_height"`
	MinMarkdownWidthForRenderer int `json:"min_markdown_width_for_renderer"`

	// Height calculations
	TitleHeight   int `json:"title_height"`
	InputHeight   int `json:"input_height"`
	BorderPadding int `json:"border_padding"`
	TotalUIHeight int `json:"total_ui_height"`

	// Sidebar constraints
	MinSidebarWidth   int `json:"min_sidebar_width"`
	MaxSidebarWidth   int `json:"max_sidebar_width"`
	SidebarWidthRatio int `json:"sidebar_width_ratio"`

	// Viewport constraints
	ViewportHeight     int `json:"viewport_height"`
	ViewportWidth      int `json:"viewport_width"`
	SidebarWidth       int `json:"sidebar_width"`
	MinWidthForSidebar int `json:"min_width_for_sidebar"`
}

func NewDefaultUILayoutConfig() UILayoutConfig {
	return UILayoutConfig{
		ContentPadding:              8,
		MarkdownPadding:             8,
		MinContentWidth:             20,
		MinMarkdownWidth:            20,
		MinViewportHeight:           8,
		MinMarkdownWidthForRenderer: 8,
		TitleHeight:                 5,
		InputHeight:                 5,
		BorderPadding:               6,
		TotalUIHeight:               13,
		SidebarWidth:                30,
		MinWidthForSidebar:          100,
		MinSidebarWidth:             25,
		MaxSidebarWidth:             40,
		ViewportHeight:              30,
		ViewportWidth:               80,
	}
}

// DatabaseConfig contains database-related settings
type DatabaseConfig struct {
	RecipeDBName     string `json:"recipe_db_name"`
	SessionLogDBName string `json:"session_log_db_name"`
}

func NewDefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		RecipeDBName:     "cookbook.db",
		SessionLogDBName: "session_log.db",
	}
}

// KeymapConfig allows customization of key bindings
type KeymapConfig struct {
	CursorUp             []string `json:"cursor_up"`
	CursorDown           []string `json:"cursor_down"`
	Yes                  []string `json:"yes"`
	No                   []string `json:"no"`
	Add                  []string `json:"add"`
	NewSession           []string `json:"new_session"`
	Back                 []string `json:"back"`
	Delete               []string `json:"delete"`
	Quit                 []string `json:"quit"`
	Enter                []string `json:"enter"`
	Help                 []string `json:"help"`
	Edit                 []string `json:"edit"`
	StateSelector        []string `json:"state_selector"`
	SessionSelector      []string `json:"session_selector"`
	SetFavourite         []string `json:"set_favourite"`
	PrevPage             []string `json:"prev_page"`
	NextPage             []string `json:"next_page"`
	ForceQuit            []string `json:"force_quit"`
	ShowFullHelp         []string `json:"show_full_help"`
	CloseFullHelp        []string `json:"close_full_help"`
	CancelWhileFiltering []string `json:"cancel_while_filtering"`
	AcceptWhileFiltering []string `json:"accept_while_filtering"`
	GoToStart            []string `json:"go_to_start"`
	GoToEnd              []string `json:"go_to_end"`
	Filter               []string `json:"filter"`
	ClearFilter          []string `json:"clear_filter"`
	EditIngredients      []string `json:"edit_ingredients"`
	EditInstructions     []string `json:"edit_instructions"`
	EditAdd              []string `json:"edit_add"`
	EditEdit             []string `json:"edit_edit"`
	EditDelete           []string `json:"edit_delete"`
}

func NewDefaultKeyBindings() KeymapConfig {
	return KeymapConfig{
		CursorUp:             []string{"k", "up"},
		CursorDown:           []string{"j", "down"},
		Yes:                  []string{"y"},
		No:                   []string{"n"},
		Add:                  []string{"ctrl+a"},
		NewSession:           []string{"ctrl+a"},
		Back:                 []string{"esc", "q"},
		Delete:               []string{"ctrl+x"},
		Quit:                 []string{"q", "esc"},
		Enter:                []string{"enter"},
		Help:                 []string{"h", "?"},
		Edit:                 []string{"ctrl+e"},
		StateSelector:        []string{"ctrl+s"},
		SessionSelector:      []string{"ctrl+n"},
		SetFavourite:         []string{"ctrl+f"},
		PrevPage:             []string{"j", "left"},
		NextPage:             []string{"k", "right"},
		ForceQuit:            []string{"ctrl+c"},
		ShowFullHelp:         []string{"?"},
		CloseFullHelp:        []string{"?"},
		CancelWhileFiltering: []string{"esc"},
		AcceptWhileFiltering: []string{"enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"},
		GoToStart:            []string{"home", "g"},
		GoToEnd:              []string{"end", "G"},
		Filter:               []string{"/"},
		ClearFilter:          []string{"esc"},
		EditIngredients:      []string{"i"},
		EditInstructions:     []string{"s"},
		EditAdd:              []string{"a"},
		EditEdit:             []string{"e"},
		EditDelete:           []string{"d"},
	}
}

func (k KeymapConfig) ToKeyMap() KeyMap {
	return NewKeyMapFromConfig(k)
}

// GeneralConfig contains general application settings
type GeneralConfig struct {
	Height       int `json:"status_line_height"`
	Padding      int `json:"status_line_padding"`
	ContentWidth int `json:"status_line_content_width"`
	ScrollSpeed  int `json:"scroll_speed"`
	MoveSpeed    int `json:"move_speed"`
}

func NewDefaultGeneralConfig() GeneralConfig {
	return GeneralConfig{
		Height:       0,
		Padding:      0,
		ContentWidth: 0,
		ScrollSpeed:  3,
		MoveSpeed:    1,
	}
}

// StatusLineConfig contains status line settings
type StatusLineConfig struct {
	Height       int `json:"height"`
	Padding      int `json:"padding"`
	ContentWidth int `json:"content_width"`
}

func NewDefaultStatusLineConfig() StatusLineConfig {
	return StatusLineConfig{
		Height:       1,
		Padding:      2,
		ContentWidth: 80,
	}
}

// ListConfig contains list view settings
type ListConfig struct {
	ViewStatusMessageTTL              int    `json:"list_view_status_message_ttl"`
	ViewStatusMessageFavouriteSet     string `json:"list_view_status_message_favourite_set"`
	ViewStatusMessageFavouriteRemoved string `json:"list_view_status_message_favourite_removed"`
	ViewStatusMessageRecipeDeleted    string `json:"list_view_status_message_recipe_deleted"`
	Title                             string `json:"list_title"`
	ItemNameSingular                  string `json:"list_item_name_singular"`
	ItemNamePlural                    string `json:"list_item_name_plural"`
}

func NewDefaultListConfig() ListConfig {
	return ListConfig{
		ViewStatusMessageTTL:              1500,
		ViewStatusMessageFavouriteSet:     " ⭐️ Favourite set!",
		ViewStatusMessageFavouriteRemoved: " ❌ Favourite removed!",
		ViewStatusMessageRecipeDeleted:    " ❌ Recipe deleted!",
		Title:                             "📚 My Cookbook",
		ItemNameSingular:                  "recipe",
		ItemNamePlural:                    "recipes",
	}
}

// LoadConfig loads configuration from the .yummy directory
func LoadConfig(configDir string) (*Config, error) {
	configPath := filepath.Join(configDir, "config.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := NewDefaultConfig()
		if err := config.Save(configDir); err != nil {
			slog.Error("Failed to create default config", "error", err)
			return nil, err
		}
		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("Failed to read config file", "error", err)
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		slog.Error("Failed to parse config file", "error", err)
		return nil, err
	}

	return &config, nil
}

func (c *Config) Save(configDir string) error {
	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal config", "error", err)
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		slog.Error("Failed to write config file", "error", err)
		return err
	}

	return nil
}

// SetGlobalConfig sets the global configuration
func SetGlobalConfig(cfg *Config) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	globalConfig = cfg
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig
}

// GetChatConfig returns the global chat configuration
func GetChatConfig() ChatConfig {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return NewDefaultConfig().Chat
	}
	return cfg.Chat
}

// GetListConfig returns the global list configuration
func GetListConfig() *ListConfig {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return &NewDefaultConfig().List
	}
	return &cfg.List
}

// GetDetailConfig returns the global detail configuration
func GetDetailConfig() *DetailConfig {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return &NewDefaultConfig().Detail
	}
	return &cfg.Detail
}

// GetMainMenuConfig returns the global main menu configuration
func GetMainMenuConfig() *MainMenuConfig {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return &NewDefaultConfig().MainMenu
	}
	return &cfg.MainMenu
}

// GetStatusLineConfig returns the global status line configuration
func GetStatusLineConfig() *StatusLineConfig {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return &NewDefaultConfig().StatusLine
	}
	return &cfg.StatusLine
}

// GetGeneralConfig returns the global general configuration
func GetGeneralConfig() *GeneralConfig {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return &NewDefaultConfig().General
	}
	return &cfg.General
}

// GetStateSelectorDialogConfig returns the global state selector dialog configuration
func GetStateSelectorDialogConfig() *StateSelectorDialogConfig {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return &NewDefaultConfig().StateSelectorDialog
	}
	return &cfg.StateSelectorDialog
}

// GetSessionSelectorDialogConfig returns the global session selector dialog configuration
func GetSessionSelectorDialogConfig() *SessionSelectorDialogConfig {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return &NewDefaultConfig().SessionSelectorDialog
	}
	return &cfg.SessionSelectorDialog
}

func GetKeymapConfig() *KeymapConfig {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return &NewDefaultConfig().Keymap
	}
	return &cfg.Keymap
}
