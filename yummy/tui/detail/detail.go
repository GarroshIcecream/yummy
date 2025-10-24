package detail

import (
	"fmt"
	"strings"

	"github.com/GarroshIcecream/yummy/yummy/config"
	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	recipes "github.com/GarroshIcecream/yummy/yummy/recipe"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type DetailModel struct {
	// Database
	cookbook       *db.CookBook
	CurrentRecipe  *recipes.RecipeRaw
	content        string
	err            error
	scrollPosition int
	renderer       *glamour.TermRenderer

	width      int
	height     int
	modelState consts.ModelState

	// UI
	keyMap config.KeyMap
	theme  *themes.Theme
}

func New(cookbook *db.CookBook, keymaps config.KeyMap, theme *themes.Theme) *DetailModel {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithWordWrap(consts.DefaultViewportWidth),
	)

	model := &DetailModel{
		cookbook:       cookbook,
		scrollPosition: 0,
		width:          consts.DefaultViewportWidth,
		height:         consts.DefaultViewportHeight,
		keyMap:         keymaps,
		modelState:     consts.ModelStateLoaded,
		theme:          theme,
		renderer:       renderer,
	}

	return model
}

func (m *DetailModel) Init() tea.Cmd {
	return nil
}

func (m *DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.RecipeSelectedMsg:
		cmd_load := messages.SendLoadRecipeMsg(m.FetchRecipe(msg.RecipeID))
		cmds = append(cmds, cmd_load)

	case messages.LoadRecipeMsg:
		m.scrollPosition = 0
		m.CurrentRecipe = msg.Recipe
		m.content = msg.Content
		m.err = msg.Err

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Edit):
			if m.CurrentRecipe != nil {
				cmd_state := messages.SendSessionStateMsg(consts.SessionStateEdit)
				cmd_edit := messages.SendEditRecipeMsg(m.CurrentRecipe.ID)
				cmds = append(cmds, cmd_state, cmd_edit)
			}
		case key.Matches(msg, m.keyMap.CursorUp):
			m.ScrollUp(consts.DefaultScrollSpeed)
		case key.Matches(msg, m.keyMap.CursorDown):
			m.ScrollDown(consts.DefaultScrollSpeed)
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.ScrollUp(consts.DefaultScrollSpeed)
			case tea.MouseButtonWheelDown:
				m.ScrollDown(consts.DefaultScrollSpeed)
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

func (m *DetailModel) FetchRecipe(recipe_id uint) messages.LoadRecipeMsg {
	recipe, err := m.cookbook.GetFullRecipe(recipe_id)
	if err != nil {
		return messages.LoadRecipeMsg{Recipe: nil, Content: "", Err: err}
	}

	// Render markdown content immediately
	markdown := recipes.FormatRecipeContent(recipe)
	content, renderErr := m.renderer.Render(markdown)
	if renderErr != nil {
		content = markdown
	}

	return messages.LoadRecipeMsg{Recipe: recipe, Content: content, Err: nil}
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

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(max(40, m.width-4)),
		glamour.WithEmoji(),
	)
	if err != nil {
		return
	}

	m.renderer = renderer
	markdown := recipes.FormatRecipeContent(m.CurrentRecipe)
	content, err := m.renderer.Render(markdown)
	if err == nil {
		m.content = content
		m.scrollPosition = 0
	}
}

func (m *DetailModel) GetSessionState() consts.SessionState {
	return consts.SessionStateDetail
}

func (m *DetailModel) GetModelState() consts.ModelState {
	return m.modelState
}

func (m *DetailModel) GetSize() (width, height int) {
	return m.width, m.height
}

// renderErrorView renders the error state
func (m *DetailModel) renderErrorView() string {
	return m.theme.Error.Render(fmt.Sprintf("❌ Error: %v", m.err))
}

// renderEmptyView renders the empty state
func (m *DetailModel) renderEmptyView() string {
	emptyMessage := m.theme.Loading.Render("📖 No recipe selected - choose one from the list")
	return m.theme.DetailContent.Render(emptyMessage)
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
	return m.theme.DetailContent.Render(visibleContent)
}
