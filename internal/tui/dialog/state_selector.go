package dialog

import (
	"fmt"
	"strconv"
	"strings"

	config "github.com/GarroshIcecream/yummy/internal/config"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StateSelectorDialogCmp struct {
	// State data
	states        []common.SessionState
	selectedIndex int

	// Config
	keymap config.StateSelectorKeyMap
	theme  *themes.Theme

	// UI
	height int
	width  int
}

// NewStateSelectorDialog creates a new state selection dialog.
func NewStateSelectorDialog(theme *themes.Theme) (*StateSelectorDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetStateSelectorKeyMap()
	stateSelectorConfig := cfg.StateSelectorDialog
	states := []common.SessionState{
		common.SessionStateMainMenu,
		common.SessionStateList,
		common.SessionStateDetail,
		common.SessionStateEdit,
		common.SessionStateChat,
	}

	return &StateSelectorDialogCmp{
		selectedIndex: 0,
		states:        states,
		height:        stateSelectorConfig.Height,
		width:         stateSelectorConfig.Width,
		keymap:        keymaps,
		theme:         theme,
	}, nil
}

func (s *StateSelectorDialogCmp) Init() tea.Cmd {
	return nil
}

func (s *StateSelectorDialogCmp) GetStateIndexFromNumberKey(msg tea.KeyMsg) *int {
	keyStr := msg.String()
	i, err := strconv.Atoi(keyStr)
	if err != nil {
		return nil
	}

	if i >= 1 && i <= len(s.states) {
		s.selectedIndex = i - 1
		return &s.selectedIndex
	}

	return nil
}

// Update handles keyboard input for the state selector dialog.
func (s *StateSelectorDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keymap.CursorUp):
			if s.selectedIndex > 0 {
				s.selectedIndex--
			} else {
				s.selectedIndex = len(s.states) - 1
			}
		case key.Matches(msg, s.keymap.CursorDown):
			if s.selectedIndex < len(s.states)-1 {
				s.selectedIndex++
			} else {
				s.selectedIndex = 0
			}
		// Handle number keys for direct state selection
		case s.GetStateIndexFromNumberKey(msg) != nil:
			s.selectedIndex = *s.GetStateIndexFromNumberKey(msg)
			selectedState := s.states[s.selectedIndex]
			cmds = append(cmds, tea.Batch(
				messages.SendSessionStateMsg(selectedState),
				messages.SendCloseDialogMsg(),
			))
		case key.Matches(msg, s.keymap.Enter):
			selectedState := s.states[s.selectedIndex]
			cmds = append(cmds, tea.Batch(
				messages.SendSessionStateMsg(selectedState),
				messages.SendCloseDialogMsg(),
			))
		}
	}

	return s, tea.Batch(cmds...)
}

// View renders the state selector dialog with a list of states.
func (s *StateSelectorDialogCmp) View() string {
	// Header: title left, esc hint right
	titleLeft := s.theme.StateSelectorTitle.Render("Switch View")
	escHint := s.theme.StateSelectorHelp.Render("esc")
	innerWidth := s.width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}
	pad := innerWidth - lipgloss.Width(titleLeft) - lipgloss.Width(escHint)
	if pad < 1 {
		pad = 1
	}
	header := titleLeft + strings.Repeat(" ", pad) + escHint

	// Separator
	sep := s.theme.StateSelectorHelp.Render(strings.Repeat("─", innerWidth))

	// Items — tight single-line rows
	var rows []string
	for i, state := range s.states {
		stateName := state.GetStateName()
		numPrefix := fmt.Sprintf("%d  ", i+1)
		line := numPrefix + stateName

		if i == s.selectedIndex {
			row := s.theme.StateSelectorSelectedItem.
				Width(innerWidth).
				Render(line)
			rows = append(rows, row)
		} else {
			rows = append(rows, s.theme.StateSelectorItem.Render(line))
		}
	}

	parts := []string{header, sep}
	parts = append(parts, rows...)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	dialogBox := s.theme.StateSelectorDialog.Render(content)
	return dialogBox
}

func (s *StateSelectorDialogCmp) SetSize(width, height int) {
	s.width = width
	s.height = height
}

func (s *StateSelectorDialogCmp) GetSize() (int, int) {
	return s.width, s.height
}

func (s *StateSelectorDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}

func (s *StateSelectorDialogCmp) GetSelectedState() common.SessionState {
	return s.states[s.selectedIndex]
}
