package detail

import (
	"fmt"
	"strings"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	recipes "github.com/GarroshIcecream/yummy/yummy/recipe"
	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type DetailModel struct {
	cookbook       *db.CookBook
	CurrentRecipe  *recipes.RecipeRaw
	content        string
	err            error
	scrollPosition int
	renderer       *glamour.TermRenderer
	width          int
	height         int
}

func New(cookbook *db.CookBook) *DetailModel {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithWordWrap(0),
	)

	model := &DetailModel{
		cookbook:       cookbook,
		renderer:       renderer,
		scrollPosition: 0,
		width:          ui.DefaultViewportWidth,
		height:         ui.DefaultViewportHeight,
	}

	return model
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
	case ui.RecipeSelectedMsg:
		cmds = append(cmds, m.SendLoadRecipeMsg(msg.RecipeID))

	case ui.LoadRecipeMsg:
		m.scrollPosition = 0
		m.CurrentRecipe = msg.Recipe
		m.content = msg.Content
		m.err = msg.Err

	// case tea.KeyMsg:
	// 	switch {
	// 	case key.Matches(msg, keys.Keys.Edit):
	// 		cmds = append(cmds, ui.SendSessionStateMsg(ui.SessionStateEdit))
	// 		cmds = append(cmds, ui.SendEditRecipeMsg(m.CurrentRecipe.ID))
	// 	}
	// case tea.MouseMsg:
	// 	if msg.Action == tea.MouseActionPress {
	// 		switch msg.Button {
	// 		case tea.MouseButtonWheelUp:
	// 			m.scrollUp(3)
	// 		case tea.MouseButtonWheelDown:
	// 			m.scrollDown(3)
	// 		}
	// 	}
	}

	cmds = append(cmds, cmd)
	return m, tea.Sequence(cmds...)
}

func (m *DetailModel) View() string {
	if m.err != nil {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.headerView(),
			styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)),
			m.footerView(),
		)
	}

	if m.CurrentRecipe == nil {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.headerView(),
			styles.InfoStyle.Render("No recipe selected"),
			m.footerView(),
		)
	}

	lines := strings.Split(m.content, "\n")
	viewportHeight := m.getViewportHeight()
	start := m.scrollPosition
	end := min(start+viewportHeight, len(lines))

	var visibleContent string
	if len(lines) == 0 {
		visibleContent = "No content available"
	} else {
		visibleLines := make([]string, 0)
		for i := start; i < end; i++ {
			if i < len(lines) {
				visibleLines = append(visibleLines, lines[i])
			}
		}
		visibleContent = strings.Join(visibleLines, "\n")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.headerView(),
		visibleContent,
		m.footerView(),
	)
}

func (m *DetailModel) SendLoadRecipeMsg(recipe_id uint) tea.Cmd {
	return func() tea.Msg {
		return m.FetchRecipe(recipe_id)
	}
}

func (m *DetailModel) FetchRecipe(recipe_id uint) ui.LoadRecipeMsg {
	recipe, err := m.cookbook.GetFullRecipe(recipe_id)
	if err != nil {
		return ui.LoadRecipeMsg{Recipe: nil, Content: "", Err: err}
	}

	markdown := recipes.FormatRecipeContent(recipe)

	return ui.LoadRecipeMsg{Recipe: recipe, Content: markdown, Err: nil}
}

func (m *DetailModel) ScrollUp(amount int) {
	m.scrollPosition = max(0, m.scrollPosition-amount)
}

func (m *DetailModel) ScrollDown(amount int) {
	contentHeight := len(strings.Split(m.content, "\n"))
	maxScroll := max(0, contentHeight-m.getViewportHeight())
	m.scrollPosition = min(maxScroll, m.scrollPosition+amount)
}

func (m *DetailModel) getViewportHeight() int {
	return max(0, m.height-4)
}

func (m *DetailModel) headerView() string {
	var recipeName string
	if m.CurrentRecipe != nil {
		recipeName = m.CurrentRecipe.Name
	} else {
		recipeName = "Loading..."
	}
	header := styles.TitleStyle.Render("🍳 " + recipeName)
	return header
}

func (m *DetailModel) footerView() string {
	if m.CurrentRecipe == nil {
		return styles.InfoStyle.Render("ESC: Back | q: Quit")
	}

	contentHeight := len(strings.Split(m.content, "\n"))
	viewportHeight := m.getViewportHeight()

	var scrollPercent float64
	if contentHeight > viewportHeight {
		scrollPercent = float64(m.scrollPosition) / float64(contentHeight-viewportHeight)
	} else {
		scrollPercent = 0
	}

	info := styles.InfoStyle.Render(fmt.Sprintf("%.0f%%", scrollPercent*100))
	nav := styles.InfoStyle.Render("ESC: Back | q: Quit")

	availableWidth := max(0, m.width-lipgloss.Width(info)-lipgloss.Width(nav)-2)
	line := strings.Repeat("─", availableWidth)

	return lipgloss.JoinHorizontal(lipgloss.Center, line, nav, " ", info)
}

func (m *DetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	
	if m.width > 0 && m.CurrentRecipe != nil {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.width-4),
			glamour.WithEmoji(),
		)
		if err == nil {
			m.renderer = renderer
			markdown := recipes.FormatRecipeContent(m.CurrentRecipe)
			content, err := m.renderer.Render(markdown)
			if err == nil {
				m.content = content
			}
		}
	}
}

func (m *DetailModel) GetSize() (width, height int) {
	return m.width, m.height
}
