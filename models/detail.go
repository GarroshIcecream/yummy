package models

import (
	"fmt"
	"log"
	"strings"

	db "github.com/GarroshIcecream/yummy/db"
	keys "github.com/GarroshIcecream/yummy/keymaps"
	recipes "github.com/GarroshIcecream/yummy/recipe"
	"github.com/GarroshIcecream/yummy/styles"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type LoadRecipeMsg struct {
	recipe *recipes.RecipeRaw
	err    error
}

type InitMsg struct{}

type DetailModel struct {
	cookbook       *db.CookBook
	recipe_id      uint
	recipe_name    string
	current_recipe *recipes.RecipeRaw
	content        string
	markdown       string
	err            error
	scrollPosition int
	renderer       *glamour.TermRenderer
	windowHeight   int
	windowWidth    int
	contentHeight  int
	headerHeight   int
	footerHeight   int
	spinner        spinner.Model
}

func NewDetailModel(cookbook db.CookBook, recipe_id uint) *DetailModel {
	log.Printf("Creating new DetailModel for recipe ID: %d\n", recipe_id)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0),
	)

	model := &DetailModel{
		cookbook:       &cookbook,
		recipe_id:      recipe_id,
		renderer:       renderer,
		scrollPosition: 0,
		windowHeight:   0,
		windowWidth:    0,
		contentHeight:  0,
		headerHeight:   4,
		footerHeight:   2,
		spinner:        s,
	}

	return model
}

func (m *DetailModel) Init() tea.Cmd {
	log.Printf("Initializing DetailModel...")
	return func() tea.Msg {
		return InitMsg{}
	}
}

func (m *DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case InitMsg:
		if m.content == "" && m.err == nil {
			log.Printf("Received InitMsg message, starting recipe load...")
			cmd = tea.Batch(
				m.spinner.Tick,
				m.loadRecipe(),
			)
			return m, cmd
		}

	case LoadRecipeMsg:
		log.Printf("Received LoadRecipeMsg")
		if msg.err != nil {
			log.Printf("Error in LoadRecipeMsg: %v\n", msg.err)
			m.err = msg.err
			return m, nil
		}

		log.Printf("Formatting recipe content...")
		m.markdown = recipes.FormatRecipeContent(msg.recipe)
		m.current_recipe = msg.recipe
		m.recipe_name = msg.recipe.Name

		log.Printf("Rendering markdown...")
		rendered, err := m.renderer.Render(m.markdown)
		if err != nil {
			log.Printf("Error rendering markdown: %v\n", err)
			m.err = err
			return m, nil
		}

		log.Printf("Updating model with recipe content...")
		m.content = rendered
		m.contentHeight = len(strings.Split(m.content, "\n"))
		log.Printf("Content height: %d lines\n", m.contentHeight)
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		if m.content == "" && m.err == nil {
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Quit):
			log.Printf("Quit command received")
			return m, tea.Quit
		case key.Matches(msg, keys.Keys.Back):
			log.Printf("Back command received")
			return NewListModel(*m.cookbook, nil), nil
		case key.Matches(msg, keys.Keys.Up):
			m.scrollUp(1)
		case key.Matches(msg, keys.Keys.Down):
			m.scrollDown(1)
		case key.Matches(msg, keys.Keys.Edit):
			log.Printf("Edit command received for recipe ID: %d", m.recipe_id)
			return NewEditModel(*m.cookbook, m.current_recipe, m.recipe_id), nil
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
		log.Printf("Window size changed: %dx%d\n", msg.Width, msg.Height)
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width

		if m.windowWidth > 0 {
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(m.windowWidth-4),
			)
			if err == nil {
				m.renderer = renderer
				if m.markdown != "" {
					log.Printf("Re-rendering content with new window width")
					rendered, err := m.renderer.Render(m.markdown)
					if err == nil {
						m.content = rendered
						m.contentHeight = len(strings.Split(m.content, "\n"))
						log.Printf("Content re-rendered, new height: %d lines\n", m.contentHeight)
					}
				}
			}
		}
	}

	return m, nil
}

func (m *DetailModel) loadRecipe() tea.Cmd {
	return func() tea.Msg {
		log.Printf("Loading recipe with ID: %d\n", m.recipe_id)
		recipe, err := m.cookbook.GetFullRecipe(m.recipe_id)
		if err != nil {
			log.Printf("Error loading recipe: %v\n", err)
			return LoadRecipeMsg{recipe: nil, err: err}
		}
		log.Printf("Successfully loaded recipe: %s\n", recipe.Name)
		return LoadRecipeMsg{recipe: recipe, err: nil}
	}
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

	if m.content == "" {
		return fmt.Sprintf("\n  %s Loading recipe...", m.spinner.View())
	}

	lines := strings.Split(m.content, "\n")
	visibleLines := make([]string, 0)

	viewportHeight := m.getViewportHeight()
	start := m.scrollPosition
	end := min(start+viewportHeight, m.contentHeight)

	for i := start; i < end; i++ {
		if i < len(lines) {
			visibleLines = append(visibleLines, lines[i])
		}
	}

	visibleContent := strings.Join(visibleLines, "\n")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.headerView(),
		visibleContent,
		m.footerView(),
	)
}

func (m *DetailModel) headerView() string {
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
