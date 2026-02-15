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

type modelEntry struct {
	Name     string
	Selected bool
}

type ModelSelectorDialogCmp struct {
	allItems      []modelEntry
	filtered      []modelEntry
	selectedIndex int
	searchInput   textinput.Model
	width         int
	height        int
	theme         *themes.Theme
}

func NewModelSelectorDialog(installedModels []string, currentModelName string, theme *themes.Theme) (*ModelSelectorDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	modelSelectorConfig := cfg.ModelSelectorDialog
	sort.Strings(installedModels)

	items := make([]modelEntry, len(installedModels))
	for i, name := range installedModels {
		items[i] = modelEntry{Name: name, Selected: name == currentModelName}
	}

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 64
	if w := modelSelectorConfig.Width - 8; w > 10 {
		ti.Width = w
	} else {
		ti.Width = 40
	}

	return &ModelSelectorDialogCmp{
		allItems:    items,
		filtered:    items,
		searchInput: ti,
		width:       modelSelectorConfig.Width,
		height:      modelSelectorConfig.Height,
		theme:       theme,
	}, nil
}

func (m *ModelSelectorDialogCmp) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ModelSelectorDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				cmds = append(cmds, messages.SendModelSelectedMsg(selected.Name))
				cmds = append(cmds, messages.SendCloseModalViewMsg())
			}
			return m, tea.Batch(cmds...)

		case "up", "ctrl+k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
			return m, nil

		case "down", "ctrl+j":
			if m.selectedIndex < len(m.filtered)-1 {
				m.selectedIndex++
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

func (m *ModelSelectorDialogCmp) applyFilter() {
	query := strings.ToLower(m.searchInput.Value())
	if query == "" {
		m.filtered = m.allItems
	} else {
		m.filtered = nil
		for _, item := range m.allItems {
			if strings.Contains(strings.ToLower(item.Name), query) {
				m.filtered = append(m.filtered, item)
			}
		}
	}
	m.selectedIndex = 0
}

func (m *ModelSelectorDialogCmp) View() string {
	innerWidth := m.width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}

	// Header
	titleLeft := m.theme.ModelSelectorTitle.Render("Select Model")
	escHint := m.theme.ModelSelectorHelp.Render("esc")
	pad := innerWidth - lipgloss.Width(titleLeft) - lipgloss.Width(escHint)
	if pad < 1 {
		pad = 1
	}
	header := titleLeft + strings.Repeat(" ", pad) + escHint

	// Search
	searchLine := m.searchInput.View()

	// Separator
	sep := m.theme.ModelSelectorHelp.Render(strings.Repeat("─", innerWidth))

	// Items
	var rows []string
	for i, item := range m.filtered {
		marker := "  "
		if item.Selected {
			marker = "* "
		}
		line := marker + item.Name

		if i == m.selectedIndex {
			row := m.theme.DialogSelectedRow.
				Width(innerWidth).
				Render(line)
			rows = append(rows, row)
		} else {
			rows = append(rows, m.theme.DialogUnselectedRow.
				Render(line))
		}
	}

	if len(m.filtered) == 0 {
		rows = append(rows, m.theme.ModelSelectorHelp.Render("No matching models"))
	}

	parts := []string{header, "", searchLine, sep}
	parts = append(parts, rows...)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	rendered := m.theme.ModelSelectorDialog.
		Width(m.width).
		Render(content)

	return m.theme.ModelSelectorContainer.Render(rendered)
}

func (m *ModelSelectorDialogCmp) SetSize(width, height int) {
	m.width = width
	m.height = height
	if w := m.width - 8; w > 10 {
		m.searchInput.Width = w
	}
}

func (m *ModelSelectorDialogCmp) GetSize() (int, int) {
	return m.width, m.height
}

func (m *ModelSelectorDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}
