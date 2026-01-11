package tui

import (
	"context"
	"fmt"
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
	models               map[common.SessionState]common.TUIModel
	config               *config.GeneralConfig

	// Database and configuration
	Cookbook   *db.CookBook
	SessionLog *db.SessionLog
	keyMap     config.ManagerKeyMap
	Ctx        context.Context

	// UI components
	statusLine       *status.StatusLine
	ModalView        bool
	CurrentModalType common.ModalType
	overlayModel     *overlay.Model
	modalModel       tea.Model
}

func New(cookbook *db.CookBook, sessionLog *db.SessionLog, themeManager *themes.ThemeManager, ctx context.Context) (*Manager, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetManagerKeyMap()
	currentTheme := themeManager.GetCurrentTheme()
	mainMenu, err := main_menu.New(cookbook, currentTheme)
	if err != nil {
		slog.Error("Failed to create main menu", "error", err)
		return nil, err
	}

	executorService, err := chat.NewExecutorService(cookbook, sessionLog)
	if err != nil {
		slog.Error("Failed to create executor service", "error", err)
		return nil, err
	}

	list, err := yummy_list.New(cookbook, currentTheme)
	if err != nil {
		slog.Error("Failed to create list", "error", err)
		return nil, err
	}

	detail, err := detail.New(cookbook, currentTheme)
	if err != nil {
		slog.Error("Failed to create detail", "error", err)
		return nil, err
	}

	edit, err := edit.New(cookbook, currentTheme, 0)
	if err != nil {
		slog.Error("Failed to create edit", "error", err)
		return nil, err
	}

	chat, err := chat.New(executorService, currentTheme)
	if err != nil {
		slog.Error("Failed to create chat", "error", err)
		return nil, err
	}

	// Create models
	models := map[common.SessionState]common.TUIModel{
		common.SessionStateMainMenu: mainMenu,
		common.SessionStateList:     list,
		common.SessionStateDetail:   detail,
		common.SessionStateEdit:     edit,
		common.SessionStateChat:     chat,
	}

	// Create status line
	statusLine := status.New(currentTheme)

	generalConfig := config.GetGeneralConfig()
	manager := &Manager{
		ThemeManager:         themeManager,
		CurrentSessionState:  common.SessionStateMainMenu,
		PreviousSessionState: common.SessionStateMainMenu,
		Cookbook:             cookbook,
		models:               models,
		statusLine:           statusLine,
		Ctx:                  ctx,
		ModalView:            false,
		config:               generalConfig,
		keyMap:               keymaps,
	}

	return manager, nil
}

func (m *Manager) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, tea.SetWindowTitle("Yummy"))

	if currentModel, exists := m.models[m.CurrentSessionState]; exists {
		cmds = append(cmds, currentModel.Init())
	}

	return tea.Batch(cmds...)
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case messages.SessionStateMsg:
		m.SetCurrentSessionState(common.SessionState(msg.SessionState))
		if m.ModalView {
			m.ModalView = false
		}

	case messages.CloseModalViewMsg:
		m.ModalView = false
		m.overlayModel = nil
		m.modalModel = nil
		return m, nil

	case messages.OpenModalViewMsg:
		if m.ModalView && m.CurrentModalType == msg.ModalType {
			cmds = append(cmds, messages.SendCloseModalViewMsg())
		} else {
			m.ModalView = true
			m.CurrentModalType = msg.ModalType
			m.modalModel = msg.ModalModel
			m.overlayModel = overlay.New(
				m.modalModel,
				m.GetCurrentModel(),
				overlay.Center,
				overlay.Center,
				0,
				0,
			)
		}

	case tea.WindowSizeMsg:
		m.statusLine.SetSize(msg.Width, m.config.Height)
		for _, model := range m.models {
			model.SetSize(msg.Width, msg.Height-m.config.Height)
		}

	case messages.ThemeSelectedMsg:
		err := m.ThemeManager.SetThemeByName(msg.ThemeName)
		if err != nil {
			slog.Error("Failed to set theme", "error", err, "theme", msg.ThemeName)
			return m, nil
		}

		// Update all models with new theme
		newTheme := m.ThemeManager.GetCurrentTheme()
		m.updateAllModelsTheme(newTheme)
		m.statusLine.SetTheme(newTheme)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.ForceQuit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Back):
			if m.ModalView {
				m.ModalView = false
				m.modalModel = nil
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

		case key.Matches(msg, m.keyMap.StateSelector):
			stateSelectorDialog, err := dialog.NewStateSelectorDialog(m.ThemeManager.GetCurrentTheme())
			if err != nil {
				slog.Error("Failed to create state selector dialog", "error", err)
				return m, nil
			}

			cmds = append(cmds, messages.SendOpenModalViewMsg(stateSelectorDialog, common.ModalTypeStateSelector))

		case key.Matches(msg, m.keyMap.ThemeSelector):
			availableThemes := m.ThemeManager.GetAvailableThemes()
			currentThemeName := m.ThemeManager.GetCurrentTheme().Name
			themeSelectorDialog, err := dialog.NewThemeSelectorDialog(availableThemes, currentThemeName, m.ThemeManager.GetCurrentTheme())
			if err != nil {
				slog.Error("Failed to create theme selector dialog", "error", err)
				return m, nil
			}

			cmds = append(cmds, messages.SendOpenModalViewMsg(themeSelectorDialog, common.ModalTypeThemeSelector))
		}
	}

	if m.ModalView {
		// Update foreground (modal) model
		fgModel, cmd := m.modalModel.Update(msg)
		m.modalModel = fgModel

		// Recreate overlay with updated models
		m.overlayModel = overlay.New(
			m.modalModel,
			m.GetCurrentModel(),
			overlay.Center,
			overlay.Center,
			0,
			0,
		)

		cmds = append(cmds, cmd)

	} else {
		if currentModel, exists := m.models[m.CurrentSessionState]; exists {
			model, cmd := currentModel.Update(msg)
			m.models[m.CurrentSessionState] = model

			cmds = append(cmds, cmd)
		}
	}

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

func (m *Manager) updateAllModelsTheme(theme *themes.Theme) {
	for _, model := range m.models {
		model.SetTheme(theme)
	}
}
