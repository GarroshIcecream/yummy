package common

import (
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	tea "github.com/charmbracelet/bubbletea"
)

type TUIModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (TUIModel, tea.Cmd)
	View() string
	GetSessionState() SessionState
	GetModelState() ModelState
	GetCurrentTheme() *themes.Theme
	SetTheme(theme *themes.Theme)
	GetSize() (width, height int)
	SetSize(width, height int)
}

type SessionState int

const (
	SessionStateMainMenu SessionState = iota
	SessionStateList
	SessionStateDetail
	SessionStateEdit
	SessionStateChat
	SessionStateCooking
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
	case SessionStateCooking:
		return "🍳"
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
	case SessionStateCooking:
		return "Cooking"
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
	StatusModeCooking         StatusMode = "COOKING"
)

type ModalType string

const (
	ModalTypeStateSelector    ModalType = "STATE"
	ModalTypeSessionSelector  ModalType = "SESSION"
	ModalTypeModelSelector    ModalType = "MODEL"
	ModalTypeThemeSelector    ModalType = "THEME"
	ModalTypeAddRecipeFromURL ModalType = "ADD_RECIPE_FROM_URL"
	ModalTypeRecipeSelector   ModalType = "RECIPE_SELECTOR"
	ModalTypeCommandPalette   ModalType = "COMMAND_PALETTE"
	ModalTypeRating           ModalType = "RATING"
)
