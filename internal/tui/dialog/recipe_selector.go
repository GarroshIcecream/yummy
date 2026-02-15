package dialog

import (
	"fmt"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxVisibleRecipes = 8

type recipeEntry struct {
	ID   uint
	Name string
}

type RecipeSelectorDialogCmp struct {
	allItems      []recipeEntry
	filtered      []recipeEntry
	selectedIndex int
	scrollOffset  int
	searchInput   textinput.Model
	width         int
	height        int
	theme         *themes.Theme
}

func NewRecipeSelectorDialog(cookbook *db.CookBook, theme *themes.Theme) (*RecipeSelectorDialogCmp, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	recipeSelectorConfig := cfg.RecipeSelectorDialog

	recipes, err := cookbook.AllRecipes()
	if err != nil {
		return nil, fmt.Errorf("failed to load recipes: %w", err)
	}

	items := make([]recipeEntry, len(recipes))
	for i, r := range recipes {
		items[i] = recipeEntry{ID: r.RecipeID, Name: r.RecipeName}
	}

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 64
	if w := recipeSelectorConfig.Width - 8; w > 10 {
		ti.Width = w
	} else {
		ti.Width = 40
	}

	return &RecipeSelectorDialogCmp{
		allItems:    items,
		filtered:    items,
		searchInput: ti,
		width:       recipeSelectorConfig.Width,
		height:      recipeSelectorConfig.Height,
		theme:       theme,
	}, nil
}

func (r *RecipeSelectorDialogCmp) Init() tea.Cmd {
	return textinput.Blink
}

func (r *RecipeSelectorDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmds = append(cmds, messages.SendCloseModalViewMsg())
			return r, tea.Batch(cmds...)

		case "enter":
			if len(r.filtered) > 0 && r.selectedIndex < len(r.filtered) {
				selected := r.filtered[r.selectedIndex]
				return r, tea.Sequence(
					messages.SendCloseModalViewMsg(),
					messages.SendSessionStateMsg(common.SessionStateDetail),
					messages.SendRecipeSelectedMsg(selected.ID),
				)
			}
			return r, tea.Batch(cmds...)

		case "up", "ctrl+k":
			if r.selectedIndex > 0 {
				r.selectedIndex--
				r.ensureVisible()
			}
			return r, nil

		case "down", "ctrl+j":
			if r.selectedIndex < len(r.filtered)-1 {
				r.selectedIndex++
				r.ensureVisible()
			}
			return r, nil
		}
	}

	prevValue := r.searchInput.Value()
	var cmd tea.Cmd
	r.searchInput, cmd = r.searchInput.Update(msg)
	cmds = append(cmds, cmd)

	if r.searchInput.Value() != prevValue {
		r.applyFilter()
	}

	return r, tea.Batch(cmds...)
}

func (r *RecipeSelectorDialogCmp) ensureVisible() {
	if r.selectedIndex < r.scrollOffset {
		r.scrollOffset = r.selectedIndex
	}
	if r.selectedIndex >= r.scrollOffset+maxVisibleRecipes {
		r.scrollOffset = r.selectedIndex - maxVisibleRecipes + 1
	}
}

func (r *RecipeSelectorDialogCmp) applyFilter() {
	query := strings.ToLower(r.searchInput.Value())
	if query == "" {
		r.filtered = r.allItems
	} else {
		r.filtered = nil
		for _, item := range r.allItems {
			if strings.Contains(strings.ToLower(item.Name), query) {
				r.filtered = append(r.filtered, item)
			}
		}
	}
	r.selectedIndex = 0
	r.scrollOffset = 0
}

func (r *RecipeSelectorDialogCmp) View() string {
	innerWidth := r.width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}

	// Header
	titleLeft := r.theme.RecipeSelectorTitle.Render("Recipes")
	escHint := r.theme.RecipeSelectorHelp.Render("esc")
	pad := innerWidth - lipgloss.Width(titleLeft) - lipgloss.Width(escHint)
	if pad < 1 {
		pad = 1
	}
	header := titleLeft + strings.Repeat(" ", pad) + escHint

	// Search
	searchLine := r.searchInput.View()

	// Separator
	sep := r.theme.RecipeSelectorHelp.Render(strings.Repeat("─", innerWidth))

	// Visible window of items
	maxNameLen := innerWidth - 4
	if maxNameLen < 10 {
		maxNameLen = 10
	}

	end := r.scrollOffset + maxVisibleRecipes
	if end > len(r.filtered) {
		end = len(r.filtered)
	}
	visible := r.filtered[r.scrollOffset:end]

	var rows []string

	// Scroll up indicator
	if r.scrollOffset > 0 {
		rows = append(rows, r.theme.RecipeSelectorHelp.Render(
			strings.Repeat(" ", innerWidth/2-1)+"▲"))
	}

	for i, item := range visible {
		globalIdx := r.scrollOffset + i
		name := item.Name
		if lipgloss.Width(name) > maxNameLen {
			name = name[:maxNameLen-1] + "…"
		}
		line := "  " + name

		if globalIdx == r.selectedIndex {
			row := r.theme.RecipeSelectorSelected.
				Width(innerWidth).
				MaxWidth(innerWidth).
				Render(line)
			rows = append(rows, row)
		} else {
			rows = append(rows, r.theme.DialogUnselectedRow.
				Render(line))
		}
	}

	// Scroll down indicator
	if end < len(r.filtered) {
		rows = append(rows, r.theme.RecipeSelectorHelp.Render(
			strings.Repeat(" ", innerWidth/2-1)+"▼"))
	}

	if len(r.filtered) == 0 {
		rows = append(rows, r.theme.RecipeSelectorHelp.Render("No matching recipes"))
	}

	parts := []string{header, "", searchLine, sep}
	parts = append(parts, rows...)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	rendered := r.theme.RecipeSelectorDialog.
		Width(r.width).
		Render(content)

	return r.theme.RecipeSelectorContainer.Render(rendered)
}

func (r *RecipeSelectorDialogCmp) SetSize(width, height int) {
	r.width = width
	r.height = height
	if w := r.width - 8; w > 10 {
		r.searchInput.Width = w
	}
}

func (r *RecipeSelectorDialogCmp) GetSize() (int, int) {
	return r.width, r.height
}

func (r *RecipeSelectorDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}
