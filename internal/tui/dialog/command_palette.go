package dialog

import (
	"fmt"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Action constants for command palette commands.
const (
	ActionThemeSelector  = "theme_selector"
	ActionModelSelector  = "model_selector"
	ActionStateSelector  = "state_selector"
	ActionAddRecipe      = "add_recipe"
	ActionRecipeSelector = "recipe_selector"
)

// CommandItem represents a single command in the palette.
type CommandItem struct {
	Name     string
	Shortcut string
	Action   string
}

type CommandPaletteDialogCmp struct {
	allItems      []CommandItem
	filtered      []CommandItem
	selectedIndex int
	searchInput   textinput.Model
	width         int
	height        int
	theme         *themes.Theme
}

func NewCommandPaletteDialog(theme *themes.Theme) (*CommandPaletteDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	dialogConfig := cfg.CommandPaletteDialog
	km := cfg.Keymap

	items := []CommandItem{
		{Name: "Switch View", Shortcut: strings.Join(km.StateSelector, " / "), Action: ActionStateSelector},
		{Name: "Change Theme", Shortcut: strings.Join(km.ThemeSelector, " / "), Action: ActionThemeSelector},
		{Name: "Change Model", Shortcut: strings.Join(km.ModelSelector, " / "), Action: ActionModelSelector},
		{Name: "Add Recipe from URL", Shortcut: strings.Join(km.Add, " / "), Action: ActionAddRecipe},
		{Name: "Find Recipe", Shortcut: strings.Join(km.RecipeSelector, " / "), Action: ActionRecipeSelector},
	}

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 64
	if w := dialogConfig.Width - 8; w > 10 {
		ti.Width = w
	} else {
		ti.Width = 40
	}

	return &CommandPaletteDialogCmp{
		allItems:    items,
		filtered:    items,
		searchInput: ti,
		width:       dialogConfig.Width,
		height:      dialogConfig.Height,
		theme:       theme,
	}, nil
}

func (c *CommandPaletteDialogCmp) Init() tea.Cmd {
	return textinput.Blink
}

func (c *CommandPaletteDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmds = append(cmds, messages.SendCloseModalViewMsg())
			return c, tea.Batch(cmds...)

		case "enter":
			if len(c.filtered) > 0 && c.selectedIndex < len(c.filtered) {
				selected := c.filtered[c.selectedIndex]
				cmds = append(cmds,
					messages.SendCloseModalViewMsg(),
					messages.CmdHandler(messages.CommandPaletteActionMsg{Action: selected.Action}),
				)
			}
			return c, tea.Batch(cmds...)

		case "up", "ctrl+k":
			if c.selectedIndex > 0 {
				c.selectedIndex--
			}
			return c, nil

		case "down", "ctrl+j":
			if c.selectedIndex < len(c.filtered)-1 {
				c.selectedIndex++
			}
			return c, nil
		}
	}

	// Update search input
	prevValue := c.searchInput.Value()
	var cmd tea.Cmd
	c.searchInput, cmd = c.searchInput.Update(msg)
	cmds = append(cmds, cmd)

	// Re-filter when search text changes
	if c.searchInput.Value() != prevValue {
		c.applyFilter()
	}

	return c, tea.Batch(cmds...)
}

func (c *CommandPaletteDialogCmp) applyFilter() {
	query := strings.ToLower(c.searchInput.Value())
	if query == "" {
		c.filtered = c.allItems
	} else {
		c.filtered = nil
		for _, item := range c.allItems {
			if strings.Contains(strings.ToLower(item.Name), query) {
				c.filtered = append(c.filtered, item)
			}
		}
	}
	c.selectedIndex = 0
}

func (c *CommandPaletteDialogCmp) View() string {
	innerWidth := c.width - 6 // border (2) + padding (4)
	if innerWidth < 30 {
		innerWidth = 30
	}

	// Header: "Commands" left, "esc" right
	titleLeft := c.theme.CommandPaletteTitle.Render("Commands")
	escHint := c.theme.CommandPaletteHelp.Render("esc")
	titlePad := innerWidth - lipgloss.Width(titleLeft) - lipgloss.Width(escHint)
	if titlePad < 1 {
		titlePad = 1
	}
	header := titleLeft + strings.Repeat(" ", titlePad) + escHint

	// Search input
	searchLine := c.searchInput.View()

	// Separator
	sep := c.theme.CommandPaletteShortcut.Render(strings.Repeat("─", innerWidth))

	// Command items — one line each: name left, shortcut right
	var rows []string
	for i, item := range c.filtered {
		name := item.Name
		shortcut := item.Shortcut

		nameWidth := lipgloss.Width(name)
		shortcutWidth := lipgloss.Width(shortcut)
		gap := innerWidth - nameWidth - shortcutWidth
		if gap < 2 {
			gap = 2
		}

		line := name + strings.Repeat(" ", gap) + shortcut

		if i == c.selectedIndex {
			row := c.theme.CommandPaletteSelected.
				Width(innerWidth).
				Render(line)
			rows = append(rows, row)
		} else {
			nameRendered := c.theme.DialogUnselectedRow.
				Render(name)
			shortcutRendered := c.theme.CommandPaletteShortcut.Render(shortcut)
			row := nameRendered + strings.Repeat(" ", gap) + shortcutRendered
			rows = append(rows, row)
		}
	}

	if len(c.filtered) == 0 {
		rows = append(rows, c.theme.CommandPaletteHelp.Render("No matching commands"))
	}

	// Assemble
	parts := []string{header, "", searchLine, sep}
	parts = append(parts, rows...)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	rendered := c.theme.CommandPaletteDialog.
		Width(c.width).
		Render(content)

	return c.theme.CommandPaletteContainer.Render(rendered)
}

func (c *CommandPaletteDialogCmp) SetSize(width, height int) {
	c.width = width
	c.height = height
	if w := c.width - 8; w > 10 {
		c.searchInput.Width = w
	}
}

func (c *CommandPaletteDialogCmp) GetSize() (int, int) {
	return c.width, c.height
}

func (c *CommandPaletteDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}
