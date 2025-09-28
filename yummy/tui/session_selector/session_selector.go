package session_selector

import (
	"fmt"
	"time"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SessionItem struct {
	SessionID         uint
	CreatedAt         time.Time
	UpdatedAt         time.Time
	MessageCount      int64
	TotalInputTokens  int64
	TotalOutputTokens int64
}

func (s SessionItem) Title() string {
	created := s.CreatedAt.Format("Jan 2, 15:04")
	messageCount := s.MessageCount
	if messageCount == 0 {
		return fmt.Sprintf("Session #%d - %s (No messages)", s.SessionID, created)
	}
	return fmt.Sprintf("Session #%d - %s (%d messages)", s.SessionID, created, messageCount)
}

func (s SessionItem) Description() string {
	if s.MessageCount == 0 {
		return "Empty session"
	}
	totalTokens := s.TotalInputTokens + s.TotalOutputTokens
	return fmt.Sprintf("Tokens: %d | Last updated: %s", totalTokens, s.UpdatedAt.Format("Jan 2, 15:04"))
}

func (s SessionItem) FilterValue() string {
	return fmt.Sprintf("session %d", s.SessionID)
}

type SessionSelectorModel struct {
	cookbook *db.CookBook
	keyMap   config.KeyMap
	list     list.Model
	visible  bool
	width    int
	height   int
	onSelect func(uint) tea.Cmd
	onCancel func() tea.Cmd
}

func New(cookbook *db.CookBook, keymaps config.KeyMap) *SessionSelectorModel {
	items := []list.Item{}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "📚 Select Previous Session"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true).
		Margin(1, 0, 1, 2)

	l.Styles.PaginationStyle = lipgloss.NewStyle().
		MarginLeft(2)

	l.Styles.HelpStyle = lipgloss.NewStyle().
		MarginLeft(2).
		Foreground(lipgloss.Color("#626262"))

	return &SessionSelectorModel{
		cookbook: cookbook,
		keyMap:   keymaps,
		list:     l,
		visible:  false,
		width:    60,
		height:   20,
	}
}

func (m *SessionSelectorModel) Init() tea.Cmd {
	return nil
}

func (m *SessionSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Back):
			if m.visible {
				m.visible = false
				if m.onCancel != nil {
					cmds = append(cmds, m.onCancel())
				}
				return m, tea.Batch(cmds...)
			}
		case key.Matches(msg, m.keyMap.Enter):
			if m.visible {
				if selectedItem, ok := m.list.SelectedItem().(SessionItem); ok {
					m.visible = false
					if m.onSelect != nil {
						cmds = append(cmds, m.onSelect(selectedItem.SessionID))
					}
					return m, tea.Batch(cmds...)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(msg.Height - 6)

	case LoadSessionsMsg:
		m.loadSessions()
	}

	if m.visible {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *SessionSelectorModel) View() string {
	if !m.visible {
		return ""
	}

	// Create a centered dialog box
	dialogWidth := m.width - 8
	if dialogWidth > 80 {
		dialogWidth = 80
	}

	dialogHeight := m.height - 8
	if dialogHeight > 24 {
		dialogHeight = 24
	}

	// Create the dialog box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#04B575")).
		Padding(1, 2).
		Width(dialogWidth).
		Height(dialogHeight)

	content := m.list.View()

	// Add help text at the bottom
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("↑/↓ Navigate • Enter Select • Esc Cancel")

	content = lipgloss.JoinVertical(lipgloss.Left, content, helpText)

	// Create a full-screen container that centers the dialog
	// Use Place to center the dialog in the middle of the screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		boxStyle.Render(content),
	)
}

func (m *SessionSelectorModel) Show() {
	m.visible = true
	m.loadSessions()
}

func (m *SessionSelectorModel) Hide() {
	m.visible = false
}

func (m *SessionSelectorModel) IsVisible() bool {
	return m.visible
}

func (m *SessionSelectorModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetWidth(width - 4)
	m.list.SetHeight(height - 6)
}

func (m *SessionSelectorModel) SetOnSelect(callback func(uint) tea.Cmd) {
	m.onSelect = callback
}

func (m *SessionSelectorModel) SetOnCancel(callback func() tea.Cmd) {
	m.onCancel = callback
}

func (m *SessionSelectorModel) loadSessions() {
	sessions, err := m.cookbook.GetNonEmptySessions()
	if err != nil {
		// If there's an error, show an empty list
		m.list.SetItems([]list.Item{})
		return
	}

	items := make([]list.Item, len(sessions))
	for i, session := range sessions {
		// Get stats for each session
		messageCount, inputTokens, outputTokens, err := m.cookbook.GetSessionStats(session.ID)
		if err != nil {
			// If we can't get stats, use zeros
			messageCount = 0
			inputTokens = 0
			outputTokens = 0
		}

		items[i] = SessionItem{
			SessionID:         session.ID,
			CreatedAt:         session.CreatedAt,
			UpdatedAt:         session.UpdatedAt,
			MessageCount:      messageCount,
			TotalInputTokens:  inputTokens,
			TotalOutputTokens: outputTokens,
		}
	}

	m.list.SetItems(items)
}

// LoadSessionsMsg is a message to trigger loading sessions
type LoadSessionsMsg struct{}

func SendLoadSessionsMsg() tea.Cmd {
	return func() tea.Msg {
		return LoadSessionsMsg{}
	}
}
