package tui

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/GarroshIcecream/yummy/yummy/recipe"
	chat "github.com/GarroshIcecream/yummy/yummy/tui/chat"
	detail "github.com/GarroshIcecream/yummy/yummy/tui/detail"
	dialog "github.com/GarroshIcecream/yummy/yummy/tui/dialog"
	edit "github.com/GarroshIcecream/yummy/yummy/tui/edit"
	yummy_list "github.com/GarroshIcecream/yummy/yummy/tui/list"
	main_menu "github.com/GarroshIcecream/yummy/yummy/tui/main_menu"
	status "github.com/GarroshIcecream/yummy/yummy/tui/status"
	utils "github.com/GarroshIcecream/yummy/yummy/tui/utils"
	"github.com/charmbracelet/bubbles/key"
	list "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TUIModel interface {
	tea.Model
	GetModelState() utils.ModelState
	SetSize(width, height int)
	GetSize() (width, height int)
}

type Manager struct {
	CurrentSessionState  utils.SessionState
	PreviousSessionState utils.SessionState
	models               map[utils.SessionState]TUIModel
	Cookbook             *db.CookBook
	keyMap               config.KeyMap
	ModelState           utils.ModelState
	Ctx                  context.Context
	statusLine           *status.StatusLine
	width                int
	height               int
}

func New(cookbook *db.CookBook, ctx context.Context) (*Manager, error) {
	keymaps := config.DefaultKeyMap()

	models := map[utils.SessionState]TUIModel{
		utils.SessionStateMainMenu:        main_menu.New(cookbook, keymaps),
		utils.SessionStateList:            yummy_list.New(cookbook, keymaps, false),
		utils.SessionStateDetail:          detail.New(cookbook, keymaps),
		utils.SessionStateEdit:            edit.New(cookbook, keymaps, nil),
		utils.SessionStateChat:            chat.New(cookbook, keymaps),
		utils.SessionStateStateSelector:   dialog.NewStateSelectorDialog(),
		utils.SessionStateSessionSelector: dialog.NewSessionSelectorDialog(cookbook, keymaps),
	}

	manager := Manager{
		Cookbook:             cookbook,
		keyMap:               keymaps,
		models:               models,
		CurrentSessionState:  utils.SessionStateMainMenu,
		PreviousSessionState: utils.SessionStateMainMenu,
		Ctx:                  ctx,
		statusLine:           status.New(utils.MainMenuContentWidth, utils.StatusLineHeight),
		width:                utils.MainMenuContentWidth,
		height:               utils.DefaultViewportHeight,
		ModelState:           utils.ModelStateLoaded,
	}

	return &manager, nil
}

func (m *Manager) SetCurrentSessionState(state utils.SessionState) {
	if m.CurrentSessionState == state {
		return
	}
	m.PreviousSessionState = m.CurrentSessionState
	m.CurrentSessionState = state
}

func (m *Manager) GetCurrentSessionState() utils.SessionState {
	return m.CurrentSessionState
}

func (m *Manager) GetModel(state utils.SessionState) TUIModel {
	return m.models[state]
}

