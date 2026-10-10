package tui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/key"
	list "charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/llms"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	chat "github.com/GarroshIcecream/yummy/internal/tui/chat"
	detail "github.com/GarroshIcecream/yummy/internal/tui/detail"
	dialog "github.com/GarroshIcecream/yummy/internal/tui/dialog"
	edit "github.com/GarroshIcecream/yummy/internal/tui/edit"
	yummy_list "github.com/GarroshIcecream/yummy/internal/tui/list"
	main_menu "github.com/GarroshIcecream/yummy/internal/tui/main_menu"
	overlay "github.com/GarroshIcecream/yummy/internal/tui/overlay"
	status "github.com/GarroshIcecream/yummy/internal/tui/status"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

type Manager struct {
	// Session state management
	CurrentSessionState  common.SessionState
	PreviousSessionState common.SessionState
	ThemeManager         *themes.ThemeManager
	models               map[common.SessionState]common.TUIModel
	config               *config.StatusLineConfig
	configDir            string
	themePreviewOriginal string
	themePreviewActive   bool

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
	width            int
	height           int
}

func New(cookbook *db.CookBook, sessionLog *db.SessionLog, themeManager *themes.ThemeManager, configDir string, ctx context.Context) (*Manager, error) {
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

	// Create shared LLM service (with tools)
	llmService, err := llms.NewLLMService(cookbook)
	if err != nil {
		slog.Error("Failed to create LLM service", "error", err)
		return nil, err
	}

	executorService, err := chat.NewExecutorService(llmService, cookbook, sessionLog)
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

	cookingModel, err := detail.NewCookingModel(currentTheme, llmService)
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

	statusLineConfig := config.GetStatusLineConfig()
	manager := &Manager{
		ThemeManager:         themeManager,
		CurrentSessionState:  common.SessionStateMainMenu,
		PreviousSessionState: common.SessionStateMainMenu,
		Cookbook:             cookbook,
		models:               models,
		statusLine:           statusLine,
		Ctx:                  ctx,
		ModalView:            false,
		config:               statusLineConfig,
		configDir:            configDir,
		keyMap:               keymaps,
	}

	return manager, nil
}

