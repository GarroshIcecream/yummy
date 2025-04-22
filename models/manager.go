package models

import (
	db "github.com/GarroshIcecream/yummy/db"
	tea "github.com/charmbracelet/bubbletea"
)

type Manager struct {
	currentModel tea.Model
	width        int
	height       int
}

func NewManager() (*Manager, error) {
	cookbook, err := db.NewCookBook()
	if err != nil {
		return nil, err
	}

	manager := Manager{
		currentModel: NewListModel(*cookbook, nil),
	}

	return &manager, nil
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	nextModel, cmd := m.currentModel.Update(msg)

	if nextModel != m.currentModel {
		m.currentModel = nextModel
		// Get the init command from the new model
		initCmd := m.currentModel.Init()
		return m, tea.Batch(
			cmd,
			initCmd,
			func() tea.Msg {
				return tea.WindowSizeMsg{
					Width:  m.width,
					Height: m.height,
				}
			},
		)
	}

	return m, cmd
}

func (m *Manager) Init() tea.Cmd {
	return m.currentModel.Init()
}

func (m Manager) View() string {
	return m.currentModel.View()
}
