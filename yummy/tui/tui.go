package tui

import (
	"context"
	"log/slog"

	"github.com/GarroshIcecream/yummy/yummy/config"
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
	CurrentSessionState  common.SessionState
	PreviousSessionState common.SessionState
	ThemeManager         *themes.ThemeManager
	ModalView            bool
	models               map[common.SessionState]common.TUIModel
	config               *config.GeneralConfig

	// Database and configuration
	Cookbook   *db.CookBook
	SessionLog *db.SessionLog
	keyMap     config.KeyMap
	Ctx        context.Context

	// UI components
	statusLine          *status.StatusLine
	overlayModel        *overlay.Model
	stateSelectorDialog *dialog.StateSelectorDialogCmp
}

func New(cookbook *db.CookBook, sessionLog *db.SessionLog, themeManager *themes.ThemeManager, cfg *config.Config, ctx context.Context) (*Manager, error) {
	// Create keymap with custom bindings from config
	keymaps := config.CreateKeyMapFromConfig(cfg.Keymap)
	currentTheme := themeManager.GetCurrentTheme()

	executorService, err := chat.NewExecutorService(cookbook, sessionLog, cfg.Chat.DefaultModel, cfg.Chat.SystemPrompt)
	if err != nil {
		slog.Error("Failed to create executor service", "error", err)
		return nil, err
	}

	// Create models
	models := map[common.SessionState]common.TUIModel{
		common.SessionStateMainMenu: main_menu.New(cookbook, keymaps, currentTheme, &cfg.MainMenu),
		common.SessionStateList:     yummy_list.New(cookbook, keymaps, currentTheme, &cfg.List),
		common.SessionStateDetail:   detail.New(cookbook, keymaps, currentTheme, &cfg.Detail),
		common.SessionStateEdit:     edit.New(cookbook, keymaps, currentTheme, 0),
		common.SessionStateChat:     chat.New(executorService, keymaps, currentTheme, &cfg.Chat),
	}

	// Create status line
	statusLine := status.New(currentTheme, &cfg.StatusLine)

	// Create state selector dialog for overlay
	stateSelectorDialog := dialog.NewStateSelectorDialog(currentTheme, &cfg.StateSelectorDialog, keymaps)
	overlayModel := overlay.New(
		stateSelectorDialog,
		models[common.SessionStateMainMenu],
		overlay.Center,
		overlay.Center,
		0,
		0,
	)

	manager := &Manager{
		ThemeManager:         themeManager,
		CurrentSessionState:  common.SessionStateMainMenu,
		PreviousSessionState: common.SessionStateMainMenu,
		Cookbook:             cookbook,
		keyMap:               keymaps,
		models:               models,
		statusLine:           statusLine,
		Ctx:                  ctx,
		stateSelectorDialog:  stateSelectorDialog,
		ModalView:            false,
		overlayModel:         overlayModel,
		config:               &cfg.General,
	}

	return manager, nil
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
		m.SetCurrentSessionState(common.SessionState(msg.SessionState))
		if m.ModalView {
			m.ModalView = false
		}

	case messages.CloseDialogMsg:
		m.ModalView = false

	case tea.WindowSizeMsg:
		m.statusLine.SetSize(msg.Width, m.config.Height)
		for _, model := range m.models {
			model.SetSize(msg.Width, msg.Height-m.config.Height)
		}

		// Update state selector dialog size
		if m.stateSelectorDialog != nil {
			m.stateSelectorDialog.SetSize(msg.Width, msg.Height-m.config.Height)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.ForceQuit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Back):
			if m.ModalView {
				m.ModalView = false
				return m, nil
			}

			if m.CurrentSessionState == common.SessionStateList {
				if listModel, ok := m.models[common.SessionStateList].(*yummy_list.ListModel); ok {
					if listModel.RecipeList.FilterState() != list.Filtering {
						m.SetCurrentSessionState(m.PreviousSessionState)
					} else {
						listModel.RecipeList.SetFilterState(list.Unfiltered)
					}
				}
			} else {
				m.SetCurrentSessionState(m.PreviousSessionState)
			}

			return m, nil
		case key.Matches(msg, m.keyMap.Add):
			if m.CurrentSessionState == common.SessionStateList {
				if listModel, ok := m.models[common.SessionStateList].(*yummy_list.ListModel); ok {
					if listModel.RecipeList.FilterState() != list.Filtering {
						m.SetCurrentSessionState(m.PreviousSessionState)
					}
				}
			}
			return m, nil
		case key.Matches(msg, m.keyMap.StateSelector):
			if m.ModalView {
				m.ModalView = false
				return m, nil
			} else {
				m.ModalView = true
				m.overlayModel.Background = m.GetCurrentModel()
				return m, nil
			}
		}
	}

	if m.ModalView {
		fg, fgCmd := m.overlayModel.Foreground.Update(msg)
		m.overlayModel.Foreground = fg
		cmds = append(cmds, fgCmd)
	} else {
		if currentModel, exists := m.models[m.CurrentSessionState]; exists {
			var model tea.Model
			model, cmd = currentModel.Update(msg)
			if updatedModel, ok := model.(common.TUIModel); ok {
				m.models[m.CurrentSessionState] = updatedModel
			} else {
				slog.Error("Model for state is not a TUIModel", "state", m.CurrentSessionState)
				return m, nil
			}
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Manager) View() string {
	var content string

	// If state selector overlay is showing, render the overlay
	if m.ModalView {
		content = m.overlayModel.View()
	} else {
		content = m.GetCurrentModel().View()
	}

	// Render status line
	if m.GetCurrentModel().GetModelState() == common.ModelStateLoaded {
		currentModel := m.GetCurrentModel()
		statusInfo := status.CreateStatusInfo(currentModel)
		statusLine := m.statusLine.Render(statusInfo)
		content = lipgloss.JoinVertical(lipgloss.Left, content, statusLine)
	}

	return content
}

func (m *Manager) SetCurrentSessionState(state common.SessionState) {
	if m.CurrentSessionState == state {
		return
	}
	m.PreviousSessionState = m.CurrentSessionState
	m.CurrentSessionState = state
}

func (m *Manager) GetCurrentModel() common.TUIModel {
	return m.models[m.CurrentSessionState]
}

func (m *Manager) GetModel(state common.SessionState) common.TUIModel {
	return m.models[state]
}
