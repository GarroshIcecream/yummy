package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

	// Model Selector Dialog Settings
	ModelSelectorDialog ModelSelectorDialogConfig `json:"model_selector_dialog"`

	// Theme Selector Dialog Settings
	ThemeSelectorDialog ThemeSelectorDialogConfig `json:"theme_selector_dialog"`

	// Add Recipe From URL Dialog Settings
	AddRecipeFromURLDialog AddRecipeFromURLDialogConfig `json:"add_recipe_from_url_dialog"`

	// Recipe Selector Dialog Settings
	RecipeSelectorDialog RecipeSelectorDialogConfig `json:"recipe_selector_dialog"`

	// Command Palette Dialog Settings
	CommandPaletteDialog CommandPaletteDialogConfig `json:"command_palette_dialog"`
}

// NewDefaultConfig returns the default configuration
func NewDefaultConfig() *Config {
	return &Config{
		Theme:                  "default",
		StateSelectorDialog:    NewDefaultStateSelectorDialogConfig(),
		SessionSelectorDialog:  NewDefaultSessionSelectorDialogConfig(),
		ModelSelectorDialog:    NewDefaultModelSelectorDialogConfig(),
		ThemeSelectorDialog:    NewDefaultThemeSelectorDialogConfig(),
		AddRecipeFromURLDialog: NewDefaultAddRecipeFromURLDialogConfig(),
		RecipeSelectorDialog:   NewDefaultRecipeSelectorDialogConfig(),
		CommandPaletteDialog:   NewDefaultCommandPaletteDialogConfig(),
		Chat:                   NewDefaultChatConfig(),
		Database:               NewDefaultDatabaseConfig(),
		Keymap:                 NewDefaultKeyBindings(),
		StatusLine:             NewDefaultStatusLineConfig(),
		MainMenu:               NewDefaultMainMenuConfig(),
		Detail:                 NewDefaultDetailConfig(),
		List:                   NewDefaultListConfig(),
		General:                NewDefaultGeneralConfig(),
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
		MenuItemWidth:        56,
		MainMenuContentWidth: 58,
		MainMenuWelcomeText:  "Your personal recipe manager",
		MainMenuSubtitleText: "",
		MainMenuHelpText:     "",
	}
}

