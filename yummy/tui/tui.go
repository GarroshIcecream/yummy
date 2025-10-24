package tui

import (
	"context"
	"log"

	"github.com/GarroshIcecream/yummy/yummy/config"
	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	chat "github.com/GarroshIcecream/yummy/yummy/tui/chat"
	detail "github.com/GarroshIcecream/yummy/yummy/tui/detail"
	dialog "github.com/GarroshIcecream/yummy/yummy/tui/dialog"
	edit "github.com/GarroshIcecream/yummy/yummy/tui/edit"
	yummy_list "github.com/GarroshIcecream/yummy/yummy/tui/list"
	main_menu "github.com/GarroshIcecream/yummy/yummy/tui/main_menu"
	status "github.com/GarroshIcecream/yummy/yummy/tui/status"
	"github.com/charmbracelet/bubbles/key"
	list "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

type Manager struct {
	// Session state management
	CurrentSessionState  consts.SessionState
	PreviousSessionState consts.SessionState
	ThemeManager         *themes.ThemeManager
	ModelState           consts.ModelState
	models               map[consts.SessionState]common.TUIModel

	// Database and configuration
	Cookbook   *db.CookBook
	SessionLog *db.SessionLog
	keyMap     config.KeyMap
	Ctx        context.Context

	// UI components
	statusLine          *status.StatusLine
	width               int
	height              int
	overlayModel        *overlay.Model
	stateSelectorDialog *dialog.StateSelectorDialogCmp
	showStateSelector   bool
}

func New(cookbook *db.CookBook, sessionLog *db.SessionLog, themeManager *themes.ThemeManager, ctx context.Context) (*Manager, error) {
	keymaps := config.DefaultKeyMap()
	currentTheme := themeManager.GetCurrentTheme()

	// Create models
	models := map[consts.SessionState]common.TUIModel{
		consts.SessionStateMainMenu: main_menu.New(cookbook, keymaps, currentTheme),
		consts.SessionStateList:     yummy_list.New(cookbook, keymaps, currentTheme, false),
		consts.SessionStateDetail:   detail.New(cookbook, keymaps, currentTheme),
		consts.SessionStateEdit:     edit.New(cookbook, keymaps, currentTheme, nil),
		consts.SessionStateChat:     chat.New(cookbook, sessionLog, keymaps, currentTheme),
	}

	// Create state selector dialog for overlay
	stateSelectorDialog := dialog.NewStateSelectorDialog(currentTheme)

	// Create status line
	statusLine := status.New(currentTheme)

	manager := &Manager{
		ThemeManager:        themeManager,
		Cookbook:            cookbook,
		keyMap:              keymaps,
		models:              models,
		statusLine:          statusLine,
		width:               consts.MainMenuContentWidth,
		height:              consts.DefaultViewportHeight,
		ModelState:          consts.ModelStateLoaded,
		Ctx:                 ctx,
		stateSelectorDialog: stateSelectorDialog,
		showStateSelector:   false,
	}

	return manager, nil
}

func (m *Manager) SetCurrentSessionState(state consts.SessionState) {
	if m.CurrentSessionState == state {
		return
	}
	m.PreviousSessionState = m.CurrentSessionState
	m.CurrentSessionState = state
}

func (m *Manager) GetCurrentSessionState() consts.SessionState {
	return m.CurrentSessionState
}

func (m *Manager) GetModel(state consts.SessionState) common.TUIModel {
	return m.models[state]
}

func (m *Manager) Init() tea.Cmd {
	if currentModel, exists := m.models[m.CurrentSessionState]; exists {
		return currentModel.Init()
	}
	return nil
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case messages.SessionStateMsg:
		m.SetCurrentSessionState(consts.SessionState(msg.SessionState))
		if m.showStateSelector {
			m.showStateSelector = false
			m.overlayModel = nil
		}

	case messages.CloseDialogMsg:
		m.showStateSelector = false
		m.overlayModel = nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusLine.SetSize(msg.Width, consts.StatusLineHeight)
		for _, model := range m.models {
			model.SetSize(msg.Width, msg.Height-consts.StatusLineHeight)
		}

		// Update state selector dialog size
		if m.stateSelectorDialog != nil {
			m.stateSelectorDialog.SetSize(msg.Width, msg.Height-consts.StatusLineHeight)
		}

	case tea.KeyMsg:
		// If state selector overlay is showing, handle its input first
		if m.showStateSelector {
			var overlayModel tea.Model
			overlayModel, cmd = m.overlayModel.Update(msg)
			if updatedOverlay, ok := overlayModel.(*overlay.Model); ok {
				m.overlayModel = updatedOverlay
			}
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, m.keyMap.ForceQuit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.CursorUp):
			if m.CurrentSessionState == consts.SessionStateDetail {
				if detailModel, ok := m.models[consts.SessionStateDetail].(*detail.DetailModel); ok {
					detailModel.ScrollUp(consts.DefaultScrollSpeed)
				}
			}
		case key.Matches(msg, m.keyMap.CursorDown):
			if m.CurrentSessionState == consts.SessionStateDetail {
				if detailModel, ok := m.models[consts.SessionStateDetail].(*detail.DetailModel); ok {
					detailModel.ScrollDown(consts.DefaultScrollSpeed)
				}
			}

		case key.Matches(msg, m.keyMap.Edit):
			if m.CurrentSessionState == consts.SessionStateDetail {
				if detailModel, ok := m.models[consts.SessionStateDetail].(*detail.DetailModel); ok {
					cmds = append(cmds, messages.SendSessionStateMsg(consts.SessionStateEdit))
					cmds = append(cmds, messages.SendEditRecipeMsg(detailModel.CurrentRecipe.ID))
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keyMap.Back):
			if m.CurrentSessionState == consts.SessionStateList {
				if listModel, ok := m.models[consts.SessionStateList].(*yummy_list.ListModel); ok {
					if listModel.RecipeList.FilterState() != list.Filtering {
						m.SetCurrentSessionState(m.PreviousSessionState)
					}
				}
			} else {
				m.SetCurrentSessionState(m.PreviousSessionState)
			}

			return m, nil
		case key.Matches(msg, m.keyMap.Add):
			if m.CurrentSessionState == consts.SessionStateList {
				if listModel, ok := m.models[consts.SessionStateList].(*yummy_list.ListModel); ok {
					if listModel.RecipeList.FilterState() != list.Filtering {
						m.SetCurrentSessionState(m.PreviousSessionState)
					}
				}
			}
			return m, nil
		case key.Matches(msg, m.keyMap.StateSelector):
			// Create overlay with current model as background and state selector as foreground
			m.showStateSelector = true
			if currentModel, exists := m.models[m.CurrentSessionState]; exists {
				m.overlayModel = overlay.New(
					m.stateSelectorDialog, // foreground (state selector)
					currentModel,          // background (current model)
					overlay.Center,        // x position
					overlay.Center,        // y position
					0,                     // x offset
					0,                     // y offset
				)
			}
			return m, nil
		}
	}

	// Update the current model (only if overlay is not showing)
	if !m.showStateSelector {
		if currentModel, exists := m.models[m.CurrentSessionState]; exists {
			var model tea.Model
			model, cmd = currentModel.Update(msg)
			if updatedModel, ok := model.(common.TUIModel); ok {
				m.models[m.CurrentSessionState] = updatedModel
			} else {
				log.Printf("Model for state %v is not a TUIModel", m.CurrentSessionState)
			}
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Manager) View() string {
	var content string

	// If state selector overlay is showing, render the overlay
	if m.showStateSelector && m.overlayModel != nil {
		content = m.overlayModel.View()
	} else {
		// Render the current model normally
		if currentModel, exists := m.models[m.CurrentSessionState]; exists {
			content = currentModel.View()
		}
	}

	// Render status line if not loading
	if m.models[m.CurrentSessionState].GetModelState() != consts.ModelStateLoading {
		currentModel := m.models[m.CurrentSessionState]
		statusInfo := status.CreateStatusInfo(currentModel)
		statusLine := m.statusLine.Render(statusInfo)
		content = lipgloss.JoinVertical(lipgloss.Left, content, statusLine)
	}

	return content
}
