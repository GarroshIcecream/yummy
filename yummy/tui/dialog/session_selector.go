package dialog

import (
	"fmt"
	"time"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SessionItem struct {
	SessionID         uint
	CreatedAt         time.Time
	UpdatedAt         time.Time
	MessageCount      int
	TotalInputTokens  int
	TotalOutputTokens int
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

type SessionSelectorDialogCmp struct {
	sessionLog *db.SessionLog
	keyMap     config.KeyMap
	list       list.Model
	width      int
	height     int
	theme      *themes.Theme
}

func NewSessionSelectorDialog(sessionLog *db.SessionLog, keymaps config.KeyMap, theme *themes.Theme, config *config.SessionSelectorDialogConfig) *SessionSelectorDialogCmp {
	items := []list.Item{}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "📚 Select Previous Session"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = theme.SessionSelectorTitle
	l.Styles.PaginationStyle = theme.SessionSelectorPagination
	l.Styles.HelpStyle = theme.SessionSelectorHelp

	return &SessionSelectorDialogCmp{
		sessionLog: sessionLog,
		keyMap:     keymaps,
		list:       l,
		width:      config.Width,
		height:     config.Height,
		theme:      theme,
	}
}

func (m *SessionSelectorDialogCmp) Init() tea.Cmd {
	return nil
}

func (m *SessionSelectorDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Back, m.keyMap.Quit):
			cmds = append(cmds, messages.SendSessionStateMsg(common.SessionStateChat))
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keyMap.SessionSelector):
			cmds = append(cmds, messages.SendSessionStateMsg(common.SessionStateChat))
			return m, tea.Batch(cmds...)
		case key.Matches(msg, m.keyMap.Enter):
			if selectedItem, ok := m.list.SelectedItem().(SessionItem); ok {
				cmds = append(cmds, tea.Sequence(
					messages.SendSessionStateMsg(common.SessionStateChat),
					messages.SendSessionSelectedMsg(selectedItem.SessionID),
				))
				return m, tea.Batch(cmds...)
			}
		}

	case messages.LoadSessionsMsg:
		m.loadSessions()
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *SessionSelectorDialogCmp) View() string {
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
	centeredDialogStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Width(m.width).
		Height(m.height)

	return centeredDialogStyle.Render(boxStyle.Render(content))
}

func (m *SessionSelectorDialogCmp) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetWidth(width - 4)
	m.list.SetHeight(height - 6)
}

func (m *SessionSelectorDialogCmp) GetSize() (int, int) {
	return m.width, m.height
}

func (m *SessionSelectorDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}

func (m *SessionSelectorDialogCmp) loadSessions() {
	sessions, err := m.sessionLog.GetNonEmptySessions()
	if err != nil {
		m.list.SetItems([]list.Item{})
		return
	}

	items := make([]list.Item, len(sessions))
	for i, session := range sessions {
		stats, err := m.sessionLog.GetSessionStats(session.ID)
		if err != nil {
			stats = db.SessionStats{}
		}

		items[i] = SessionItem{
			SessionID:         session.ID,
			CreatedAt:         session.CreatedAt,
			UpdatedAt:         session.UpdatedAt,
			MessageCount:      stats.MessageCount,
			TotalInputTokens:  stats.InputTokens,
			TotalOutputTokens: stats.OutputTokens,
		}
	}

	m.list.SetItems(items)
}
