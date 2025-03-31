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
)

type ListModel struct {
	cookbook       *db.CookBook
	err            error
	recipeList     list.Model
	selectedRecipe *uint
}

func NewListModel(cookbook db.CookBook, recipe_id *uint) *ListModel {
	// Get all recipes
	recipes, err := cookbook.AllRecipes()

	// Initialize list
	var items []list.Item
	for _, recipe := range recipes {
		items = append(items, recipe)
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "My Cookbook"
	l.Styles = styles.PunkyStyle

	// Set initial size to ensure visibility
	h, v := styles.DocStyle.GetFrameSize()
	l.SetSize(80-h, 20-v)

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
			return NewInputModel(*m.cookbook, nil), nil
		case key.Matches(msg, keys.Keys.Enter):
			if i, ok := m.recipeList.SelectedItem().(db.RecipeWithDescription); ok {
				m.selectedRecipe = &i.RecipeID
				return NewDetailModel(*m.cookbook, i.RecipeID), nil
			}
		case key.Matches(msg, keys.Keys.Back):
			return m, tea.Quit
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
