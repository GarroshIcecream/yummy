package main

import (
	"fmt"
	"recipe_me/db"
	"recipe_me/models"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the application state
type Model struct {
	cookbook       *db.CookBook
	recipeList     list.Model
	selectedRecipe *models.RecipeRaw
	err            error
	state          string // can be "list" or "detail"
	ready          bool
	viewport       viewport.Model
}

// item represents a recipe in the list
type item struct {
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// Initialize the model
func NewModel() (Model, error) {
	cookbook, err := db.NewCookBook()
	if err != nil {
		return Model{}, err
	}

	if err := cookbook.Open(); err != nil {
		return Model{}, err
	}

	// Get all recipes
	recipes, err := cookbook.AllRecipes()
	if err != nil {
		return Model{}, err
	}

	// Convert recipes to list items
	items := make([]list.Item, len(recipes))
	for i, recipe := range recipes {
		items[i] = item{
			title: recipe.RecipeName,
			desc:  fmt.Sprintf("ID: %d", recipe.ID),
		}
	}

	// Initialize list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "My Cookbook"

	return Model{
		cookbook:   cookbook,
		recipeList: l,
		state:      "list",
	}, nil
}

func (m Model) headerView() string {
	if m.selectedRecipe == nil {
		return ""
	}
	title := styles.titleStyle.Render(m.selectedRecipe.Name)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) footerView() string {
	if !m.ready {
		return ""
	}
	info := styles.infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.state == "list" {
				if i, ok := m.recipeList.SelectedItem().(item); ok {
					recipe, err := m.cookbook.GetFullRecipe(i.title)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.selectedRecipe = recipe
					m.state = "detail"

					// Initialize viewport immediately
					m.viewport = viewport.New(m.recipeList.Width(), m.recipeList.Height())
					m.viewport.SetContent(formatRecipeContent(m.selectedRecipe))
					m.ready = true
				}
			}
		case "esc":
			if m.state == "detail" {
				m.state = "list"
				m.selectedRecipe = nil
			}
		}

	case tea.WindowSizeMsg:
		if m.state == "list" {
			h, v := docStyle.GetFrameSize()
			m.recipeList.SetSize(msg.Width-h, msg.Height-v)
		} else if m.state == "detail" {
			headerHeight := lipgloss.Height(m.headerView())
			footerHeight := lipgloss.Height(m.footerView())
			verticalMarginHeight := headerHeight + footerHeight

			if !m.ready {
				m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
				m.viewport.YPosition = headerHeight
				m.viewport.SetContent(formatRecipeContent(m.selectedRecipe))
				m.ready = true
			} else {
				m.viewport.Width = msg.Width
				m.viewport.Height = msg.Height - verticalMarginHeight
			}
		}
	}

	// Handle viewport updates when in detail view
	if m.state == "detail" {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Handle list updates when in list view
	if m.state == "list" {
		m.recipeList, cmd = m.recipeList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// Update the View method
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	if m.state == "list" {
		return docStyle.Render(m.recipeList.View())
	}

	// Detail view
	if m.selectedRecipe != nil {
		if !m.ready {
			return "\n  Loading recipe..."
		}
		return fmt.Sprintf("%s\n%s\n%s",
			m.headerView(),
			m.viewport.View(),
			m.footerView())
	}

	return "Loading..."
}

func main() {
	m, err := NewModel()
	if err != nil {
		fmt.Printf("Error initializing app: %v\n", err)
		return
	}

	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // Add mouse support
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}

}