// SessionSelectorDialogConfig contains session selector dialog settings
type SessionSelectorDialogConfig struct {
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
type ModelSelectorDialogConfig struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

func NewDefaultModelSelectorDialogConfig() ModelSelectorDialogConfig {
	return ModelSelectorDialogConfig{
		Height: 30,
		Width:  80,
	}
}

// ThemeSelectorDialogConfig contains theme selector dialog settings
type ThemeSelectorDialogConfig struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

func NewDefaultThemeSelectorDialogConfig() ThemeSelectorDialogConfig {
	return ThemeSelectorDialogConfig{
		Height: 30,
		Width:  80,
	}
}

// StateSelectorDialogConfig contains state selector dialog settings
type StateSelectorDialogConfig struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

func NewDefaultStateSelectorDialogConfig() StateSelectorDialogConfig {
	return StateSelectorDialogConfig{
		Height: 30,
		Width:  80,
	}
}

// AddRecipeFromURLDialogConfig contains add recipe from URL dialog settings
type AddRecipeFromURLDialogConfig struct {
	Height             int    `json:"height"`
	Width              int    `json:"width"`
	PythonPath         string `json:"python_path"`          // optional; path to Python for recipe-scrapers (e.g. "python3" or venv path). Empty = auto-detect (python3 then python).
	LLMIngredientModel string `json:"llm_ingredient_model"` // Ollama model for ingredient parsing (empty = use chat default)
	// AIParsing uses the LLM when importing from a URL: structured ingredient
	// parsing, ingredient group labels and categories for uncategorised
	// recipes. Off = plain rule-based parsing only. Toggle from the command
	// palette (ctrl+p).
	AIParsing bool `json:"ai_parsing"`
}

func NewDefaultAddRecipeFromURLDialogConfig() AddRecipeFromURLDialogConfig {
	return AddRecipeFromURLDialogConfig{
		Height:             12,
		Width:              60,
		PythonPath:         "",
		LLMIngredientModel: "",
		AIParsing:          true,
	}
}

// RecipeSelectorDialogConfig contains recipe selector dialog settings
type RecipeSelectorDialogConfig struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

func NewDefaultRecipeSelectorDialogConfig() RecipeSelectorDialogConfig {
	return RecipeSelectorDialogConfig{
		Height: 30,
		Width:  60,
	}
}

// CommandPaletteDialogConfig contains command palette dialog settings
type CommandPaletteDialogConfig struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

func NewDefaultCommandPaletteDialogConfig() CommandPaletteDialogConfig {
	return CommandPaletteDialogConfig{
		Height: 15,
		Width:  50,
	}
}

// ChatConfig contains chat-related settings
type ChatConfig struct {
	DefaultModel             string  `json:"default_model"`
	Temperature              float64 `json:"temperature"`
	MaxTokens                int     `json:"max_tokens"`
	MaxIterations            int     `json:"max_iterations"`
	SystemPrompt             string  `json:"system_prompt"`
	SummaryPrompt            string  `json:"summary_prompt"`
	SummaryMaxLength         int     `json:"summary_max_length"`
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
		DefaultModel:             "gemma3:4b",
		Temperature:              0.9,
		MaxTokens:                1000,
		MaxIterations:            15,
		SystemPrompt:             DefaultSystemPrompt,
		SummaryPrompt:            `Extract 3-5 key words or short phrases (separated by commas) that best describe this cooking conversation. Focus on the main topics, recipes, or ingredients discussed. Do not use full sentences, only keywords. Conversation: %s`,
		SummaryMaxLength:         60,
		TextAreaPlaceholder:      "Ask about cooking, recipes, ingredients...",
		TextAreaMaxChar:          400,
		UserName:                 "You",
		AssistantName:            "Yummy",
		AssistantAvatar:          "",
		UserAvatar:               "",
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
	Quit                 []string `json:"quit"`
	CursorUp             []string `json:"cursor_up"`
	CursorDown           []string `json:"cursor_down"`
	Yes                  []string `json:"yes"`
	No                   []string `json:"no"`
	Add                  []string `json:"add"`
	NewSession           []string `json:"new_session"`
	Back                 []string `json:"back"`
	Delete               []string `json:"delete"`
	Enter                []string `json:"enter"`
	Help                 []string `json:"help"`
	Edit                 []string `json:"edit"`
	StateSelector        []string `json:"state_selector"`
	SessionSelector      []string `json:"session_selector"`
	ModelSelector        []string `json:"model_selector"`
	ThemeSelector        []string `json:"theme_selector"`
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
	RecipeSelector       []string `json:"recipe_selector"`
	CommandPalette       []string `json:"command_palette"`
	SetRating            []string `json:"set_rating"`
	CookingMode          []string `json:"cooking_mode"`
	ToggleIngredients    []string `json:"toggle_ingredients"`
	ToggleChat           []string `json:"toggle_chat"`
	ToggleTimer          []string `json:"toggle_timer"`
	ResetTimer           []string `json:"reset_timer"`
	ChatScrollUp         []string `json:"chat_scroll_up"`
	ChatScrollDown       []string `json:"chat_scroll_down"`
	ToggleSidebar        []string `json:"toggle_sidebar"`
}

func NewDefaultKeyBindings() KeymapConfig {
	return KeymapConfig{
		Enter:                []string{"enter"},
		Yes:                  []string{"y"},
		No:                   []string{"n"},
		Back:                 []string{"esc"},
		Add:                  []string{"ctrl+a"},
		NewSession:           []string{"ctrl+n"},
		Delete:               []string{"ctrl+x"},
		Edit:                 []string{"ctrl+e"},
		StateSelector:        []string{"ctrl+s"},
		SessionSelector:      []string{"ctrl+o"},
		ModelSelector:        []string{"ctrl+m"},
		ThemeSelector:        []string{"ctrl+t"},
		SetFavourite:         []string{"ctrl+f"},
		ForceQuit:            []string{"ctrl+c"},
		Quit:                 []string{"q"},
		CursorUp:             []string{"k", "up"},
		NextPage:             []string{"right", "pgdown"},
		CursorDown:           []string{"j", "down"},
		PrevPage:             []string{"left", "pgup"},
		ShowFullHelp:         []string{"?"},
		CloseFullHelp:        []string{"?"},
		Help:                 []string{"h", "?"},
		ClearFilter:          []string{"esc"},
		CancelWhileFiltering: []string{"esc"},
		AcceptWhileFiltering: []string{"enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"},
		GoToStart:            []string{"home", "g"},
		GoToEnd:              []string{"end", "G"},
		Filter:               []string{"/"},
		EditIngredients:      []string{"i"},
		EditInstructions:     []string{"s"},
		EditAdd:              []string{"a"},
		EditEdit:             []string{"e"},
		EditDelete:           []string{"d"},
		RecipeSelector:       []string{"ctrl+r"},
		CommandPalette:       []string{"ctrl+p"},
		SetRating:            []string{"r"},
		CookingMode:          []string{"c"},
		ToggleIngredients:    []string{"i"},
		ToggleChat:           []string{"a"},
		ToggleTimer:          []string{"space"},
		ResetTimer:           []string{"r"},
		ChatScrollUp:         []string{"ctrl+u"},
		ChatScrollDown:       []string{"ctrl+d"},
		ToggleSidebar:        []string{"ctrl+l"},
	}
}

func (k KeymapConfig) ToKeyMap() KeyMap {
	return NewKeyMapFromConfig(k)
}

// GeneralConfig contains general application settings
type GeneralConfig struct {
	ScrollSpeed int `json:"scroll_speed"`
	MoveSpeed   int `json:"move_speed"`
	// Mouse enables mouse capture (wheel scrolling). It is off by default
	// because a terminal app that captures the mouse disables the terminal's
	// own text selection and link clicking.
	Mouse bool `json:"mouse"`
}

func NewDefaultGeneralConfig() GeneralConfig {
	return GeneralConfig{
		ScrollSpeed: 3,
		MoveSpeed:   1,
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
	ViewStatusMessageRecipeAdded      string `json:"list_view_status_message_recipe_added"`
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
		ViewStatusMessageRecipeAdded:      " ✅ Recipe added!",
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

	// Load existing config on top of defaults so new fields get sensible values
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("Failed to read config file", "error", err)
		return nil, err
	}

	config := NewDefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		slog.Error("Failed to parse config file", "error", err)
		return nil, err
	}
	config.migrate()

	return config, nil
}

