package dialog

import (
	"fmt"
	"sort"

	"github.com/GarroshIcecream/yummy/yummy/config"
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ModelItem struct {
	Name     string
	Selected bool
}

var _ list.Item = ModelItem{}

func (m ModelItem) Title() string {
	if m.Selected {
		return "🔥 " + m.Name
	}
	return m.Name
}

func (m ModelItem) Description() string {
	return ""
}

func (m ModelItem) FilterValue() string {
	return m.Name
}

type ModelSelectorDialogCmp struct {
	keyMap config.ModelSelectorKeyMap
	list   list.Model
	width  int
	height int
	theme  *themes.Theme
}

func NewModelSelectorDialog(installedModels []string, currentModelName string, theme *themes.Theme) (*ModelSelectorDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetModelSelectorKeyMap()
	modelSelectorConfig := cfg.ModelSelectorDialog
	sort.Strings(installedModels)

	items := make([]list.Item, len(installedModels))
	for i, modelName := range installedModels {
		items[i] = ModelItem{
			Name:     modelName,
			Selected: modelName == currentModelName,
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles = theme.ModelSelectorDelegateStyles

	l := list.New(items, delegate, 80, 24)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = theme.ModelSelectorTitle
	l.Styles.PaginationStyle = theme.ModelSelectorPagination
	l.Styles.HelpStyle = theme.ModelSelectorHelp

	return &ModelSelectorDialogCmp{
		keyMap: keymaps,
		list:   l,
		width:  modelSelectorConfig.Width,
		height: modelSelectorConfig.Height,
		theme:  theme,
	}, nil
}

func (m *ModelSelectorDialogCmp) Init() tea.Cmd {
	return nil
}

func (m *ModelSelectorDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.ModelSelector):
			cmds = append(cmds, messages.SendCloseModalViewMsg())

		case key.Matches(msg, m.keyMap.Enter):
			if selectedItem, ok := m.list.SelectedItem().(ModelItem); ok {
				cmds = append(cmds, messages.SendModelSelectedMsg(selectedItem.Name))
				cmds = append(cmds, messages.SendCloseModalViewMsg())
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ModelSelectorDialogCmp) View() string {
	title := m.theme.ModelSelectorTitle.Render("🤖 Select Model")
	listContent := m.list.View()
	helpText := m.theme.ModelSelectorHelp.Render("↑/↓ Navigate • Enter Select • Esc Cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		listContent,
		helpText,
	)

	content = m.theme.ModelSelectorDialog.
		Width(m.width).
		Height(m.height).
		Render(content)

	return m.theme.ModelSelectorContainer.Render(content)
}

func (m *ModelSelectorDialogCmp) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetWidth(width - 4)
	m.list.SetHeight(height - 6)
}

func (m *ModelSelectorDialogCmp) GetSize() (int, int) {
	return m.width, m.height
}

func (m *ModelSelectorDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}
