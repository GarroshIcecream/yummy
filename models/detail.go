package models

import (
	"fmt"
	"strings"
	"time"

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

type DetailModel struct {
	cookbook       *db.CookBook
	recipe_id      uint
	recipe_name    string
	current_recipe *recipes.RecipeRaw
	content        string
	err            error
	ready          bool
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
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0),
	)

	return &DetailModel{
		cookbook:       &cookbook,
		recipe_id:      recipe_id,
		ready:          false,
		renderer:       renderer,
		scrollPosition: 0,
		windowHeight:   0,
		windowWidth:    0,
		contentHeight:  0,
		headerHeight:   4,
		footerHeight:   2,
		spinner:        s,
	}
}

func (m *DetailModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			recipe, err := m.cookbook.GetFullRecipe(m.recipe_id)
			if err != nil {
				m.err = err
				m.ready = true
				return nil
			}

			markdown := recipes.FormatRecipeContent(recipe)
			rendered, err := m.renderer.Render(markdown)
			if err != nil {
				m.err = err
				m.ready = true
				return nil
			}

			m.recipe_name = recipe.Name
			m.content = rendered
			m.contentHeight = len(strings.Split(m.content, "\n"))
			m.ready = true
			return nil
		},
	)
}

func (m *DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case spinner.TickMsg:
		time.Sleep(5 * time.Millisecond)
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, m.spinner.Tick)

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

		if m.windowWidth > 0 {
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(m.windowWidth-4),
			)
			if err == nil {
				m.renderer = renderer
			}
		}
	}

	if !m.ready {
		time.Sleep(5 * time.Millisecond)
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, m.spinner.Tick)
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
