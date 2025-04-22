package models

import (
	"fmt"

	db "github.com/GarroshIcecream/yummy/db"
	keys "github.com/GarroshIcecream/yummy/keymaps"
	"github.com/GarroshIcecream/yummy/recipe"
	styles "github.com/GarroshIcecream/yummy/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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
			if m.recipeList.FilterState() != list.Filtering {
				return NewInputModel(*m.cookbook, nil), nil
			}
		case key.Matches(msg, keys.Keys.Delete):
			if m.recipeList.FilterState() != list.Filtering {
				if i, ok := m.recipeList.SelectedItem().(recipe.RecipeWithDescription); ok {
					if err := m.cookbook.DeleteRecipe(i.RecipeID); err != nil {
						m.err = err
						return m, nil
					}

					recipes, err := m.cookbook.AllRecipes()
					if err != nil {
						m.err = err
						return m, nil
					}

					var items []list.Item
					for _, recipe := range recipes {
						items = append(items, recipe)
					}
					m.recipeList.SetItems(items)
					return m, nil
				}
			}
		case key.Matches(msg, keys.Keys.Enter):
			if m.recipeList.FilterState() != list.Filtering {
				if i, ok := m.recipeList.SelectedItem().(recipe.RecipeWithDescription); ok {
					m.selectedRecipe = &i.RecipeID
					return NewDetailModel(*m.cookbook, i.RecipeID), nil
				}
			}
		case key.Matches(msg, keys.Keys.Back):
			if m.recipeList.FilterState() != list.Filtering {
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
		return fmt.Sprintf("Error: %v", m.err)
	}

	return styles.DocStyle.Render(m.recipeList.View())
}
