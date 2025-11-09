package dialog

import (
	"fmt"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	utils "github.com/GarroshIcecream/yummy/yummy/utils"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SessionSelectorDialogCmp struct {
	sessionLog *db.SessionLog
	keyMap     config.SessionSelectorKeyMap
	list       list.Model
	width      int
	height     int
	theme      *themes.Theme
}

func NewSessionSelectorDialog(sessionLog *db.SessionLog, theme *themes.Theme, currentSessionID uint) (*SessionSelectorDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetSessionSelectorKeyMap()
	sessionSelectorConfig := cfg.SessionSelectorDialog
	sessions, err := sessionLog.GetNonEmptySessions()
	if err != nil {
		return nil, err
	}

	items := make([]list.Item, len(sessions))
	for i, session := range sessions {
		if session.SessionID == currentSessionID {
			session.Selected = true
		}
		items[i] = session
	}

	l := list.New(items, list.NewDefaultDelegate(), 80, 24)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = theme.SessionSelectorTitle
	l.Styles.PaginationStyle = theme.SessionSelectorPagination
	l.Styles.HelpStyle = theme.SessionSelectorHelp

	return &SessionSelectorDialogCmp{
		sessionLog: sessionLog,
		keyMap:     keymaps,
		list:       l,
		width:      sessionSelectorConfig.Width,
		height:     sessionSelectorConfig.Height,
		theme:      theme,
	}, nil
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
		case key.Matches(msg, m.keyMap.SessionSelector):
			cmds = append(cmds, messages.SendCloseModalViewMsg())

		case key.Matches(msg, m.keyMap.Enter):
			if selectedItem, ok := m.list.SelectedItem().(*utils.SessionItem); ok {
				cmds = append(cmds, messages.SendSessionStateMsg(common.SessionStateChat))
				cmds = append(cmds, messages.SendSessionSelectedMsg(selectedItem.SessionID))
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Sequence(cmds...)
}

func (m *SessionSelectorDialogCmp) View() string {
	content := m.list.View()
	helpText := m.theme.SessionSelectorHelp.Render("↑/↓ Navigate • Enter Select • Esc Cancel")
	content = m.theme.SessionSelectorDialog.
		Width(m.width).
		Height(m.height).
		Render(lipgloss.JoinVertical(lipgloss.Left, content, helpText))

	return m.theme.SessionSelectorContainer.Render(content)
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
