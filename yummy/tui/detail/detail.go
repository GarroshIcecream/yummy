package detail

import (
	"strings"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	utils "github.com/GarroshIcecream/yummy/yummy/utils"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type DetailModel struct {
	// Configuration
	cookbook   *db.CookBook
	keyMap     config.KeyMap
	theme      *themes.Theme
	config     *config.DetailConfig
	modelState common.ModelState

	// Recipe
	Recipe          *utils.RecipeRaw
	renderedContent string
	content         string

	// UI
	width          int
	height         int
	scrollPosition int
	renderer       *glamour.TermRenderer
}

func New(cookbook *db.CookBook, keymaps config.KeyMap, theme *themes.Theme, config *config.DetailConfig) *DetailModel {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithWordWrap(config.ViewportWidth),
	)

	model := &DetailModel{
		cookbook:       cookbook,
		scrollPosition: 0,
		Recipe:         nil,
		width:          config.ViewportWidth,
		height:         config.ViewportHeight,
		keyMap:         keymaps,
		modelState:     common.ModelStateLoaded,
		theme:          theme,
		renderer:       renderer,
		config:         config,
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
		cmds = append(cmds, m.FetchRecipeData(msg.RecipeID))

	case messages.LoadRecipeMsg:
		m.scrollPosition = 0
		m.Recipe = msg.Recipe
		m.content = msg.Content
		m.renderedContent = msg.Markdown

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Edit):
			if m.Recipe != nil {
				cmdState := messages.SendSessionStateMsg(common.SessionStateEdit)
				cmdEdit := messages.SendEditRecipeMsg(m.Recipe.RecipeID)
				cmds = append(cmds, cmdState, cmdEdit)
			}
		case key.Matches(msg, m.keyMap.CursorUp):
			m.ScrollUp(m.config.ScrollSpeed)
		case key.Matches(msg, m.keyMap.CursorDown):
			m.ScrollDown(m.config.ScrollSpeed)
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.ScrollUp(m.config.ScrollSpeed)
			case tea.MouseButtonWheelDown:
				m.ScrollDown(m.config.ScrollSpeed)
			}
		}
	}

	return m, tea.Sequence(cmds...)
}

func (m *DetailModel) View() string {
	if m.Recipe == nil || m.renderedContent == "" {
		return m.renderEmptyView()
	}

	return m.renderContentView()
}

func (m *DetailModel) FetchRecipeData(recipe_id uint) tea.Cmd {
	return func() tea.Msg {
		recipe, err := m.cookbook.GetFullRecipe(recipe_id)
		if err != nil {
			return messages.LoadRecipeMsg{Recipe: nil, Markdown: "", Content: ""}
		}

		// Render markdown content immediately
		content := recipe.FormatRecipeMarkdown()
		markdown, err := m.renderer.Render(content)
		if err != nil {
			return messages.LoadRecipeMsg{Recipe: recipe, Markdown: "", Content: ""}
		}

		return messages.LoadRecipeMsg{Recipe: recipe, Markdown: markdown, Content: content}
	}
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
	if m.renderedContent == "" {
		return 0
	}
	return len(strings.Split(m.renderedContent, "\n"))
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
	if m.width > 0 && m.Recipe != nil {
		m.refreshContent()
	}
}

// refreshContent re-renders the markdown content with current settings
func (m *DetailModel) refreshContent() {
	if m.Recipe == nil {
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
	content, err := m.renderer.Render(m.content)
	if err == nil {
		m.renderedContent = content
		m.scrollPosition = 0
	}
}

// renderEmptyView renders the empty state
func (m *DetailModel) renderEmptyView() string {
	emptyMessage := m.theme.Loading.Render(m.config.NoRecipeSelectedMessage)
	return m.theme.DetailContent.Render(emptyMessage)
}

// renderContentView renders the main content with proper markdown rendering
func (m *DetailModel) renderContentView() string {
	lines := strings.Split(m.renderedContent, "\n")
	viewportHeight := m.GetViewportHeight()
	start := m.scrollPosition
	end := min(start+viewportHeight, len(lines))

	var visibleContent string
	if len(lines) == 0 {
		visibleContent = m.config.NoContentAvailableMessage
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

func (m *DetailModel) GetSessionState() common.SessionState {
	return common.SessionStateDetail
}

func (m *DetailModel) GetModelState() common.ModelState {
	return m.modelState
}

func (m *DetailModel) GetSize() (width, height int) {
	return m.width, m.height
}
