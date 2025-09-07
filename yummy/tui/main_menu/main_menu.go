package main_menu

import (
	"fmt"
	"strings"

	"github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuModel struct {
	items       []menuItem
	selected    int
	width       int
	height      int
	keyMap      KeyMap
	showHelp    bool
}

type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Help   key.Binding
	Quit   key.Binding
}

type menuItem struct {
	title       string
	description string
	state       ui.SessionState
	icon        string
}

func New() *MainMenuModel {
	items := []menuItem{
		{
			title:       "Go to Cookbook List View",
			description: "View your saved recipes",
			state:       ui.SessionStateList,
			icon:        "📚",
		},
		{
			title:       "Fetch Random Recipe",
			description: "Get a random recipe from the web",
			state:       ui.SessionStateDetail,
			icon:        "🎲",
		},
		{
			title:       "Chat with AI Assistant",
			description: "Ask questions about cooking and recipes",
			state:       ui.SessionStateChat,
			icon:        "🤖",
		},
	}

	return &MainMenuModel{
		items:    items,
		selected: 0,
		keyMap:   defaultKeyMap(),
		showHelp: false,
	}
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "move down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?/h", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (m *MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			if m.selected > 0 {
				m.selected--
			}
			return m, nil

		case key.Matches(msg, m.keyMap.Down):
			if m.selected < len(m.items)-1 {
				m.selected++
			}
			return m, nil

		case key.Matches(msg, m.keyMap.Enter):
			if len(m.items) > 0 {
				selectedItem := m.items[m.selected]
				return m, ui.SendSessionStateMsg(selectedItem.state)
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *MainMenuModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// Title
	title := m.renderTitle()
	content.WriteString(title)
	content.WriteString("\n\n")

	// Menu items
	menuContent := m.renderMenuItems()
	content.WriteString(menuContent)

	// Help section
	if m.showHelp {
		content.WriteString("\n")
		content.WriteString(m.renderHelp())
	}

	// Center the content
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(content.String())
}

func (m *MainMenuModel) renderTitle() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B6B")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(1, 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B"))

	return titleStyle.Render("🍳 Yummy Recipe Manager")
}

func (m *MainMenuModel) renderMenuItems() string {
	var items strings.Builder

	for i, item := range m.items {
		itemStyle := m.getItemStyle(i == m.selected)
		itemContent := m.renderMenuItem(item, i == m.selected)
		items.WriteString(itemStyle.Render(itemContent))
		items.WriteString("\n")
	}

	return items.String()
}

func (m *MainMenuModel) renderMenuItem(item menuItem, isSelected bool) string {
	var content strings.Builder

	// Bullet point with icon
	bullet := "•"
	if isSelected {
		bullet = "▶"
	}

	bulletStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(isSelected)

	content.WriteString(bulletStyle.Render(bullet))
	content.WriteString(" ")

	// Icon
	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#98FB98"))

	content.WriteString(iconStyle.Render(item.icon))
	content.WriteString(" ")

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(isSelected).
		Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(titleStyle.Render(item.title))
	content.WriteString("\n")

	// Description
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB")).
		Italic(true).
		PaddingLeft(3)

	content.WriteString(descStyle.Render(item.description))

	return content.String()
}

func (m *MainMenuModel) getItemStyle(isSelected bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Padding(1, 2).
		Margin(0, 1)

	if isSelected {
		style = style.
			Background(lipgloss.Color("#2a2a2a")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6B6B"))
	}

	return style
}

func (m *MainMenuModel) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87CEEB")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#87CEEB")).
		Padding(1, 2)

	helpText := fmt.Sprintf(
		"Navigation:\n%s • %s • %s • %s",
		m.keyMap.Up.Help().Key,
		m.keyMap.Down.Help().Key,
		m.keyMap.Enter.Help().Key,
		m.keyMap.Quit.Help().Key,
	)

	return helpStyle.Render(helpText)
}

// SetSize sets the width and height of the model
func (m *MainMenuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetSize returns the current width and height of the model
func (m *MainMenuModel) GetSize() (width, height int) {
	return m.width, m.height
}