package models

import (
	"fmt"
	"log"

	db "github.com/GarroshIcecream/yummy/db"
	keys "github.com/GarroshIcecream/yummy/keymaps"
	styles "github.com/GarroshIcecream/yummy/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ListModel struct {
	cookbook       *db.CookBook
	err            error
	recipeList     list.Model
	selectedRecipe *uint
}

func NewListModel(cookbook db.CookBook, recipe_id *uint) *ListModel {
	recipes, err := cookbook.AllRecipes()

	var items []list.Item
	for _, recipe := range recipes {
		items = append(items, recipe)
	}

	d := list.NewDefaultDelegate()

	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 2)

	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Padding(0, 2)

	selectedColor := lipgloss.Color("#FF6B6B")
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(selectedColor).
		BorderLeftForeground(selectedColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		Padding(0, 2)

	d.Styles.SelectedDesc = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FFB6B6"))

	l := list.New(items, d, 80, 40)
	l.Title = "📚 My Cookbook"
	l.SetStatusBarItemName("recipe", "recipes")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Keys.Add, keys.Keys.Delete}
	}

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Keys.Add, keys.Keys.Delete}
	}

	// Create a gradient-like effect for the title bar
	l.Styles.TitleBar = lipgloss.NewStyle().
		Background(lipgloss.Color("#FF6B6B")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("#FF8E8E"))

	// Make the title pop with a subtle glow effect
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		MarginLeft(2)

	// Style pagination with dots
	l.Styles.PaginationStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		PaddingLeft(2)

	// Style the help text to be more subtle
	l.Styles.HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Padding(1, 0, 0, 2)

	// Make the filter prompt stand out
	l.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)

	// Style the filter cursor
	l.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF6B6B"))

	// Style the "no items" message
	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Italic(true)

	// Style the pagination dots
	l.Styles.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		SetString("●")

	l.Styles.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		SetString("○")

	l.Styles.DividerDot = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6")).
		SetString(" • ")

	// Style the status bar
	l.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A0A0A0")).
		Padding(0, 0, 1, 2)

	// Style the status bar when filtering
	l.Styles.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)

	// Style the filter count
	l.Styles.StatusBarFilterCount = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFB6B6"))

	return &ListModel{
		cookbook:       &cookbook,
		err:            err,
		recipeList:     l,
		selectedRecipe: recipe_id,
	}
}

func (m *ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Keys.Add):
			return NewEditModel(*m.cookbook, nil, 0), nil
		case key.Matches(msg, keys.Keys.Delete):
			if i, ok := m.recipeList.SelectedItem().(db.RecipeWithDescription); ok {
				if err := m.cookbook.DeleteRecipe(i.RecipeID); err != nil {
					log.Printf("Error deleting recipe: %v", err)
					return m, nil
				}
				// Refresh the list
				return NewListModel(*m.cookbook, nil), nil
			}
		case key.Matches(msg, keys.Keys.Enter):
			if i, ok := m.recipeList.SelectedItem().(db.RecipeWithDescription); ok {
				m.selectedRecipe = &i.RecipeID
				return NewDetailModel(*m.cookbook, i.RecipeID), nil
			}
		case key.Matches(msg, keys.Keys.Back):
			if !m.recipeList.FilteringEnabled() {
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		h, v := styles.DocStyle.GetFrameSize()
		m.recipeList.SetSize(msg.Width-h, msg.Height-v)
	}

	m.recipeList, cmd = m.recipeList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ListModel) View() string {
	if m.err != nil {
		log.Fatalf("Encountered unknown error while viewing ListModel:")
		return fmt.Sprintf("Error: %v", m.err)
	}

	return styles.DocStyle.Render(m.recipeList.View())
}
