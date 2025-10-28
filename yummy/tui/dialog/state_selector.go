package dialog

import (
	"fmt"
	"strconv"

	config "github.com/GarroshIcecream/yummy/yummy/config"
	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StateSelectorDialogCmp struct {
	// State data
	states        []consts.SessionState
	selectedIndex int

	// Config
	keymap config.KeyMap
	theme  *themes.Theme

	// UI
	height int
	width  int
}

// NewStateSelectorDialog creates a new state selection dialog.
func NewStateSelectorDialog(theme *themes.Theme, config *config.StateSelectorDialogConfig, keymaps config.KeyMap) *StateSelectorDialogCmp {
	states := []consts.SessionState{
		consts.SessionStateMainMenu,
		consts.SessionStateList,
		consts.SessionStateDetail,
		consts.SessionStateEdit,
		consts.SessionStateChat,
	}

	return &StateSelectorDialogCmp{
		selectedIndex: 0,
		states:        states,
		height:        config.Height,
		width:         config.Width,
		keymap:        keymaps,
		theme:         theme,
	}
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
	title := "🎯 Select State"

	// Create the list of states
	var stateItems []string
	for i, state := range s.states {
		// get state name and emoji from mapping
		stateName := state.GetStateName()

		var style lipgloss.Style
		if i == s.selectedIndex {
			style = s.theme.StateSelectorSelectedItem
		} else {
			style = s.theme.StateSelectorItem
		}

		// Add number prefix and emoji
		numberPrefix := fmt.Sprintf("%d. ", i+1)
		indicator := "  "
		if i == s.selectedIndex {
			indicator = state.GetStateEmoji() + " "
		}

		stateItems = append(stateItems, style.Render(numberPrefix+indicator+stateName))
	}

	// Create the dialog box with better styling
	content := lipgloss.JoinVertical(lipgloss.Left, stateItems...)
	fullContent := lipgloss.JoinVertical(
		lipgloss.Center,
		s.theme.StateSelectorTitle.Render(title),
		content,
	)

	dialogBox := s.theme.StateSelectorDialog.Render(fullContent)
	return dialogBox
}

func (s *StateSelectorDialogCmp) SetSize(width, height int) {
	s.width = width
	s.height = height
}

func (s *StateSelectorDialogCmp) GetSize() (int, int) {
	return s.width, s.height
}

func (s *StateSelectorDialogCmp) GetModelState() consts.ModelState {
	return consts.ModelStateLoaded
}

func (s *StateSelectorDialogCmp) GetSelectedState() consts.SessionState {
	return s.states[s.selectedIndex]
}
