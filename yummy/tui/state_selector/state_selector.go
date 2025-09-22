package state_selector

import (
	config "github.com/GarroshIcecream/yummy/yummy/config"
	"github.com/GarroshIcecream/yummy/yummy/tui/styles"
	"github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StateSelectorDialogCmp struct {
	wWidth  int
	wHeight int

	selectedIndex int
	states        []ui.SessionState
	stateNames    []string
	keymap        config.KeyMap
	height        int
	width         int
	emojis        []string
}

// NewStateSelectorDialog creates a new state selection dialog.
func New() *StateSelectorDialogCmp {
	states := []ui.SessionState{
		ui.SessionStateMainMenu,
		ui.SessionStateList,
		ui.SessionStateDetail,
		ui.SessionStateEdit,
		ui.SessionStateChat,
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
		height:        ui.DefaultViewportHeight,
		width:         ui.DefaultViewportWidth,
		stateNames:    stateNames,	
		emojis:        emojis,
		keymap:        config.DefaultKeyMap(),
	}
}

func (s *StateSelectorDialogCmp) Init() tea.Cmd {
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
		case key.Matches(msg, s.keymap.Up):
			if s.selectedIndex > 0 {
				s.selectedIndex--
			} else {
				s.selectedIndex = len(s.states) - 1
			}
		case key.Matches(msg, s.keymap.Down):
			if s.selectedIndex < len(s.states)-1 {
				s.selectedIndex++
			} else {
				s.selectedIndex = 0
			}
		
		case key.Matches(msg, s.keymap.Enter):
			selectedState := s.states[s.selectedIndex]
			cmds = append(cmds, tea.Batch(
				ui.SendSessionStateMsg(selectedState),
				ui.SendCloseDialogMsg(),
			))
		case key.Matches(msg, s.keymap.Back, s.keymap.Quit):
			cmds = append(cmds, ui.SendCloseDialogMsg())
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
		
		indicator := "  "
		if i == s.selectedIndex {
			indicator = s.emojis[i] + " "
		}
		
		stateItems = append(stateItems, style.Render(indicator+stateName))
	}
	
	// Create the dialog box with better styling
	content := lipgloss.JoinVertical(lipgloss.Left, stateItems...)
	fullContent := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.StateSelectorTitleStyle.Render(title),
		content,
	)

	dialogBox := styles.StateSelectorDialogStyle.Render(fullContent)
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

func (s *StateSelectorDialogCmp) GetModelState() ui.ModelState {
	return ui.ModelStateLoaded
}

func (s *StateSelectorDialogCmp) GetSelectedState() ui.SessionState {
	return s.states[s.selectedIndex]
}

func (s *StateSelectorDialogCmp) GetSelectedStateName() string {
	return s.stateNames[s.selectedIndex]
}

func (s *StateSelectorDialogCmp) GetSelectedStateEmoji() string {
	return s.emojis[s.selectedIndex]
}