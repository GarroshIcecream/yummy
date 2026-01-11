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

type ThemeItem struct {
	Name     string
	Selected bool
}

var _ list.Item = ThemeItem{}

func (t ThemeItem) Title() string {
	if t.Selected {
		return "🔥 " + t.Name
	}
	return t.Name
}

func (t ThemeItem) Description() string {
	return ""
}

func (t ThemeItem) FilterValue() string {
	return t.Name
}

type ThemeSelectorDialogCmp struct {
	keyMap config.ThemeSelectorKeyMap
	list   list.Model
	width  int
	height int
	theme  *themes.Theme
}

func NewThemeSelectorDialog(availableThemes []string, currentThemeName string, theme *themes.Theme) (*ThemeSelectorDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetThemeSelectorKeyMap()
	themeSelectorConfig := cfg.ThemeSelectorDialog
	sort.Strings(availableThemes)

	items := make([]list.Item, len(availableThemes))
	for i, themeName := range availableThemes {
		items[i] = ThemeItem{
			Name:     themeName,
			Selected: themeName == currentThemeName,
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles = theme.ThemeSelectorDelegateStyles

	l := list.New(items, delegate, 80, 24)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = theme.ThemeSelectorTitle
	l.Styles.PaginationStyle = theme.ThemeSelectorPagination
	l.Styles.HelpStyle = theme.ThemeSelectorHelp

	return &ThemeSelectorDialogCmp{
		keyMap: keymaps,
		list:   l,
		width:  themeSelectorConfig.Width,
		height: themeSelectorConfig.Height,
		theme:  theme,
	}, nil
}

func (t *ThemeSelectorDialogCmp) Init() tea.Cmd {
	return nil
}

func (t *ThemeSelectorDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.keyMap.ThemeSelector):
			cmds = append(cmds, messages.SendCloseModalViewMsg())

		case key.Matches(msg, t.keyMap.Enter):
			if selectedItem, ok := t.list.SelectedItem().(ThemeItem); ok {
				cmds = append(cmds, messages.SendThemeSelectedMsg(selectedItem.Name))
				cmds = append(cmds, messages.SendCloseModalViewMsg())
			}
		}
	}

	t.list, cmd = t.list.Update(msg)
	cmds = append(cmds, cmd)

	return t, tea.Batch(cmds...)
}

func (t *ThemeSelectorDialogCmp) View() string {
	title := t.theme.ThemeSelectorTitle.Render("🎨 Select Theme")
	listContent := t.list.View()
	helpText := t.theme.ThemeSelectorHelp.Render("↑/↓ Navigate • Enter Select • Esc Cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		listContent,
		helpText,
	)

	content = t.theme.ThemeSelectorDialog.
		Width(t.width).
		Height(t.height).
		Render(content)

	return t.theme.ThemeSelectorContainer.Render(content)
}

func (t *ThemeSelectorDialogCmp) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.list.SetWidth(width - 4)
	t.list.SetHeight(height - 6)
}

func (t *ThemeSelectorDialogCmp) GetSize() (int, int) {
	return t.width, t.height
}

func (t *ThemeSelectorDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}
