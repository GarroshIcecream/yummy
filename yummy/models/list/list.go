package list

import (
	"fmt"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	keys "github.com/GarroshIcecream/yummy/yummy/keymaps"
	recipe "github.com/GarroshIcecream/yummy/yummy/recipe"
	styles "github.com/GarroshIcecream/yummy/yummy/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ListModel struct {
	cookbook       *db.CookBook
	err            error
	RecipeList     list.Model
	SelectedRecipe *uint
}

func New(cookbook *db.CookBook, recipe_id *uint) *ListModel {
	recipes, err := cookbook.AllRecipes()

	var items []list.Item
	for _, recipe := range recipes {
		items = append(items, recipe)
	}

	d := list.NewDefaultDelegate()
	d = styles.ApplyDelegateStyles(d)

	l := list.New(items, d, 80, 40)
	l = styles.ApplyListStyles(l)
	l.Title = "📚 My Cookbook"
	l.SetStatusBarItemName("recipe", "recipes")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Keys.Add, keys.Keys.Delete}
	}

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keys.Keys.Add, keys.Keys.Delete}
	}

	return &ListModel{
		cookbook:       cookbook,
		err:            err,
		RecipeList:     l,
		SelectedRecipe: recipe_id,
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
		case key.Matches(msg, keys.Keys.Delete):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.RecipeList.SelectedItem().(recipe.RecipeWithDescription); ok {
					if err := m.cookbook.DeleteRecipe(i.RecipeID); err != nil {
						m.err = err
						return m, nil
					}

					m.RefreshRecipeList()
					return m, nil
				}
			}
		case key.Matches(msg, keys.Keys.Enter):
			if m.RecipeList.FilterState() != list.Filtering {
				if i, ok := m.RecipeList.SelectedItem().(recipe.RecipeWithDescription); ok {
					m.SelectedRecipe = &i.RecipeID
				}
			}
		}

	case tea.WindowSizeMsg:
		h, v := styles.DocStyle.GetFrameSize()
		m.RecipeList.SetSize(msg.Width-h, msg.Height-v)
	}

	m.RecipeList, cmd = m.RecipeList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	return styles.DocStyle.Render(m.RecipeList.View())
}

func (m *ListModel) RefreshRecipeList() tea.Cmd {
	recipes, err := m.cookbook.AllRecipes()
	if err != nil {
		m.err = err
		return nil
	}

	var items []list.Item
	for _, recipe := range recipes {
		items = append(items, recipe)
	}

	cmd := m.RecipeList.SetItems(items)

	return cmd
}
