package main_menu

import (
	"github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type MainMenuModel struct {
	List list.Model
}

type menuItem struct {
	title string
	desc  string
	state ui.SessionState
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

func New() *MainMenuModel {
	items := []list.Item{
		menuItem{title: "Go to Cookbook List View", desc: "View your saved recipes", state: ui.SessionStateList},
		menuItem{title: "Fetch Random Recipe", desc: "Get a random recipe from the web", state: ui.SessionStateDetail},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Main Menu"

	return &MainMenuModel{List: l}
}

func (m *MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.List.SetSize(msg.Width, msg.Height)
	}
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *MainMenuModel) View() string {
	return m.List.View()
}