// DefaultSystemPrompt is the assistant's default system prompt. The tools are
// described to the model by the agent itself, so they are deliberately not
// listed here (a hard-coded list went stale as tools were added).
const DefaultSystemPrompt = `You are a helpful cooking assistant specialized in recipes, ingredients, and culinary knowledge. You have access to the user's personal cookbook through tools.

Use the tools whenever the question is about the user's own recipes:
- to find a specific dish, search by name; to browse or filter (category, time, rating, favourites, ingredients) list recipes; to answer "what can I cook with ...", find recipes by ingredients; to learn which categories exist, get the cookbook overview.
- tool results contain recipe IDs; fetch the full recipe by ID before quoting ingredients, amounts or steps.
- when a list says more pages are available and the user wants more, request the next page.
- never invent recipes, IDs, amounts or times that a tool did not return. If nothing matches, say so and suggest alternatives.

When the user references a recipe with @[RecipeName], the full recipe data is already provided in the message context. Do not search for it again; use the provided data directly.

Answer in markdown, concisely: short paragraphs and lists, emphasis where it helps. Be accurate, practical and encouraging.`

// migrate replaces values saved by older versions that are known to be broken.
func (c *Config) migrate() {
	defaults := NewDefaultKeyBindings()
	// Older defaults bound next/prev page to k/j, which collide with cursor
	// up/down (k and j) in every list and made cooking mode step backwards on j.
	if equalStrings(c.Keymap.NextPage, []string{"k", "right"}) {
		c.Keymap.NextPage = defaults.NextPage
	}
	if equalStrings(c.Keymap.PrevPage, []string{"j", "left"}) {
		c.Keymap.PrevPage = defaults.PrevPage
	}
	// The old default system prompt hard-coded a list of only two tools, which
	// kept the model from using the newer ones. Replace it when unmodified.
	if strings.Contains(c.Chat.SystemPrompt, "- getRecipeById: Get a specific recipe by its unique ID. Use this after finding a recipe with searchRecipeByName") {
		c.Chat.SystemPrompt = DefaultSystemPrompt
	}

	// Chat keys were remapped: ^n new chat, ^o sessions, ^m model, ^l sidebar.
	// Only migrate configs still holding the complete old default set, so a
	// user's own customisation is never overwritten.
	if equalStrings(c.Keymap.NewSession, []string{"ctrl+a"}) &&
		equalStrings(c.Keymap.SessionSelector, []string{"ctrl+n"}) &&
		equalStrings(c.Keymap.ModelSelector, []string{"ctrl+l"}) {
		c.Keymap.NewSession = defaults.NewSession
		c.Keymap.SessionSelector = defaults.SessionSelector
		c.Keymap.ModelSelector = defaults.ModelSelector
		if len(c.Keymap.ToggleSidebar) == 0 || equalStrings(c.Keymap.ToggleSidebar, []string{"ctrl+b"}) {
			c.Keymap.ToggleSidebar = defaults.ToggleSidebar
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
