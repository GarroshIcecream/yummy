package consts

import "time"

type SessionState int

const (
	SessionStateMainMenu SessionState = iota
	SessionStateList
	SessionStateDetail
	SessionStateEdit
	SessionStateChat
)

func (s SessionState) GetStateEmoji() string {
	switch s {
	case SessionStateMainMenu:
		return "🏠"
	case SessionStateList:
		return "📝"
	case SessionStateDetail:
		return "🔍"
	case SessionStateEdit:
		return "📝"
	case SessionStateChat:
		return "💬"
	default:
		return "❌"
	}
}

func (s SessionState) GetStateName() string {
	switch s {
	case SessionStateMainMenu:
		return "Main Menu"
	case SessionStateList:
		return "Recipe List"
	case SessionStateDetail:
		return "Recipe Detail"
	case SessionStateEdit:
		return "Edit Recipe"
	case SessionStateChat:
		return "Chat Assistant"
	default:
		return "Unknown State"
	}
}

type FilterField string

const (
	AuthorField      FilterField = "author"
	CategoryField    FilterField = "categories"
	IngredientsField FilterField = "ingredients"
	FavouriteField   FilterField = "favourite"
	TitleField       FilterField = "title"
	DescriptionField FilterField = "description"
	URLField         FilterField = "url"
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
	MainMenuLogoText = `
    ██╗   ██╗██╗   ██╗███╗   ███╗███╗   ███╗██╗   ██╗
    ╚██╗ ██╔╝██║   ██║████╗ ████║████╗ ████║╚██╗ ██╔╝
     ╚████╔╝ ██║   ██║██╔████╔██║██╔████╔██║ ╚████╔╝
      ╚██╔╝  ██║   ██║██║╚██╔╝██║██║╚██╔╝██║  ╚██╔╝
       ██║   ╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║   ██║
       ╚═╝    ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝   ╚═╝`
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
