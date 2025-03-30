package models

import (
	"log"

	db "github.com/GarroshIcecream/recipe_me/db"
	tea "github.com/charmbracelet/bubbletea"
)

type Manager struct {
	currentModel tea.Model
}

func NewManager() (*Manager, error) {
	cookbook, err := db.NewCookBook()
	if err != nil {
		log.Fatalf("Failed to create new Cookbook: %s", err)
		return nil, err
	}

	if err := cookbook.Open(); err != nil {
		log.Fatalf("Failed to open Cookbook: %s", err)
		return nil, err
	}

	manager := Manager{
		currentModel: NewListModel(*cookbook, nil),
	}

	return &manager, nil
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	nextModel, cmd := m.currentModel.Update(msg)

	if nextModel != m.currentModel {
		m.currentModel = nextModel
	}

	return m, cmd
}

func (m *Manager) Init() tea.Cmd {
	return nil
}

func (m Manager) View() string {
	return m.currentModel.View()
}
