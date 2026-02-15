package dialog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type themeEntry struct {
	Name     string
	Selected bool
}

type ThemeSelectorDialogCmp struct {
	allItems      []themeEntry
	filtered      []themeEntry
	selectedIndex int
	searchInput   textinput.Model
	width         int
	height        int
	theme         *themes.Theme
}

func NewThemeSelectorDialog(availableThemes []string, currentThemeName string, theme *themes.Theme) (*ThemeSelectorDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	themeSelectorConfig := cfg.ThemeSelectorDialog
	sort.Strings(availableThemes)

	items := make([]themeEntry, len(availableThemes))
	for i, name := range availableThemes {
		items[i] = themeEntry{Name: name, Selected: name == currentThemeName}
	}

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 64
	if w := themeSelectorConfig.Width - 8; w > 10 {
		ti.Width = w
	} else {
		ti.Width = 40
	}

	return &ThemeSelectorDialogCmp{
		allItems:    items,
		filtered:    items,
		searchInput: ti,
		width:       themeSelectorConfig.Width,
		height:      themeSelectorConfig.Height,
		theme:       theme,
	}, nil
}

func (t *ThemeSelectorDialogCmp) Init() tea.Cmd {
	return textinput.Blink
}

func (t *ThemeSelectorDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmds = append(cmds, messages.SendCloseModalViewMsg())
			return t, tea.Batch(cmds...)

		case "enter":
			if len(t.filtered) > 0 && t.selectedIndex < len(t.filtered) {
				selected := t.filtered[t.selectedIndex]
				cmds = append(cmds, messages.SendThemeSelectedMsg(selected.Name))
				cmds = append(cmds, messages.SendCloseModalViewMsg())
			}
			return t, tea.Batch(cmds...)

		case "up", "ctrl+k":
			if t.selectedIndex > 0 {
				t.selectedIndex--
			}
			return t, nil

		case "down", "ctrl+j":
			if t.selectedIndex < len(t.filtered)-1 {
				t.selectedIndex++
			}
			return t, nil
		}
	}

	prevValue := t.searchInput.Value()
	var cmd tea.Cmd
	t.searchInput, cmd = t.searchInput.Update(msg)
	cmds = append(cmds, cmd)

	if t.searchInput.Value() != prevValue {
		t.applyFilter()
	}

	return t, tea.Batch(cmds...)
}

func (t *ThemeSelectorDialogCmp) applyFilter() {
	query := strings.ToLower(t.searchInput.Value())
	if query == "" {
		t.filtered = t.allItems
	} else {
		t.filtered = nil
		for _, item := range t.allItems {
			if strings.Contains(strings.ToLower(item.Name), query) {
				t.filtered = append(t.filtered, item)
			}
		}
	}
	t.selectedIndex = 0
}

func (t *ThemeSelectorDialogCmp) View() string {
	innerWidth := t.width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}

	// Header
	titleLeft := t.theme.ThemeSelectorTitle.Render("Select Theme")
	escHint := t.theme.ThemeSelectorHelp.Render("esc")
	pad := innerWidth - lipgloss.Width(titleLeft) - lipgloss.Width(escHint)
	if pad < 1 {
		pad = 1
	}
	header := titleLeft + strings.Repeat(" ", pad) + escHint

	// Search
	searchLine := t.searchInput.View()

	// Separator
	sep := t.theme.ThemeSelectorHelp.Render(strings.Repeat("─", innerWidth))

	// Items
	var rows []string
	for i, item := range t.filtered {
		marker := "  "
		if item.Selected {
			marker = "* "
		}
		line := marker + item.Name

		if i == t.selectedIndex {
			row := t.theme.DialogSelectedRow.
				Width(innerWidth).
				Render(line)
			rows = append(rows, row)
		} else {
			rows = append(rows, t.theme.DialogUnselectedRow.
				Render(line))
		}
	}

	if len(t.filtered) == 0 {
		rows = append(rows, t.theme.ThemeSelectorHelp.Render("No matching themes"))
	}

	parts := []string{header, "", searchLine, sep}
	parts = append(parts, rows...)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	rendered := t.theme.ThemeSelectorDialog.
		Width(t.width).
		Render(content)

	return t.theme.ThemeSelectorContainer.Render(rendered)
}

func (t *ThemeSelectorDialogCmp) SetSize(width, height int) {
	t.width = width
	t.height = height
	if w := t.width - 8; w > 10 {
		t.searchInput.Width = w
	}
}

func (t *ThemeSelectorDialogCmp) GetSize() (int, int) {
	return t.width, t.height
}

func (t *ThemeSelectorDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}
