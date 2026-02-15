package detail

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	dialog "github.com/GarroshIcecream/yummy/internal/tui/dialog"
	utils "github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

type DetailModel struct {
	// Configuration
	cookbook   *db.CookBook
	theme      *themes.Theme
	keyMap     config.DetailKeyMap
	config     config.DetailConfig
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

func NewDetailModel(cookbook *db.CookBook, theme *themes.Theme) (*DetailModel, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetDetailKeyMap()
	detailConfig := cfg.Detail
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithWordWrap(detailConfig.ViewportWidth),
	)
	if err != nil {
		slog.Error("Failed to create renderer", "error", err)
		return nil, err
	}

	model := &DetailModel{
		cookbook:       cookbook,
		scrollPosition: 0,
		Recipe:         nil,
		width:          0,
		height:         detailConfig.ViewportHeight,
		keyMap:         keymaps,
		modelState:     common.ModelStateLoaded,
		theme:          theme,
		renderer:       renderer,
		config:         detailConfig,
	}

	return model, nil
}

func (m *DetailModel) Init() tea.Cmd {
	return nil
}

func (m *DetailModel) Update(msg tea.Msg) (common.TUIModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.RecipeSelectedMsg:
		cmds = append(cmds, m.FetchRecipeData(msg.RecipeID))

	case messages.LoadRecipeMsg:
		m.scrollPosition = 0
		m.Recipe = msg.Recipe
		m.content = msg.Content
		m.renderedContent = msg.Markdown
		m.modelState = common.ModelStateLoaded

	case messages.RatingSelectedMsg:
		if m.Recipe != nil && m.Recipe.RecipeID == msg.RecipeID {
			if err := m.cookbook.SetRating(m.Recipe.RecipeID, msg.Rating); err != nil {
				slog.Error("Failed to set rating", "error", err)
			} else {
				m.Recipe.Metadata.Rating = msg.Rating
				m.content = m.Recipe.FormatRecipeMarkdown()
				m.refreshContentKeepScroll()
			}
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Edit):
			if m.Recipe != nil {
				cmdState := messages.SendSessionStateMsg(common.SessionStateEdit)
				cmdEdit := messages.SendEditRecipeMsg(m.Recipe)
				cmds = append(cmds, cmdState, cmdEdit)
			}
		case key.Matches(msg, m.keyMap.SetRating):
			if m.Recipe != nil {
				ratingDialog := dialog.NewRatingDialog(
					m.Recipe.RecipeID,
					m.Recipe.Metadata.Rating,
					m.theme,
				)
				cmds = append(cmds, messages.SendOpenModalViewMsg(ratingDialog, common.ModalTypeRating))
			}
		case key.Matches(msg, m.keyMap.CookingMode):
			if m.Recipe != nil && len(m.Recipe.Metadata.Instructions) > 0 {
				cmds = append(cmds,
					messages.SendSessionStateMsg(common.SessionStateCooking),
					messages.SendEnterCookingModeMsg(m.Recipe),
				)
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
	lines := strings.Split(m.renderedContent, "\n")
	style := m.theme.DetailContent.Width(m.width)

	if m.Recipe == nil {
		emptyMessage := m.theme.Loading.Render(m.config.NoRecipeSelectedMessage)
		m.modelState = common.ModelStateLoaded
		return style.Height(m.height).Render(emptyMessage)
	}

	viewportHeight := m.GetViewportHeight()
	start := m.scrollPosition
	end := min(start+viewportHeight, len(lines))

	var visibleContent string
	if len(lines) == 0 || m.renderedContent == "" {
		visibleContent = m.config.NoContentAvailableMessage
	} else {
		visibleLines := make([]string, 0, end-start)
		for i := start; i < end && i < len(lines); i++ {
			visibleLines = append(visibleLines, lines[i])
		}
		visibleContent = strings.Join(visibleLines, "\n")
	}

	return style.Render(visibleContent)
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
	footerHeight := config.GetStatusLineConfig().Height
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
		glamour.WithWordWrap(m.width),
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

func (m *DetailModel) refreshContentKeepScroll() {
	if m.Recipe == nil {
		return
	}
	savedScroll := m.scrollPosition
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.width),
		glamour.WithEmoji(),
	)
	if err != nil {
		return
	}
	m.renderer = renderer
	content, err := m.renderer.Render(m.content)
	if err == nil {
		m.renderedContent = content
		m.scrollPosition = savedScroll
	}
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

func (m *DetailModel) GetCurrentTheme() *themes.Theme {
	return m.theme
}

func (m *DetailModel) SetTheme(theme *themes.Theme) {
	m.theme = theme
}
