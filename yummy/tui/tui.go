package tui

import (
	"context"
	"log"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	chat "github.com/GarroshIcecream/yummy/yummy/tui/chat"
	detail "github.com/GarroshIcecream/yummy/yummy/tui/detail"
	edit "github.com/GarroshIcecream/yummy/yummy/tui/edit"
	yummy_list "github.com/GarroshIcecream/yummy/yummy/tui/list"
	main_menu "github.com/GarroshIcecream/yummy/yummy/tui/main_menu"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/key"
	list "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type TUIModel interface {
	tea.Model
	SetSize(width, height int)
	GetSize() (width, height int)
}

type Manager struct {
	CurrentSessionState  ui.SessionState
	PreviousSessionState ui.SessionState
	models               map[ui.SessionState]TUIModel
	Cookbook             *db.CookBook
	keyMap               config.KeyMap
	Ctx                  context.Context
}

func New(cookbook *db.CookBook, ctx context.Context) (*Manager, error) {
	keymaps := config.DefaultKeyMap()

	models := map[ui.SessionState]TUIModel{
		ui.SessionStateMainMenu: main_menu.New(cookbook, keymaps),
		ui.SessionStateList:     yummy_list.New(cookbook, keymaps),
		ui.SessionStateDetail:   detail.New(cookbook, keymaps),
		ui.SessionStateEdit:     edit.New(cookbook, keymaps, nil),
		ui.SessionStateChat:     chat.New(cookbook, keymaps),
	}

	manager := Manager{
		Cookbook:             cookbook,
		keyMap:               keymaps,
		models:               models,
		CurrentSessionState:  ui.SessionStateMainMenu,
		PreviousSessionState: ui.SessionStateMainMenu,
		Ctx:                  ctx,
	}

	return &manager, nil
}

func (m *Manager) SetCurrentSessionState(state ui.SessionState) {
	m.PreviousSessionState = m.CurrentSessionState
	m.CurrentSessionState = state
}

func (m *Manager) GetCurrentSessionState() ui.SessionState {
	return m.CurrentSessionState
}

func (m *Manager) GetModel(state ui.SessionState) TUIModel {
	return m.models[state]
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case ui.SessionStateMsg:
		m.SetCurrentSessionState(msg.SessionState)

	case tea.WindowSizeMsg:
		for _, model := range m.models {
			model.SetSize(msg.Width, msg.Height)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Up):
			if m.CurrentSessionState == ui.SessionStateDetail {
				if detailModel, ok := m.models[ui.SessionStateDetail].(*detail.DetailModel); ok {
					detailModel.ScrollUp(3)
				}
			}
		case key.Matches(msg, m.keyMap.Down):
			if m.CurrentSessionState == ui.SessionStateDetail {
				if detailModel, ok := m.models[ui.SessionStateDetail].(*detail.DetailModel); ok {
					detailModel.ScrollDown(3)
				}
			}

		case key.Matches(msg, m.keyMap.Edit):
			if m.CurrentSessionState == ui.SessionStateDetail {
				if detailModel, ok := m.models[ui.SessionStateDetail].(*detail.DetailModel); ok {
					cmds = append(cmds, ui.SendSessionStateMsg(ui.SessionStateEdit))
					cmds = append(cmds, ui.SendEditRecipeMsg(detailModel.CurrentRecipe.ID))
				}
			}

		case key.Matches(msg, m.keyMap.Back):
			if m.CurrentSessionState == ui.SessionStateList {
				if listModel, ok := m.models[ui.SessionStateList].(*yummy_list.ListModel); ok {
					if listModel.RecipeList.FilterState() != list.Filtering {
						m.SetCurrentSessionState(m.PreviousSessionState)
					}
				}
			} else {
				m.SetCurrentSessionState(m.PreviousSessionState)
			}
		case key.Matches(msg, m.keyMap.Add):
			if m.CurrentSessionState == ui.SessionStateList {
				if listModel, ok := m.models[ui.SessionStateList].(*yummy_list.ListModel); ok {
					if listModel.RecipeList.FilterState() != list.Filtering {
						m.SetCurrentSessionState(m.PreviousSessionState)
					}
				}
			}
		}
	}

	// Update the current model
	if currentModel, exists := m.models[m.CurrentSessionState]; exists {
		var model tea.Model
		model, cmd = currentModel.Update(msg)
		if updatedModel, ok := model.(TUIModel); ok {
			m.models[m.CurrentSessionState] = updatedModel
		} else {
			log.Printf("Model for state %v is not a TUIModel", m.CurrentSessionState)
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Manager) Init() tea.Cmd {
	if currentModel, exists := m.models[m.CurrentSessionState]; exists {
		return currentModel.Init()
	}
	return nil
}

func (m Manager) View() string {
	if currentModel, exists := m.models[m.CurrentSessionState]; exists {
		return currentModel.View()
	}
	return ""
}
