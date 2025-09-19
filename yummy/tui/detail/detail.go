package detail

import (
	"fmt"
	"strings"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	recipes "github.com/GarroshIcecream/yummy/yummy/recipe"
	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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
	keyMap         config.KeyMap
}

func New(cookbook *db.CookBook, keymaps config.KeyMap) *DetailModel {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithWordWrap(ui.DefaultViewportWidth),
	)

	model := &DetailModel{
		cookbook:       cookbook,
		renderer:       renderer,
		scrollPosition: 0,
		width:          ui.DefaultViewportWidth,
		height:         ui.DefaultViewportHeight,
		keyMap:         keymaps,
	}

	return model
}

func (m *DetailModel) Init() tea.Cmd {
	return nil
}

func (m *DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ui.RecipeSelectedMsg:
		cmd_load := ui.SendLoadRecipeMsg(m.FetchRecipe(msg.RecipeID))
		cmds = append(cmds, cmd_load)

	case ui.LoadRecipeMsg:
		m.scrollPosition = 0
		m.CurrentRecipe = msg.Recipe
		m.content = msg.Content
		m.err = msg.Err

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Edit):
			if m.CurrentRecipe != nil {
				cmd_state := ui.SendSessionStateMsg(ui.SessionStateEdit)
				cmd_edit := ui.SendEditRecipeMsg(m.CurrentRecipe.ID)
				cmds = append(cmds, cmd_state, cmd_edit)
			}
		case key.Matches(msg, m.keyMap.Up):
			m.ScrollUp(ui.DefaultScrollSpeed)
		case key.Matches(msg, m.keyMap.Down):
			m.ScrollDown(ui.DefaultScrollSpeed)
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.ScrollUp(ui.DefaultScrollSpeed)
			case tea.MouseButtonWheelDown:
				m.ScrollDown(ui.DefaultScrollSpeed)
			}
		}
	}

	return m, tea.Sequence(cmds...)
}

func (m *DetailModel) View() string {
	if m.err != nil {
		return m.renderErrorView()
	}

	if m.CurrentRecipe == nil {
		return m.renderEmptyView()
	}

	return m.renderContentView()
}

func (m *DetailModel) FetchRecipe(recipe_id uint) ui.LoadRecipeMsg {
	recipe, err := m.cookbook.GetFullRecipe(recipe_id)
	if err != nil {
		return ui.LoadRecipeMsg{Recipe: nil, Content: "", Err: err}
	}

	// Render markdown content immediately
	markdown := recipes.FormatRecipeContent(recipe)
	content, renderErr := m.renderer.Render(markdown)
	if renderErr != nil {
		content = markdown
	}

	return ui.LoadRecipeMsg{Recipe: recipe, Content: content, Err: nil}
}

func (m *DetailModel) ScrollUp(amount int) {
	if amount <= 0 {
		return
	}
	m.scrollPosition = max(0, m.scrollPosition-amount)
}

func (m *DetailModel) ScrollDown(amount int) {
	if amount <= 0 {
		return
	}
	contentHeight := m.GetContentHeight()
	viewportHeight := m.GetViewportHeight()

	if contentHeight <= viewportHeight {
		return
	}

	maxScroll := contentHeight - viewportHeight
	m.scrollPosition = min(maxScroll, m.scrollPosition+amount)
}

// getContentHeight returns the height of the content in lines
func (m *DetailModel) GetContentHeight() int {
	if m.content == "" {
		return 0
	}
	return len(strings.Split(m.content, "\n"))
}

func (m *DetailModel) GetViewportHeight() int {
	footerHeight := 3 // Footer + scroll bar
	return max(1, m.height-footerHeight)
}

func (m *DetailModel) GetScrollPosition() int {
	return m.scrollPosition
}

func (m *DetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Re-render content when size changes to ensure proper word wrapping
	if m.width > 0 && m.CurrentRecipe != nil {
		m.refreshContent()
	}
}

// refreshContent re-renders the markdown content with current settings
func (m *DetailModel) refreshContent() {
	if m.CurrentRecipe == nil {
		return
	}

	// Create a new renderer with current width
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(max(40, m.width-4)), // Ensure minimum width
		glamour.WithEmoji(),
	)
	if err != nil {
		// Fallback to existing renderer if creation fails
		return
	}

	m.renderer = renderer
	markdown := recipes.FormatRecipeContent(m.CurrentRecipe)
	content, err := m.renderer.Render(markdown)
	if err == nil {
		m.content = content
		// Reset scroll position when content changes
		m.scrollPosition = 0
	}
}

func (m *DetailModel) GetSize() (width, height int) {
	return m.width, m.height
}

// renderErrorView renders the error state
func (m *DetailModel) renderErrorView() string {
	return styles.ErrorStyle.Render(fmt.Sprintf("❌ Error: %v", m.err))
}

// renderEmptyView renders the empty state
func (m *DetailModel) renderEmptyView() string {
	emptyMessage := styles.LoadingStyle.Render("📖 No recipe selected - choose one from the list")
	return styles.DetailContentStyle.Render(emptyMessage)
}

// renderContentView renders the main content with proper markdown rendering
func (m *DetailModel) renderContentView() string {
	// Use the rendered content if available, otherwise render markdown on the fly
	var content string
	if m.content != "" {
		content = m.content
	} else if m.CurrentRecipe != nil {
		// Fallback: render markdown if content is empty
		markdown := recipes.FormatRecipeContent(m.CurrentRecipe)
		if rendered, err := m.renderer.Render(markdown); err == nil {
			content = rendered
		} else {
			content = markdown // Fallback to raw markdown
		}
	}

	// Handle scrolling for the rendered content
	lines := strings.Split(content, "\n")
	viewportHeight := m.GetViewportHeight()
	start := m.scrollPosition
	end := min(start+viewportHeight, len(lines))

	var visibleContent string
	if len(lines) == 0 {
		visibleContent = "📝 No content available"
	} else {
		visibleLines := make([]string, 0, end-start)
		for i := start; i < end && i < len(lines); i++ {
			visibleLines = append(visibleLines, lines[i])
		}
		visibleContent = strings.Join(visibleLines, "\n")
	}

	// Apply content styling
	return styles.DetailContentStyle.Render(visibleContent)
}