func (m *Manager) Init() tea.Cmd {
	var cmds []tea.Cmd

	if currentModel, exists := m.models[m.CurrentSessionState]; exists {
		cmds = append(cmds, currentModel.Init())
	}

	return tea.Batch(cmds...)
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case messages.RecipeSelectedMsg:
		// Close the recipe selector modal if open, then always navigate to detail.
		// This must happen unconditionally because the list model sends
		// RecipeSelectedMsg and SessionStateMsg(Detail) as a sequence; by the time
		// RecipeSelectedMsg is processed the OpenModalViewMsg command from
		// SessionStateMsg hasn't fired yet, so m.ModalView would still be false.
		if m.ModalView && m.CurrentModalType == common.ModalTypeRecipeSelector {
			m.ModalView = false
			m.overlayModel = nil
			m.modalModel = nil
		}
		m.SetCurrentSessionState(common.SessionStateDetail)

	case messages.SessionStateMsg:
		previousState := m.CurrentSessionState
		m.SetCurrentSessionState(common.SessionState(msg.SessionState))
		if m.ModalView {
			m.ModalView = false
		}

		// Entering the chat view needs an explicit render: the chat viewport's
		// content is populated only by RenderConversationAsMarkdownMsg, so
		// without this the empty-state welcome (and any existing history) never
		// paints and the panel shows as a blank void.
		if msg.SessionState == common.SessionStateChat && previousState != common.SessionStateChat {
			cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())
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
		m.restoreThemePreview()
		m.ModalView = false
		m.overlayModel = nil
		m.modalModel = nil
		return m, nil

	case messages.RecipeScrapersVersionMsg:
		if menuModel, ok := m.models[common.SessionStateMainMenu]; ok {
			updated, cmd := menuModel.Update(msg)
			m.models[common.SessionStateMainMenu] = updated
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case messages.SaveMsg:
		if listModel, ok := m.models[common.SessionStateList].(*yummy_list.ListModel); ok {
			cmds = append(cmds, listModel.RefreshRecipeList())
		}

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
			availableThemes := m.ThemeManager.GetAvailableThemeInfos()
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

		case dialog.ActionCreateRecipe:
			cmds = append(cmds, tea.Sequence(
				messages.SendSessionStateMsg(common.SessionStateEdit),
				messages.SendEditRecipeMsg(nil),
			))

		case dialog.ActionAddRecipe:
			d, err := dialog.NewAddRecipeFromURLDialog(m.Cookbook, theme)
			if err != nil {
				slog.Error("Failed to create add recipe dialog", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(d, common.ModalTypeAddRecipeFromURL))

		case dialog.ActionToggleAI:
			if cfg := config.GetGlobalConfig(); cfg != nil {
				cfg.AddRecipeFromURLDialog.AIParsing = !cfg.AddRecipeFromURLDialog.AIParsing
				if err := cfg.Save(m.configDir); err != nil {
					slog.Error("Failed to persist AI parsing setting", "error", err)
				}
			}

		case dialog.ActionScraperUpdate:
			cfg := config.GetGlobalConfig()
			pythonPath := ""
			if cfg != nil {
				pythonPath = cfg.AddRecipeFromURLDialog.PythonPath
			}
			d := dialog.NewScraperUpdateDialog(pythonPath, theme)
			cmds = append(cmds, messages.SendOpenModalViewMsg(d, common.ModalTypeScraperUpdate))
		}

	case messages.OpenModalViewMsg:
		if m.ModalView && m.CurrentModalType == msg.ModalType {
			cmds = append(cmds, messages.SendCloseModalViewMsg())
		} else {
			m.ModalView = true
			m.CurrentModalType = msg.ModalType
			if msg.ModalType == common.ModalTypeThemeSelector {
				currentTheme := m.ThemeManager.GetCurrentTheme()
				if currentTheme != nil {
					m.themePreviewOriginal = currentTheme.Name
					m.themePreviewActive = true
				}
			}
			m.modalModel = msg.ModalModel
			m.sizeModalModel()
			cmds = append(cmds, m.modalModel.Init())

			m.overlayModel = m.newOverlay()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - m.config.Height
		m.statusLine.SetSize(msg.Width, m.config.Height)
		for _, model := range m.models {
			model.SetSize(m.width, m.height)
		}
		m.sizeModalModel()

	case messages.ThemePreviewMsg:
		if err := m.applyThemeByName(msg.ThemeName); err != nil {
			slog.Error("Failed to preview theme", "error", err, "theme", msg.ThemeName)
			return m, nil
		}

	case messages.ThemeSelectedMsg:
		if err := m.applyThemeByName(msg.ThemeName); err != nil {
			slog.Error("Failed to set theme", "error", err, "theme", msg.ThemeName)
			return m, nil
		}
		m.themePreviewActive = false
		m.themePreviewOriginal = ""

		cfg := config.GetGlobalConfig()
		if cfg != nil && cfg.Theme != msg.ThemeName {
			cfg.Theme = msg.ThemeName
			if err := cfg.Save(m.configDir); err != nil {
				slog.Error("Failed to persist theme selection", "error", err, "theme", msg.ThemeName)
			}
		}

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.ForceQuit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Back):
			if m.ModalView {
				m.restoreThemePreview()
				m.ModalView = false
				m.overlayModel = nil
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

			// In the list, esc first clears an active or applied filter.
			if m.CurrentSessionState == common.SessionStateList {
				if listModel, ok := m.models[common.SessionStateList].(*yummy_list.ListModel); ok {
					if listModel.RecipeList.FilterState() != list.Unfiltered {
						listModel.RecipeList.ResetFilter()
						return m, nil
					}
				}
			}

			m.SetCurrentSessionState(m.parentState(m.CurrentSessionState))
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

			availableThemes := m.ThemeManager.GetAvailableThemeInfos()
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

		case key.Matches(msg, m.keyMap.Add) && !m.ModalView && m.addKeyAvailable():
			cmds = append(cmds, messages.SendCommandPaletteActionMsg(dialog.ActionAddRecipe))
			return m, tea.Batch(cmds...)

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

		// Recreate overlay with updated models
		m.overlayModel = m.newOverlay()

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

func (m Manager) View() tea.View {
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

	if theme := m.ThemeManager.GetCurrentTheme(); theme != nil && m.width > 0 {
		totalHeight := m.height
		if m.GetCurrentModel().GetModelState() == common.ModelStateLoaded {
			totalHeight += m.config.Height
		}
		content = paintCanvas(content, theme, m.width, totalHeight)
	}

	v := tea.NewView(content)
	v.WindowTitle = "Yummy"
	v.ProgramStatus = m.programStatus()
	v.AltScreen = true
	// Mouse capture is opt-in (general.mouse): while it is on, the terminal
	// can't select text or open links, so by default leave the mouse to it.
	if cfg := config.GetGeneralConfig(); cfg != nil && cfg.Mouse {
		v.MouseMode = tea.MouseModeCellMotion
	}
	return v
}

// busyReporter is implemented by views and dialogs that run long background
// work (an LLM answer, a recipe import).
type busyReporter interface {
	Busy() (bool, string)
}

// programStatus reports what the app is doing to terminals that support the
// Program Status Protocol (OSC 7501), e.g. a "working" indicator on the tab
// while the assistant answers. Other terminals ignore it.
func (m *Manager) programStatus() *tea.ProgramStatus {
	candidates := []any{m.GetCurrentModel(), m.models[common.SessionStateChat]}
	if m.ModalView {
		candidates = append([]any{m.modalModel}, candidates...)
	}
	for _, c := range candidates {
		if r, ok := c.(busyReporter); ok {
			if busy, msg := r.Busy(); busy {
				return &tea.ProgramStatus{State: tea.ProgramStateWorking, App: "yummy", Message: msg}
			}
		}
	}
	return &tea.ProgramStatus{State: tea.ProgramStateIdle, App: "yummy"}
}

// addKeyAvailable reports whether the global add-recipe key applies: not in
// views that edit text (chat, recipe editor, cooking chat), where it would
// steal the key from the input.
func (m *Manager) addKeyAvailable() bool {
	switch m.CurrentSessionState {
	case common.SessionStateMainMenu, common.SessionStateDetail:
		return true
	case common.SessionStateCooking:
		if cookingModel, ok := m.models[common.SessionStateCooking].(*detail.CookingModel); ok {
			return !cookingModel.IsChatOpen()
		}
	}
	// The list handles ctrl+a itself (and must not while filtering).
	return false
}

// parentState is where esc leads from a view. Navigation goes up a fixed
// hierarchy towards the main menu rather than back to whichever view was
// last open, so repeated esc always ends at the main menu:
//
//	main menu
//	├── recipe list
//	│   └── recipe detail
//	│       ├── cooking mode
//	│       └── edit recipe
//	└── chat assistant
//
// A new recipe (not yet saved) has no detail view, so its editor returns to
// the list or main menu it was opened from.
func (m *Manager) parentState(state common.SessionState) common.SessionState {
	switch state {
	case common.SessionStateDetail:
		return common.SessionStateList
	case common.SessionStateCooking:
		return common.SessionStateDetail
	case common.SessionStateEdit:
		if editModel, ok := m.models[common.SessionStateEdit].(*edit.EditModel); ok && editModel.IsNew() {
			if m.PreviousSessionState == common.SessionStateList {
				return common.SessionStateList
			}
			return common.SessionStateMainMenu
		}
		return common.SessionStateDetail
	default:
		return common.SessionStateMainMenu
	}
}

// newOverlay centers the current modal over the current view. Every dialog
// opens in the same place so they read as one family.
func (m *Manager) newOverlay() *overlay.Model {
	return overlay.New(
		overlay.WrapTeaViewable(m.modalModel),
		m.GetCurrentModel(),
		overlay.Center,
		overlay.Center,
		0,
		0,
	)
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

func (m *Manager) applyThemeByName(name string) error {
	if err := m.ThemeManager.SetThemeByName(name); err != nil {
		return err
	}

	newTheme := m.ThemeManager.GetCurrentTheme()
	m.updateAllModelsTheme(newTheme)
	m.statusLine.SetTheme(newTheme)

	if m.CurrentModalType == common.ModalTypeThemeSelector {
		if selector, ok := m.modalModel.(*dialog.ThemeSelectorDialogCmp); ok {
			selector.SetTheme(newTheme)
		}
	}

	return nil
}

func (m *Manager) restoreThemePreview() {
	if !m.themePreviewActive || m.themePreviewOriginal == "" {
		return
	}
	if err := m.applyThemeByName(m.themePreviewOriginal); err != nil {
		slog.Error("Failed to restore preview theme", "error", err, "theme", m.themePreviewOriginal)
	}
	m.themePreviewActive = false
	m.themePreviewOriginal = ""
}

// paintCanvas makes the whole frame sit on the theme's canvas: it pads the
// frame to the full screen height and re-establishes the base background and
// text color after every SGR reset, so no cell anywhere in the app — gaps
// between panels, glamour output (which resets to the default background after
// every span), huh forms, margins — falls through to the terminal's own colors.
// Spans that set their own background keep it.
func paintCanvas(content string, theme *themes.Theme, width, height int) string {
	if lines := strings.Count(content, "\n") + 1; lines < height {
		content += strings.Repeat("\n", height-lines)
	}
	return utils.FillColors(content, theme.AppBackground.GetForeground(), theme.BaseBg, width)
}

func (m *Manager) sizeModalModel() {
	if m.modalModel == nil || m.width <= 0 || m.height <= 0 {
		return
	}

	if sizable, ok := m.modalModel.(interface{ SetSize(int, int) }); ok {
		sizable.SetSize(m.width, m.height)
	}
}
