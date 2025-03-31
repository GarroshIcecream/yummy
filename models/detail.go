package models

import (
	"fmt"
	"log"
	"strings"

	db "github.com/GarroshIcecream/yummy/db"
	keys "github.com/GarroshIcecream/yummy/keymaps"
	recipes "github.com/GarroshIcecream/yummy/recipe"
	styles "github.com/GarroshIcecream/yummy/styles"
	"github.com/charmbracelet/bubbles/key"
	viewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

type DetailModel struct {
	cookbook       *db.CookBook
	current_recipe *recipes.RecipeRaw
	err            error
	ready          bool
	viewport       viewport.Model
}

func NewDetailModel(cookbook db.CookBook, recipe_id uint) *DetailModel {
	recipe, err := cookbook.GetFullRecipe(recipe_id)

	return &DetailModel{
		cookbook:       &cookbook,
		current_recipe: recipe,
		err:            err,
		ready:          false,
	}
}

func (m *DetailModel) Init() tea.Cmd {
	return nil
}

func (m *DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Quit):
			log.Println("Application quitting...")
			return m, tea.Quit
		case key.Matches(msg, keys.Keys.Back):
			log.Println("Going back to ListModel...")
			// Create new list model and send window size update
			newModel := NewListModel(*m.cookbook, nil)
			return newModel, tea.Batch(
				cmd,
				func() tea.Msg {
					return tea.WindowSizeMsg{
						Width:  m.viewport.Width,
						Height: m.viewport.Height + lipgloss.Height(m.headerView()) + lipgloss.Height(m.footerView()),
					}
				},
			)
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			if m.current_recipe != nil {
				m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
				m.viewport.YPosition = headerHeight
				m.viewport.SetContent(recipes.FormatRecipeContent(m.current_recipe))
				m.ready = true
			} else {
				log.Println("Current recipe is nil, cannot initialize viewport.")
			}
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *DetailModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s",
		m.headerView(),
		m.viewport.View(),
		m.footerView())
}

func (m *DetailModel) headerView() string {
	title := styles.TitleStyle.Render(m.current_recipe.Name)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m *DetailModel) footerView() string {
	info := styles.InfoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