func (m *Manager) GetModelState(state utils.SessionState) utils.ModelState {
	return m.models[state].GetModelState()
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case utils.SessionStateMsg:
		m.SetCurrentSessionState(msg.SessionState)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusLine.SetSize(msg.Width, utils.StatusLineHeight)
		for _, model := range m.models {
			model.SetSize(msg.Width, msg.Height-utils.StatusLineHeight)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.ForceQuit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.CursorUp):
			if m.CurrentSessionState == utils.SessionStateDetail {
				if detailModel, ok := m.models[utils.SessionStateDetail].(*detail.DetailModel); ok {
					detailModel.ScrollUp(utils.DefaultScrollSpeed)
				}
			}
		case key.Matches(msg, m.keyMap.CursorDown):
			if m.CurrentSessionState == utils.SessionStateDetail {
				if detailModel, ok := m.models[utils.SessionStateDetail].(*detail.DetailModel); ok {
					detailModel.ScrollDown(utils.DefaultScrollSpeed)
				}
			}

		case key.Matches(msg, m.keyMap.Edit):
			if m.CurrentSessionState == utils.SessionStateDetail {
				if detailModel, ok := m.models[utils.SessionStateDetail].(*detail.DetailModel); ok {
					cmds = append(cmds, utils.SendSessionStateMsg(utils.SessionStateEdit))
					cmds = append(cmds, utils.SendEditRecipeMsg(detailModel.CurrentRecipe.ID))
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keyMap.Back):
			if m.CurrentSessionState == utils.SessionStateList {
				if listModel, ok := m.models[utils.SessionStateList].(*yummy_list.ListModel); ok {
					if listModel.RecipeList.FilterState() != list.Filtering {
						m.SetCurrentSessionState(m.PreviousSessionState)
					}
				}
			} else {
				m.SetCurrentSessionState(m.PreviousSessionState)
			}

			return m, nil
		case key.Matches(msg, m.keyMap.Add):
			if m.CurrentSessionState == utils.SessionStateList {
				if listModel, ok := m.models[utils.SessionStateList].(*yummy_list.ListModel); ok {
					if listModel.RecipeList.FilterState() != list.Filtering {
						m.SetCurrentSessionState(m.PreviousSessionState)
					}
				}
			}
			return m, nil
		case key.Matches(msg, m.keyMap.StateSelector):
			m.SetCurrentSessionState(utils.SessionStateStateSelector)
			return m, nil
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
	var content string
	if currentModel, exists := m.models[m.CurrentSessionState]; exists {
		content = currentModel.View()
	}

	// Render status line if not loading
	if m.GetModelState(m.CurrentSessionState) != utils.ModelStateLoading {
		statusInfo := m.createStatusInfo()
		statusLine := m.statusLine.Render(statusInfo)
		content = lipgloss.JoinVertical(lipgloss.Left, content, statusLine)
	}

	return content
}

func (m *Manager) createStatusInfo() status.StatusInfo {
	additionalInfo := make(map[string]interface{})

	// Add specific information based on current session state
	switch m.CurrentSessionState {
	case utils.SessionStateList:
		if listModel, ok := m.models[utils.SessionStateList].(*yummy_list.ListModel); ok {
			additionalInfo["count"] = len(listModel.RecipeList.Items())
			additionalInfo["selected_item"] = listModel.RecipeList.SelectedItem().(recipe.RecipeWithDescription).Title()
		}

	case utils.SessionStateDetail:
		if detailModel, ok := m.models[utils.SessionStateDetail].(*detail.DetailModel); ok {
			if detailModel.CurrentRecipe != nil {
				recipeName := detailModel.CurrentRecipe.Name
				recipeID := detailModel.CurrentRecipe.ID
				author := detailModel.CurrentRecipe.Author
				if author != "" {
					author = fmt.Sprintf("(by %s)", author)
				}
				additionalInfo["recipe_name"] = strings.Join([]string{fmt.Sprintf("(#%d)", recipeID), recipeName, author}, " ")
			}
			// Add scroll position info
			additionalInfo["scroll_pos"] = detailModel.GetScrollPosition()
			additionalInfo["total_lines"] = detailModel.GetContentHeight()
		}

	case utils.SessionStateEdit:
		// Edit model doesn't expose recipe name directly, so we'll use a generic name
		additionalInfo["recipe_name"] = "Edit Recipe"

	case utils.SessionStateStateSelector:
		selectedState := m.models[utils.SessionStateStateSelector].(*dialog.StateSelectorDialogCmp).GetSelectedStateName()
		additionalInfo["state_selected"] = selectedState
	}

	return status.CreateStatusInfo(m.CurrentSessionState, additionalInfo)
}
