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
	edit "github.com/GarroshIcecream/yummy/yummy/tui/edit"
	yummy_list "github.com/GarroshIcecream/yummy/yummy/tui/list"
	main_menu "github.com/GarroshIcecream/yummy/yummy/tui/main_menu"
	state_selector "github.com/GarroshIcecream/yummy/yummy/tui/state_selector"
	status "github.com/GarroshIcecream/yummy/yummy/tui/status"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/key"
	list "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TUIModel interface {
	tea.Model
	GetModelState() ui.ModelState
	SetSize(width, height int)
	GetSize() (width, height int)
}

type Manager struct {
	CurrentSessionState  ui.SessionState
	PreviousSessionState ui.SessionState
	models               map[ui.SessionState]TUIModel
	Cookbook             *db.CookBook
	keyMap               config.KeyMap
	ModelState           ui.ModelState
	Ctx                  context.Context
	statusLine           *status.StatusLine
	width                int
	height               int
}

func New(cookbook *db.CookBook, ctx context.Context) (*Manager, error) {
	keymaps := config.DefaultKeyMap()

	models := map[ui.SessionState]TUIModel{
		ui.SessionStateMainMenu:      main_menu.New(cookbook, keymaps),
		ui.SessionStateList:          yummy_list.New(cookbook, keymaps, false),
		ui.SessionStateDetail:        detail.New(cookbook, keymaps),
		ui.SessionStateEdit:          edit.New(cookbook, keymaps, nil),
		ui.SessionStateChat:          chat.New(cookbook, keymaps),
		ui.SessionStateStateSelector: state_selector.New(),
	}

	manager := Manager{
		Cookbook:             cookbook,
		keyMap:               keymaps,
		models:               models,
		CurrentSessionState:  ui.SessionStateMainMenu,
		PreviousSessionState: ui.SessionStateMainMenu,
		Ctx:                  ctx,
		statusLine:           status.New(ui.MainMenuContentWidth, ui.StatusLineHeight),
		width:                ui.MainMenuContentWidth,
		height:               ui.DefaultViewportHeight,
		ModelState:           ui.ModelStateLoaded,
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

func (m *Manager) GetModelState(state ui.SessionState) ui.ModelState {
	return m.models[state].GetModelState()
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case ui.SessionStateMsg:
		m.SetCurrentSessionState(msg.SessionState)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusLine.SetSize(msg.Width, ui.StatusLineHeight)
		for _, model := range m.models {
			model.SetSize(msg.Width, msg.Height-ui.StatusLineHeight)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.ForceQuit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.CursorUp):
			if m.CurrentSessionState == ui.SessionStateDetail {
				if detailModel, ok := m.models[ui.SessionStateDetail].(*detail.DetailModel); ok {
					detailModel.ScrollUp(ui.DefaultScrollSpeed)
				}
			}
		case key.Matches(msg, m.keyMap.CursorDown):
			if m.CurrentSessionState == ui.SessionStateDetail {
				if detailModel, ok := m.models[ui.SessionStateDetail].(*detail.DetailModel); ok {
					detailModel.ScrollDown(ui.DefaultScrollSpeed)
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
						return m, nil
					}
				}
			} else {
				m.SetCurrentSessionState(m.PreviousSessionState)
				return m, nil
			}
		case key.Matches(msg, m.keyMap.Add):
			if m.CurrentSessionState == ui.SessionStateList {
				if listModel, ok := m.models[ui.SessionStateList].(*yummy_list.ListModel); ok {
					if listModel.RecipeList.FilterState() != list.Filtering {
						m.SetCurrentSessionState(m.PreviousSessionState)
					}
				}
			}
		case key.Matches(msg, m.keyMap.StateSelector):
			m.SetCurrentSessionState(ui.SessionStateStateSelector)
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
	if m.GetModelState(m.CurrentSessionState) != ui.ModelStateLoading {
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
	case ui.SessionStateList:
		if listModel, ok := m.models[ui.SessionStateList].(*yummy_list.ListModel); ok {
			additionalInfo["count"] = len(listModel.RecipeList.Items())
			additionalInfo["selected_item"] = listModel.RecipeList.SelectedItem().(recipe.RecipeWithDescription).Title()
		}

	case ui.SessionStateDetail:
		if detailModel, ok := m.models[ui.SessionStateDetail].(*detail.DetailModel); ok {
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

	case ui.SessionStateEdit:
		// Edit model doesn't expose recipe name directly, so we'll use a generic name
		additionalInfo["recipe_name"] = "Edit Recipe"

	case ui.SessionStateStateSelector:
		selectedState := m.models[ui.SessionStateStateSelector].(*state_selector.StateSelectorDialogCmp).GetSelectedStateName()
		additionalInfo["state_selected"] = selectedState
	}

	return status.CreateStatusInfo(m.CurrentSessionState, additionalInfo)
}
