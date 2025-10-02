package utils

import "time"

type SessionState int

const (
	SessionStateMainMenu SessionState = iota
	SessionStateList
	SessionStateDetail
	SessionStateEdit
	SessionStateChat
	SessionStateStateSelector
	SessionStateSessionSelector
)

type StateNames string

const (
	StateNameMainMenu StateNames = "Main Menu"
	StateNameList     StateNames = "Recipe List"
	StateNameDetail   StateNames = "Recipe Detail"
	StateNameEdit     StateNames = "Edit Recipe"
	StateNameChat     StateNames = "Chat Assistant"
)

type ModelState int

const (
	ModelStateLoading ModelState = iota
	ModelStateLoaded
	ModelStateError
)

type StatusMode string

const (
	StatusModeMenu            StatusMode = "MENU"
	StatusModeList            StatusMode = "COOKBOOK"
	StatusModeEdit            StatusMode = "EDIT"
	StatusModeChat            StatusMode = "CHAT"
	StatusModeRecipe          StatusMode = "RECIPE"
	StatusModeStateSelector   StatusMode = "STATE"
	StatusModeSessionSelector StatusMode = "SESSION"
)

// Mein menu constants
const (
	MenuItemWidth        = 60
	MainMenuContentWidth = 80
)

const (
	StatusLineHeight  = 1
	StatusLinePadding = 2
)

// List view constants
const (
	ListViewStatusMessageTTL              = 1500 * time.Millisecond
	ListViewStatusMessageFavouriteSet     = " ⭐️ Favourite set!"
	ListViewStatusMessageFavouriteRemoved = " ❌ Favourite removed!"
	ListViewStatusMessageRecipeDeleted    = " ❌ Recipe deleted!"
	ListTitle                             = "📚 My Cookbook"
	ListItemNameSingular                  = "recipe"
	ListItemNamePlural                    = "recipes"
)

// Chat Model constants
const (
	DefaultViewportHeight = 30
	DefaultViewportWidth  = 80
	DefaultScrollSpeed    = 3
	DefaultMoveSpeed      = 1
	TextAreaPlaceholder   = "Ask anything about cooking, recipes, ingredients, or anything else you want to know about food... 🍳 "
	TextAreaMaxChar       = 400
	TextAreaHeight        = 3
	SidebarWidth          = 30
	MinWidthForSidebar    = 100
	DefaultModel          = "gemma3:4b"
	Temperature           = 0.9
	EmptyResponse         = "Oops! Something went wrong. Please try again later"
	WelcomeMessage        = `Hello! I'm your cooking assistant. I can help you find recipes, ingredients, and anything else you need to know about food.`
	SystemPrompt          = `
	You are a cooking assistant with web scraping capabilities. You will be given questions about cooking, recipes and ingredients. 
	You can scrape web content to find relevant information when needed.
	
	IMPORTANT: You have access to a function called "scrape_website" that can scrape content from websites. 
	When a user asks about a recipe from a specific website or provides a URL, you MUST use this function.
	
	The scrape_website function takes a URL parameter and returns the scraped content from that website.
	
	When you need to scrape a website:
	1. Use the scrape_website function with the provided URL
	2. The system will scrape the content and return it to you
	3. You can then analyze the content and provide helpful information about the recipe
	
	You will also be given extracted recipes and ingredients. You will need to answer the question based on the information provided.
	Please format your responses using markdown for better readability, including headers, lists, and emphasis where appropriate.
	
	If a user provides a URL or asks about a specific recipe website, you MUST use the scrape_website function to gather information.
	
	Remember: If the user is asking about a specific recipe with no reference to a website, do NOT use the scraping tool.
	
	Available functions:
	- scrape_website(url: string): Scrapes content from a given URL
	`
)

// Ollama help messages
const (
	OllamaNotInstalledHelp = `ollama is not installed or not found in PATH.

To fix this:
1. Install Ollama from https://ollama.ai
2. Make sure Ollama is added to your system PATH
3. Restart your terminal/command prompt
4. Try running this application again

For more help, visit: https://ollama.ai/install`

	OllamaServiceNotRunningHelp = `ollama service is not running and could not be started automatically.

To fix this:
1. Start the Ollama service manually by running: ollama serve
2. Or restart your computer if Ollama is set to start automatically
3. Make sure no firewall is blocking Ollama
4. Check if there are any error messages in the Ollama logs
5. Try running this application again`
)
