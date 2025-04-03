package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"

	db "github.com/GarroshIcecream/yummy/db"
	keys "github.com/GarroshIcecream/yummy/keymaps"
	recipes "github.com/GarroshIcecream/yummy/recipe"
	styles "github.com/GarroshIcecream/yummy/styles"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
)

type DetailModel struct {
	cookbook       *db.CookBook
	current_recipe *recipes.RecipeRaw
	err            error
	ready          bool
	content        string
	scrollPosition int
	renderer       *glamour.TermRenderer
	windowHeight   int
	windowWidth    int
	contentHeight  int
	headerHeight   int
	footerHeight   int
}

func NewDetailModel(cookbook db.CookBook, recipe_id uint) *DetailModel {
	recipe, err := cookbook.GetFullRecipe(recipe_id)

	// Create a custom renderer with better styling
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0),
	)

	return &DetailModel{
		cookbook:       &cookbook,
		current_recipe: recipe,
		err:            err,
		ready:          false,
		renderer:       renderer,
		scrollPosition: 0,
		windowHeight:   0,
		windowWidth:    0,
		contentHeight:  0,
		headerHeight:   4,
		footerHeight:   2,
	}
}

func (m *DetailModel) Init() tea.Cmd {
	return nil
}

func (m *DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Keys.Back):
			return NewListModel(*m.cookbook, nil), nil
		case key.Matches(msg, keys.Keys.Up):
			m.scrollUp(1)
		case key.Matches(msg, keys.Keys.Down):
			m.scrollDown(1)
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.scrollUp(3)
			case tea.MouseButtonWheelDown:
				m.scrollDown(3)
			}
		}

	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width

		// Create new renderer with updated width
		if m.windowWidth > 0 {
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(m.windowWidth-4), // Leave some margin
			)
			if err == nil {
				m.renderer = renderer
			}
		}

		if !m.ready && m.current_recipe != nil {
			// Initial render of content
			markdown := recipes.FormatRecipeContent(m.current_recipe)
			rendered, err := m.renderer.Render(markdown)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.content = rendered
			m.contentHeight = len(strings.Split(m.content, "\n"))
			m.ready = true
		} else if m.ready && m.current_recipe != nil {
			// Re-render content with new width
			markdown := recipes.FormatRecipeContent(m.current_recipe)
			rendered, err := m.renderer.Render(markdown)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.content = rendered
			m.contentHeight = len(strings.Split(m.content, "\n"))
			// Reset scroll position after re-render
			m.scrollPosition = 0
		}
	}

	return m, nil
}

func (m *DetailModel) scrollUp(amount int) {
	m.scrollPosition = max(0, m.scrollPosition-amount)
}

func (m *DetailModel) scrollDown(amount int) {
	maxScroll := max(0, m.contentHeight-m.getViewportHeight())
	m.scrollPosition = min(maxScroll, m.scrollPosition+amount)
}

func (m *DetailModel) getViewportHeight() int {
	return m.windowHeight - m.headerHeight - m.footerHeight
}

func (m *DetailModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	if !m.ready {
		return "\n  Initializing..."
	}

	// Split content into lines and handle scrolling
	lines := strings.Split(m.content, "\n")
	visibleLines := make([]string, 0)

	// Calculate visible range
	viewportHeight := m.getViewportHeight()
	start := m.scrollPosition
	end := min(start+viewportHeight, m.contentHeight)

	// Get visible lines
	for i := start; i < end; i++ {
		if i < len(lines) {
			visibleLines = append(visibleLines, lines[i])
		}
	}

	// Join visible lines
	visibleContent := strings.Join(visibleLines, "\n")

	// Create the full view with header and footer
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.headerView(),
		visibleContent,
		m.footerView(),
	)
}

func (m *DetailModel) headerView() string {
	// Add cooking emoji and center the title
	header := styles.TitleStyle.Render("🍳 " + m.current_recipe.Name)

	return lipgloss.JoinHorizontal(lipgloss.Left, header)
}

func (m *DetailModel) footerView() string {
	scrollPercent := float64(m.scrollPosition) / float64(max(1, m.contentHeight-m.getViewportHeight()))
	info := styles.InfoStyle.Render(fmt.Sprintf("%3.f%%", scrollPercent*100))
	nav := styles.InfoStyle.Render("ESC: Back | q: Quit")
	line := strings.Repeat("─", max(0, m.windowWidth-lipgloss.Width(info)-lipgloss.Width(nav)-2))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, nav, " ", info)
}
