package manager

import (
	"log"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/GarroshIcecream/yummy/yummy/keymaps"
	"github.com/GarroshIcecream/yummy/yummy/models/chat"
	"github.com/GarroshIcecream/yummy/yummy/models/detail"
	"github.com/GarroshIcecream/yummy/yummy/models/edit"
	yummy_list "github.com/GarroshIcecream/yummy/yummy/models/list"
	"github.com/charmbracelet/bubbles/key"
	list "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type SessionState int

const (
	SessionStateList SessionState = iota
	SessionStateDetail
	SessionStateEdit
	SessionStateChat
)

type Manager struct {
	SessionState SessionState
	ListModel    *yummy_list.ListModel
	DetailModel  *detail.DetailModel
	EditModel    *edit.EditModel
	ChatModel    *chat.ChatModel
	Cookbook     *db.CookBook
}

func New(cookbook *db.CookBook) (*Manager, error) {
	manager := Manager{
		Cookbook:     cookbook,
		ListModel:    yummy_list.New(cookbook, nil),
		DetailModel:  detail.New(cookbook, nil),
		EditModel:    edit.New(cookbook, nil, nil),
		ChatModel:    chat.New(cookbook),
		SessionState: SessionStateList,
	}

	return &manager, nil
}

func (m *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keymaps.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keymaps.Keys.Back):
			switch m.SessionState {
			case SessionStateList:
				if m.ListModel.RecipeList.FilterState() != list.Filtering {
					return m, tea.Quit
				}
			case SessionStateDetail:
				m.SessionState = SessionStateList
			case SessionStateEdit:
				m.SessionState = SessionStateDetail
			case SessionStateChat:
				m.SessionState = SessionStateList
			}
		case key.Matches(msg, keymaps.Keys.Add):
			if m.SessionState == SessionStateList {
				if m.ListModel.RecipeList.FilterState() != list.Filtering {
					m.SessionState = SessionStateChat
				}
			}
		}

		switch m.SessionState {
		case SessionStateList:
			var model tea.Model
			model, cmd = m.ListModel.Update(msg)
			listModel, ok := model.(*yummy_list.ListModel)
			if !ok {
				log.Println("ListModel is not a yummy_list.ListModel")
			}
			m.ListModel = listModel
		case SessionStateDetail:
			var model tea.Model
			model, cmd = m.DetailModel.Update(msg)
			detailModel, ok := model.(*detail.DetailModel)
			if !ok {
				log.Println("DetailModel is not a detail.DetailModel")
			}
			m.DetailModel = detailModel
		case SessionStateEdit:
			var model tea.Model
			model, cmd = m.EditModel.Update(msg)
			editModel, ok := model.(*edit.EditModel)
			if !ok {
				log.Println("EditModel is not a edit.EditModel")
			}
			m.EditModel = editModel
		case SessionStateChat:
			var model tea.Model
			model, cmd = m.ChatModel.Update(msg)
			chatModel, ok := model.(*chat.ChatModel)
			if !ok {
				log.Println("ChatModel is not a chat.ChatModel")
			}
			m.ChatModel = chatModel
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Manager) Init() tea.Cmd {
	switch m.SessionState {
	case SessionStateList:
		return m.ListModel.Init()
	case SessionStateDetail:
		return m.DetailModel.Init()
	case SessionStateEdit:
		return m.EditModel.Init()
	case SessionStateChat:
		return m.ChatModel.Init()
	}
	return nil
}

func (m Manager) View() string {
	switch m.SessionState {
	case SessionStateList:
		return m.ListModel.View()
	case SessionStateDetail:
		return m.DetailModel.View()
	case SessionStateEdit:
		return m.EditModel.View()
	case SessionStateChat:
		return m.ChatModel.View()
	}
	return ""
}
