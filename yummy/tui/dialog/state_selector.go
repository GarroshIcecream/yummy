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
	wWidth  int
	wHeight int

	selectedIndex int
	states        []consts.SessionState
	stateNames    []string
	keymap        config.KeyMap
	height        int
	width         int
	emojis        []string
	theme         *themes.Theme
}

// NewStateSelectorDialog creates a new state selection dialog.
func NewStateSelectorDialog(theme *themes.Theme) *StateSelectorDialogCmp {
	states := []consts.SessionState{
		consts.SessionStateMainMenu,
		consts.SessionStateList,
		consts.SessionStateDetail,
		consts.SessionStateEdit,
		consts.SessionStateChat,
	}

	stateNames := []string{
		"Main Menu",
		"Recipe List",
		"Recipe Detail",
		"Edit Recipe",
		"Chat Assistant",
	}

	emojis := []string{
		"🏠",
		"📝",
		"🔍",
		"📝",
		"💬",
	}

	return &StateSelectorDialogCmp{
		selectedIndex: 0,
		states:        states,
		height:        consts.DefaultViewportHeight,
		width:         consts.DefaultViewportWidth,
		stateNames:    stateNames,
		emojis:        emojis,
		keymap:        config.DefaultKeyMap(),
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
		s.wWidth = msg.Width
		s.wHeight = msg.Height
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
		case key.Matches(msg, s.keymap.Back, s.keymap.Quit):
			cmds = append(cmds, messages.SendCloseDialogMsg())
		}
	}

	return s, tea.Batch(cmds...)
}

// View renders the state selector dialog with a list of states.
func (s *StateSelectorDialogCmp) View() string {
	title := "🎯 Select State"

	// Create the list of states
	var stateItems []string
	for i, stateName := range s.stateNames {
		style := lipgloss.NewStyle()
		if i == s.selectedIndex {
			style = style.
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#0078d4")).
				Bold(true).
				Padding(0, 1)
		} else {
			style = style.
				Foreground(lipgloss.Color("#cccccc")).
				Padding(0, 1)
		}

		// Add number prefix and emoji
		numberPrefix := fmt.Sprintf("%d. ", i+1)
		indicator := "  "
		if i == s.selectedIndex {
			indicator = s.emojis[i] + " "
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
	centeredDialogStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Width(s.width).
		Height(s.height)

	return centeredDialogStyle.Render(dialogBox)
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

func (s *StateSelectorDialogCmp) GetSelectedStateName() string {
	return s.stateNames[s.selectedIndex]
}

func (s *StateSelectorDialogCmp) GetSelectedStateEmoji() string {
	return s.emojis[s.selectedIndex]
}
