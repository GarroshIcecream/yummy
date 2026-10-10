package dialog

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/GarroshIcecream/yummy/internal/config"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	utils "github.com/GarroshIcecream/yummy/internal/utils"
)

type themeEntry struct {
	Name        string
	Description string
	Selected    bool
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

func NewThemeSelectorDialog(availableThemes []themes.ThemeInfo, currentThemeName string, theme *themes.Theme) (*ThemeSelectorDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	themeSelectorConfig := cfg.ThemeSelectorDialog
	sort.Slice(availableThemes, func(i, j int) bool {
		return availableThemes[i].Name < availableThemes[j].Name
	})

	items := make([]themeEntry, len(availableThemes))
	for i, availableTheme := range availableThemes {
		items[i] = themeEntry{
			Name:        availableTheme.Name,
			Description: availableTheme.Description,
			Selected:    availableTheme.Name == currentThemeName,
		}
	}

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.SetStyles(theme.DialogSearchStyles)
	ti.Focus()
	ti.CharLimit = 64
	if w := themeSelectorConfig.Width - 8; w > 10 {
		ti.SetWidth(w)
	} else {
		ti.SetWidth(40)
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
	return t.searchInput.Focus()
}

func (t *ThemeSelectorDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
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
			return t, t.previewCmd()

		case "down", "ctrl+j":
			if t.selectedIndex < len(t.filtered)-1 {
				t.selectedIndex++
			}
			return t, t.previewCmd()
		}
	}

	prevValue := t.searchInput.Value()
	var cmd tea.Cmd
	t.searchInput, cmd = t.searchInput.Update(msg)
	cmds = append(cmds, cmd)

	if t.searchInput.Value() != prevValue {
		t.applyFilter()
		cmds = append(cmds, t.previewCmd())
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

func (t *ThemeSelectorDialogCmp) View() tea.View {
	innerWidth := max(t.width-6, 20)

	// Header
	titleLeft := t.theme.ThemeSelectorTitle.Render("Select Theme")
	escHint := t.theme.ThemeSelectorHelp.Render("esc")
	pad := max(innerWidth-lipgloss.Width(titleLeft)-lipgloss.Width(escHint), 1)
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
		titleLine := marker + item.Name
		description := item.Description
		if description == "" {
			description = "No description available"
		}
		descriptionLine := "   " + description

		if i == t.selectedIndex {
			titleRow := t.theme.DialogSelectedRow.
				Width(innerWidth).
				Render(titleLine)
			descRow := t.theme.SessionSelectorSelectedDesc.
				Width(innerWidth).
				Render(descriptionLine)
			rows = append(rows, titleRow+"\n"+descRow)
		} else {
			titleRow := t.theme.DialogUnselectedRow.
				Width(innerWidth).
				Render(titleLine)
			descRow := t.theme.SessionSelectorUnselectedDesc.
				Width(innerWidth).
				Render(descriptionLine)
			rows = append(rows, titleRow+"\n"+descRow)
		}
	}

	if len(t.filtered) == 0 {
		rows = append(rows, t.theme.ThemeSelectorHelp.Render("No matching themes"))
	}

	parts := []string{header, "", searchLine, sep}
	parts = append(parts, rows...)
	content := utils.FillBackground(lipgloss.JoinVertical(lipgloss.Left, parts...), t.theme.SurfaceBg, innerWidth)

	rendered := t.theme.ThemeSelectorDialog.
		Width(t.width).
		Render(content)

	return tea.NewView(t.theme.ThemeSelectorContainer.Render(rendered))
}

func (t *ThemeSelectorDialogCmp) SetSize(width, height int) {
	t.width = clampModalWidth(width, t.width, 24)
	t.height = clampModalHeight(height, t.height, 10)
	if w := t.width - 8; w > 10 {
		t.searchInput.SetWidth(w)
	}
}

func (t *ThemeSelectorDialogCmp) SetTheme(theme *themes.Theme) {
	t.theme = theme
	t.searchInput.SetStyles(theme.DialogSearchStyles)
}

func (t *ThemeSelectorDialogCmp) previewCmd() tea.Cmd {
	if len(t.filtered) == 0 || t.selectedIndex >= len(t.filtered) {
		return nil
	}
	return messages.SendThemePreviewMsg(t.filtered[t.selectedIndex].Name)
}

func (t *ThemeSelectorDialogCmp) GetSize() (int, int) {
	return t.width, t.height
}

func (t *ThemeSelectorDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}
