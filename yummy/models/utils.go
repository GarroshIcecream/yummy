package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ModelType string

const (
	ModelTypeList   ModelType = "list"
	ModelTypeEdit   ModelType = "edit"
	ModelTypeChat   ModelType = "chat"
	ModelTypeDetail ModelType = "detail"
)

type YummyModel interface {
	Update(msg tea.Msg) (YummyModel, tea.Cmd)
	Init() tea.Cmd
	View() string
	GetName() ModelType
}
