package dialog

import (
	"fmt"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxVisibleSessions = 8

type sessionEntry struct {
	SessionID uint
	Title     string
	Desc      string
	Selected  bool
}

type SessionSelectorDialogCmp struct {
	allItems      []sessionEntry
	filtered      []sessionEntry
	selectedIndex int
	scrollOffset  int
	searchInput   textinput.Model
	width         int
	height        int
	theme         *themes.Theme
}

func NewSessionSelectorDialog(sessionLog *db.SessionLog, theme *themes.Theme, currentSessionID uint) (*SessionSelectorDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	sessionSelectorConfig := cfg.SessionSelectorDialog
	sessions, err := sessionLog.GetNonEmptySessions()
	if err != nil {
		return nil, err
	}

	items := make([]sessionEntry, len(sessions))
	for i, s := range sessions {
		title := s.Title()
		desc := s.Description()
		items[i] = sessionEntry{
			SessionID: s.SessionID,
			Title:     title,
			Desc:      desc,
			Selected:  s.SessionID == currentSessionID,
		}
	}

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 64
	if w := sessionSelectorConfig.Width - 8; w > 10 {
		ti.Width = w
	} else {
		ti.Width = 40
	}

	return &SessionSelectorDialogCmp{
		allItems:    items,
		filtered:    items,
		searchInput: ti,
		width:       sessionSelectorConfig.Width,
		height:      sessionSelectorConfig.Height,
		theme:       theme,
	}, nil
}

func (m *SessionSelectorDialogCmp) Init() tea.Cmd {
	return textinput.Blink
}

func (m *SessionSelectorDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmds = append(cmds, messages.SendCloseModalViewMsg())
			return m, tea.Batch(cmds...)

		case "enter":
			if len(m.filtered) > 0 && m.selectedIndex < len(m.filtered) {
				selected := m.filtered[m.selectedIndex]
				cmds = append(cmds,
					messages.SendCloseModalViewMsg(),
					messages.SendSessionStateMsg(common.SessionStateChat),
					messages.SendSessionSelectedMsg(selected.SessionID),
				)
			}
			return m, tea.Batch(cmds...)

		case "up", "ctrl+k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.ensureVisible()
			}
			return m, nil

		case "down", "ctrl+j":
			if m.selectedIndex < len(m.filtered)-1 {
				m.selectedIndex++
				m.ensureVisible()
			}
			return m, nil
		}
	}

	prevValue := m.searchInput.Value()
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	cmds = append(cmds, cmd)

	if m.searchInput.Value() != prevValue {
		m.applyFilter()
	}

	return m, tea.Batch(cmds...)
}

func (m *SessionSelectorDialogCmp) ensureVisible() {
	if m.selectedIndex < m.scrollOffset {
		m.scrollOffset = m.selectedIndex
	}
	if m.selectedIndex >= m.scrollOffset+maxVisibleSessions {
		m.scrollOffset = m.selectedIndex - maxVisibleSessions + 1
	}
}

func (m *SessionSelectorDialogCmp) applyFilter() {
	query := strings.ToLower(m.searchInput.Value())
	if query == "" {
		m.filtered = m.allItems
	} else {
		m.filtered = nil
		for _, item := range m.allItems {
			if strings.Contains(strings.ToLower(item.Title), query) ||
				strings.Contains(strings.ToLower(item.Desc), query) {
				m.filtered = append(m.filtered, item)
			}
		}
	}
	m.selectedIndex = 0
	m.scrollOffset = 0
}

func (m *SessionSelectorDialogCmp) View() string {
	innerWidth := m.width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}

	// Header
	titleLeft := m.theme.SessionSelectorTitle.Render("Sessions")
	escHint := m.theme.SessionSelectorHelp.Render("esc")
	pad := innerWidth - lipgloss.Width(titleLeft) - lipgloss.Width(escHint)
	if pad < 1 {
		pad = 1
	}
	header := titleLeft + strings.Repeat(" ", pad) + escHint

	// Search
	searchLine := m.searchInput.View()

	// Separator
	sep := m.theme.SessionSelectorHelp.Render(strings.Repeat("─", innerWidth))

	// Visible window
	end := m.scrollOffset + maxVisibleSessions
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	maxTitleLen := innerWidth - 4
	if maxTitleLen < 10 {
		maxTitleLen = 10
	}

	var rows []string

	if m.scrollOffset > 0 {
		rows = append(rows, m.theme.SessionSelectorHelp.Render(
			strings.Repeat(" ", innerWidth/2-1)+"▲"))
	}

	visible := m.filtered[m.scrollOffset:end]
	for i, item := range visible {
		globalIdx := m.scrollOffset + i

		marker := "  "
		if item.Selected {
			marker = "* "
		}

		title := item.Title
		if lipgloss.Width(title) > maxTitleLen-2 {
			title = title[:maxTitleLen-3] + "…"
		}
		line := marker + title

		// Summary line
		desc := item.Desc
		maxDescLen := innerWidth - 6
		if maxDescLen > 0 && lipgloss.Width(desc) > maxDescLen {
			desc = desc[:maxDescLen-1] + "…"
		}
		descLine := "    " + desc

		if globalIdx == m.selectedIndex {
			titleRow := m.theme.DialogSelectedRow.
				Width(innerWidth).
				MaxWidth(innerWidth).
				Render(line)
			descRow := m.theme.SessionSelectorSelectedDesc.
				Width(innerWidth).
				MaxWidth(innerWidth).
				Render(descLine)
			rows = append(rows, titleRow+"\n"+descRow)
		} else {
			titleRow := m.theme.DialogUnselectedRow.
				Render(line)
			descRow := m.theme.SessionSelectorUnselectedDesc.
				Render(descLine)
			rows = append(rows, titleRow+"\n"+descRow)
		}
	}

	if end < len(m.filtered) {
		rows = append(rows, m.theme.SessionSelectorHelp.Render(
			strings.Repeat(" ", innerWidth/2-1)+"▼"))
	}

	if len(m.filtered) == 0 {
		rows = append(rows, m.theme.SessionSelectorHelp.Render("No matching sessions"))
	}

	parts := []string{header, "", searchLine, sep}
	parts = append(parts, rows...)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	rendered := m.theme.SessionSelectorDialog.
		Width(m.width).
		Render(content)

	return m.theme.SessionSelectorContainer.Render(rendered)
}

func (m *SessionSelectorDialogCmp) SetSize(width, height int) {
	m.width = width
	m.height = height
	if w := m.width - 8; w > 10 {
		m.searchInput.Width = w
	}
}

func (m *SessionSelectorDialogCmp) GetSize() (int, int) {
	return m.width, m.height
}

func (m *SessionSelectorDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}
