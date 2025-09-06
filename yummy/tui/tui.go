package tui

import (
	"log"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	chat "github.com/GarroshIcecream/yummy/yummy/tui/chat"
	detail "github.com/GarroshIcecream/yummy/yummy/tui/detail"
	edit "github.com/GarroshIcecream/yummy/yummy/tui/edit"
	yummy_list "github.com/GarroshIcecream/yummy/yummy/tui/list"
	"github.com/GarroshIcecream/yummy/yummy/tui/main_menu"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/key"
	list "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Manager struct {
	SessionState  ui.SessionState
	MainMenuModel *main_menu.MainMenuModel
	ListModel     *yummy_list.ListModel
	DetailModel   *detail.DetailModel
	EditModel     *edit.EditModel
	ChatModel     *chat.ChatModel
	Cookbook      *db.CookBook
	keyMap        KeyMap
}

func New(cookbook *db.CookBook) (*Manager, error) {
	keymaps := DefaultKeyMap()
	manager := Manager{
		Cookbook:      cookbook,
		keyMap:        keymaps,
		ListModel:     yummy_list.New(cookbook),
		DetailModel:   detail.New(cookbook),
		EditModel:     edit.New(cookbook, nil),
		ChatModel:     chat.New(cookbook),
		MainMenuModel: main_menu.New(),
		SessionState:  ui.SessionStateMainMenu,
	}

	return &manager, nil
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case ui.SessionStateMsg:
		m.SessionState = msg.SessionState

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Up):
			if m.SessionState == ui.SessionStateDetail {
				m.DetailModel.ScrollUp(3)
			}
		case key.Matches(msg, m.keyMap.Down):
			if m.SessionState == ui.SessionStateDetail {
				m.DetailModel.ScrollDown(3)
			}

		case key.Matches(msg, m.keyMap.Edit):
			if m.SessionState == ui.SessionStateDetail {
				cmds = append(cmds, ui.SendSessionStateMsg(ui.SessionStateEdit))
				cmds = append(cmds, ui.SendEditRecipeMsg(m.DetailModel.CurrentRecipe.ID))
			}

		case key.Matches(msg, m.keyMap.Back):
			switch m.SessionState {
			case ui.SessionStateList:
				if m.ListModel.RecipeList.FilterState() != list.Filtering {
					m.SessionState = ui.SessionStateMainMenu
					return m, nil
				}
			case ui.SessionStateDetail:
				m.SessionState = ui.SessionStateList
				return m, nil
			case ui.SessionStateEdit:
				m.SessionState = ui.SessionStateDetail
				return m, nil
			case ui.SessionStateChat:
				m.SessionState = ui.SessionStateList
				return m, nil
			}
		case key.Matches(msg, m.keyMap.Add):
			if m.SessionState == ui.SessionStateList {
				if m.ListModel.RecipeList.FilterState() != list.Filtering {
					m.SessionState = ui.SessionStateChat
				}
			}
		}
	}

	switch m.SessionState {
	case ui.SessionStateMainMenu:
		var model tea.Model
		model, cmd = m.MainMenuModel.Update(msg)
		mainMenuModel, ok := model.(*main_menu.MainMenuModel)
		if !ok {
			log.Println("MainMenuModel is not a main_menu.MainMenuModel")
		}
		m.MainMenuModel = mainMenuModel
	case ui.SessionStateList:
		var model tea.Model
		model, cmd = m.ListModel.Update(msg)
		listModel, ok := model.(*yummy_list.ListModel)
		if !ok {
			log.Println("ListModel is not a yummy_list.ListModel")
		}
		m.ListModel = listModel
	case ui.SessionStateDetail:
		var model tea.Model
		model, cmd = m.DetailModel.Update(msg)
		detailModel, ok := model.(*detail.DetailModel)
		if !ok {
			log.Println("DetailModel is not a detail.DetailModel")
		}
		m.DetailModel = detailModel
	case ui.SessionStateEdit:
		var model tea.Model
		model, cmd = m.EditModel.Update(msg)
		editModel, ok := model.(*edit.EditModel)
		if !ok {
			log.Println("EditModel is not a edit.EditModel")
		}
		m.EditModel = editModel
	case ui.SessionStateChat:
		var model tea.Model
		model, cmd = m.ChatModel.Update(msg)
		chatModel, ok := model.(*chat.ChatModel)
		if !ok {
			log.Println("ChatModel is not a chat.ChatModel")
		}
		m.ChatModel = chatModel
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Manager) Init() tea.Cmd {
	switch m.SessionState {
	case ui.SessionStateMainMenu:
		return m.MainMenuModel.Init()
	case ui.SessionStateList:
		return m.ListModel.Init()
	case ui.SessionStateDetail:
		return m.DetailModel.Init()
	case ui.SessionStateEdit:
		return m.EditModel.Init()
	case ui.SessionStateChat:
		return m.ChatModel.Init()
	}
	return nil
}

func (m Manager) View() string {
	switch m.SessionState {
	case ui.SessionStateMainMenu:
		return m.MainMenuModel.View()
	case ui.SessionStateList:
		return m.ListModel.View()
	case ui.SessionStateDetail:
		return m.DetailModel.View()
	case ui.SessionStateEdit:
		return m.EditModel.View()
	case ui.SessionStateChat:
		return m.ChatModel.View()
	}
	return ""
}
