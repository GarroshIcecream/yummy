package common

import (
	tea "github.com/charmbracelet/bubbletea"
)

type TUIModel interface {
	tea.Model
	GetSessionState() SessionState
	GetModelState() ModelState
	SetSize(width, height int)
	GetSize() (width, height int)
}

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
