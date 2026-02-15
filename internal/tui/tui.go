package tui

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	chat "github.com/GarroshIcecream/yummy/internal/tui/chat"
	detail "github.com/GarroshIcecream/yummy/internal/tui/detail"
	dialog "github.com/GarroshIcecream/yummy/internal/tui/dialog"
	edit "github.com/GarroshIcecream/yummy/internal/tui/edit"
	yummy_list "github.com/GarroshIcecream/yummy/internal/tui/list"
	main_menu "github.com/GarroshIcecream/yummy/internal/tui/main_menu"
	status "github.com/GarroshIcecream/yummy/internal/tui/status"
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
	mainMenu, err := main_menu.NewMainMenuModel(cookbook, currentTheme)
	if err != nil {
		slog.Error("Failed to create main menu", "error", err)
		return nil, err
	}

	executorService, err := chat.NewExecutorService(cookbook, sessionLog)
	if err != nil {
		slog.Error("Failed to create executor service", "error", err)
		return nil, err
	}

	list, err := yummy_list.NewListModel(cookbook, currentTheme)
	if err != nil {
		slog.Error("Failed to create list", "error", err)
		return nil, err
	}

	detailModel, err := detail.NewDetailModel(cookbook, currentTheme)
	if err != nil {
		slog.Error("Failed to create detail", "error", err)
		return nil, err
	}

	editModel, err := edit.NewEditModel(cookbook, currentTheme, 0)
	if err != nil {
		slog.Error("Failed to create edit", "error", err)
		return nil, err
	}

	chatModel, err := chat.NewChatModel(executorService, currentTheme)
	if err != nil {
		slog.Error("Failed to create chat", "error", err)
		return nil, err
	}

	cookingModel, err := detail.NewCookingModel(currentTheme)
	if err != nil {
		slog.Error("Failed to create cooking mode", "error", err)
		return nil, err
	}

	// Create models
	models := map[common.SessionState]common.TUIModel{
		common.SessionStateMainMenu: mainMenu,
		common.SessionStateList:     list,
		common.SessionStateDetail:   detailModel,
		common.SessionStateEdit:     editModel,
		common.SessionStateChat:     chatModel,
		common.SessionStateCooking:  cookingModel,
	}

	// Create status line
	statusLine := status.NewStatusLine(currentTheme)

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
		previousState := m.CurrentSessionState
		m.SetCurrentSessionState(common.SessionState(msg.SessionState))
		if m.ModalView {
			m.ModalView = false
		}

		// When entering detail view from another state with no recipe selected, open the recipe selector
		if msg.SessionState == common.SessionStateDetail && previousState != common.SessionStateDetail {
			if detailModel, ok := m.models[common.SessionStateDetail].(*detail.DetailModel); ok {
				if detailModel.Recipe == nil {
					recipeSelectorDialog, err := dialog.NewRecipeSelectorDialog(m.Cookbook, m.ThemeManager.GetCurrentTheme())
					if err != nil {
						slog.Error("Failed to create recipe selector dialog", "error", err)
					} else {
						cmds = append(cmds, messages.SendOpenModalViewMsg(recipeSelectorDialog, common.ModalTypeRecipeSelector))
					}
				}
			}
		}

	case messages.CloseModalViewMsg:
		m.ModalView = false
		m.overlayModel = nil
		m.modalModel = nil
		return m, nil

	case messages.CommandPaletteActionMsg:
		theme := m.ThemeManager.GetCurrentTheme()
		switch msg.Action {
		case dialog.ActionStateSelector:
			d, err := dialog.NewStateSelectorDialog(theme)
			if err != nil {
				slog.Error("Failed to create state selector dialog", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(d, common.ModalTypeStateSelector))

		case dialog.ActionThemeSelector:
			availableThemes := m.ThemeManager.GetAvailableThemes()
			currentThemeName := theme.Name
			d, err := dialog.NewThemeSelectorDialog(availableThemes, currentThemeName, theme)
			if err != nil {
				slog.Error("Failed to create theme selector dialog", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(d, common.ModalTypeThemeSelector))

		case dialog.ActionModelSelector:
			if chatModel, ok := m.models[common.SessionStateChat].(*chat.ChatModel); ok {
				currentModelName := chatModel.ExecutorService.GetCurrentModelName()
				installedModels := chatModel.ExecutorService.GetInstalledModels()
				d, err := dialog.NewModelSelectorDialog(installedModels, currentModelName, theme)
				if err != nil {
					slog.Error("Failed to create model selector dialog", "error", err)
					return m, nil
				}
				cmds = append(cmds, messages.SendOpenModalViewMsg(d, common.ModalTypeModelSelector))
			} else {
				slog.Error("Chat model not available for model selector")
			}

		case dialog.ActionRecipeSelector:
			d, err := dialog.NewRecipeSelectorDialog(m.Cookbook, theme)
			if err != nil {
				slog.Error("Failed to create recipe selector dialog", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(d, common.ModalTypeRecipeSelector))

		case dialog.ActionAddRecipe:
			d, err := dialog.NewAddRecipeFromURLDialog(m.Cookbook, theme)
			if err != nil {
				slog.Error("Failed to create add recipe dialog", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(d, common.ModalTypeAddRecipeFromURL))
		}

	case messages.OpenModalViewMsg:
		if m.ModalView && m.CurrentModalType == msg.ModalType {
			cmds = append(cmds, messages.SendCloseModalViewMsg())
		} else {
			m.ModalView = true
			m.CurrentModalType = msg.ModalType
			m.modalModel = msg.ModalModel

			yPos := overlay.Center
			if msg.ModalType == common.ModalTypeRating {
				yPos = overlay.Top
			}

			m.overlayModel = overlay.New(
				m.modalModel,
				m.GetCurrentModel(),
				overlay.Center,
				yPos,
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

			// Let the cooking model handle esc when its chat panel is open
			if m.CurrentSessionState == common.SessionStateCooking {
				if cookingModel, ok := m.models[common.SessionStateCooking].(*detail.CookingModel); ok {
					if cookingModel.IsChatOpen() {
						break
					}
				}
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
			// Skip in cooking mode — let ctrl+t reach the cooking model for timer control
			if m.CurrentSessionState == common.SessionStateCooking {
				break
			}

			availableThemes := m.ThemeManager.GetAvailableThemes()
			currentThemeName := m.ThemeManager.GetCurrentTheme().Name
			themeSelectorDialog, err := dialog.NewThemeSelectorDialog(availableThemes, currentThemeName, m.ThemeManager.GetCurrentTheme())
			if err != nil {
				slog.Error("Failed to create theme selector dialog", "error", err)
				return m, nil
			}

			cmds = append(cmds, messages.SendOpenModalViewMsg(themeSelectorDialog, common.ModalTypeThemeSelector))

		case key.Matches(msg, m.keyMap.RecipeSelector):
			recipeSelectorDialog, err := dialog.NewRecipeSelectorDialog(m.Cookbook, m.ThemeManager.GetCurrentTheme())
			if err != nil {
				slog.Error("Failed to create recipe selector dialog", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(recipeSelectorDialog, common.ModalTypeRecipeSelector))

		case key.Matches(msg, m.keyMap.CommandPalette):
			palette, err := dialog.NewCommandPaletteDialog(m.ThemeManager.GetCurrentTheme())
			if err != nil {
				slog.Error("Failed to create command palette", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(palette, common.ModalTypeCommandPalette))
		}
	}

	if m.ModalView {
		// Update foreground (modal) model
		fgModel, cmd := m.modalModel.Update(msg)
		m.modalModel = fgModel

		// Recreate overlay with updated models, preserving position
		yPos := overlay.Center
		if m.CurrentModalType == common.ModalTypeRating {
			yPos = overlay.Top
		}
		m.overlayModel = overlay.New(
			m.modalModel,
			m.GetCurrentModel(),
			overlay.Center,
			yPos,
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
