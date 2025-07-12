package detail

import (
	"fmt"
	"log"
	"strings"
	"time"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	keys "github.com/GarroshIcecream/yummy/yummy/keymaps"
	recipes "github.com/GarroshIcecream/yummy/yummy/recipe"
	styles "github.com/GarroshIcecream/yummy/yummy/styles"
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
	recipe_id      *uint
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
	isLoading      bool
}

// Creates a new DetailModel for a given recipe ID
func New(cookbook *db.CookBook, recipe_id *uint) *DetailModel {
	start := time.Now()
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	// Create a new renderer for the markdown
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithWordWrap(0),
	)

	model := &DetailModel{
		cookbook:       cookbook,
		recipe_id:      recipe_id,
		renderer:       renderer,
		scrollPosition: 0,
		windowHeight:   0,
		windowWidth:    0,
		contentHeight:  0,
		headerHeight:   4,
		footerHeight:   2,
		spinner:        s,
		isLoading:      true,
	}

	log.Printf("DetailModel created in %v", time.Since(start))
	return model
}

func (m *DetailModel) Init() tea.Cmd {
	return func() tea.Msg {
		return InitMsg{}
	}
}

func (m *DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case InitMsg:
		log.Println("Init Meessage received")
		cmds = append(cmds, m.spinner.Tick)
		cmds = append(cmds, func() tea.Msg {
			return m.loadRecipe()
		})
	case LoadRecipeMsg:
		log.Println("LoadRecipeMsg received")
		m.isLoading = false

	case spinner.TickMsg:
		if m.isLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		switch {
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
				if m.markdown != "" {
					rendered, err := m.renderer.Render(m.markdown)
					if err == nil {
						m.content = rendered
						m.contentHeight = len(strings.Split(m.content, "\n"))
					}
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *DetailModel) loadRecipe() LoadRecipeMsg {
	log.Println("Starting loadRecipe...")
	start := time.Now()

	recipe, err := m.cookbook.GetFullRecipe(*m.recipe_id)
	if err != nil {
		log.Printf("Error getting recipe: %v", err)
		return LoadRecipeMsg{recipe: nil, err: err}
	}

	m.markdown = recipes.FormatRecipeContent(recipe)
	m.current_recipe = recipe

	rendered, err := m.renderer.Render(m.markdown)
	if err != nil {
		return LoadRecipeMsg{recipe: nil, err: err}
	}

	m.content = rendered

	m.content = m.markdown
	m.contentHeight = len(strings.Split(m.content, "\n"))
	log.Printf("Total loadRecipe completed in %v", time.Since(start))

	return LoadRecipeMsg{recipe: recipe, err: nil}
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

	if m.isLoading {
		return fmt.Sprintf("\n  %s Loading and formatting recipe...", m.spinner.View())
	}

	content := m.content
	if m.current_recipe != nil && m.current_recipe.URL != "" {
		urlStyle := "\033]8;;" + m.current_recipe.URL + "\033\\"
		urlText := m.current_recipe.URL + "\033]8;;\033\\"
		content = strings.Replace(content, m.current_recipe.URL, urlStyle+urlText, -1)
	}

	lines := strings.Split(content, "\n")
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
	var recipeName string
	if m.current_recipe != nil {
		recipeName = m.current_recipe.Name
	} else {
		recipeName = "Loading..."
	}
	header := styles.TitleStyle.Render("🍳 " + recipeName)

	return lipgloss.JoinHorizontal(lipgloss.Left, header)
}

func (m *DetailModel) footerView() string {
	scrollPercent := float64(m.scrollPosition) / float64(max(1, m.contentHeight-m.getViewportHeight()))
	info := styles.InfoStyle.Render(fmt.Sprintf("%3.f%%", scrollPercent*100))
	nav := styles.InfoStyle.Render("ESC: Back | q: Quit")
	line := strings.Repeat("─", max(0, m.windowWidth-lipgloss.Width(info)-lipgloss.Width(nav)-2))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, nav, " ", info)
}
