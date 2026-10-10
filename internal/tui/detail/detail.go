package detail

import (
	"fmt"
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"github.com/GarroshIcecream/yummy/internal/config"
	db "github.com/GarroshIcecream/yummy/internal/db"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	dialog "github.com/GarroshIcecream/yummy/internal/tui/dialog"
	"github.com/GarroshIcecream/yummy/internal/tui/markdown"
	utils "github.com/GarroshIcecream/yummy/internal/utils"
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
}

func NewDetailModel(cookbook *db.CookBook, theme *themes.Theme) (*DetailModel, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetDetailKeyMap()
	detailConfig := cfg.Detail

	model := &DetailModel{
		cookbook:       cookbook,
		scrollPosition: 0,
		Recipe:         nil,
		width:          detailConfig.ViewportWidth,
		height:         detailConfig.ViewportHeight,
		keyMap:         keymaps,
		modelState:     common.ModelStateLoaded,
		theme:          theme,
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
		m.renderedContent = ""
		m.render()
		m.modelState = common.ModelStateLoaded

	case messages.RatingSelectedMsg:
		if m.Recipe != nil && m.Recipe.RecipeID == msg.RecipeID {
			if err := m.cookbook.SetRating(m.Recipe.RecipeID, msg.Rating); err != nil {
				slog.Error("Failed to set rating", "error", err)
			} else {
				m.Recipe.Metadata.Rating = msg.Rating
				m.content = m.Recipe.FormatRecipeMarkdown()
				m.render()
			}
		}

	case tea.KeyPressMsg:
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
		case key.Matches(msg, m.keyMap.PageUp):
			m.ScrollUp(m.GetViewportHeight())
		case key.Matches(msg, m.keyMap.PageDown):
			m.ScrollDown(m.GetViewportHeight())
		case key.Matches(msg, m.keyMap.GoToStart):
			m.scrollPosition = 0
		case key.Matches(msg, m.keyMap.GoToEnd):
			m.scrollPosition = m.maxScroll()
		}

	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.ScrollUp(m.config.ScrollSpeed)
		case tea.MouseWheelDown:
			m.ScrollDown(m.config.ScrollSpeed)
		}
	}

	return m, tea.Sequence(cmds...)
}

func (m *DetailModel) View() string {
	lines := strings.Split(m.renderedContent, "\n")
	style := m.theme.DetailContent.Width(m.width).Height(m.height)

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

		// The markdown is rendered in Update (render), against the current
		// width and theme, not here in the command goroutine.
		return messages.LoadRecipeMsg{Recipe: recipe, Content: recipe.FormatRecipeMarkdown()}
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
	m.scrollPosition = min(m.maxScroll(), m.scrollPosition+amount)
}

func (m *DetailModel) maxScroll() int {
	return max(0, m.GetContentHeight()-m.GetViewportHeight())
}

// getContentHeight returns the height of the content in lines
func (m *DetailModel) GetContentHeight() int {
	if m.renderedContent == "" {
		return 0
	}
	return len(strings.Split(m.renderedContent, "\n"))
}

// GetViewportHeight returns the number of content lines visible at once. The
// manager has already reserved the status line out of m.height.
func (m *DetailModel) GetViewportHeight() int {
	return max(1, m.height)
}

func (m *DetailModel) GetScrollPosition() int {
	return m.scrollPosition
}

func (m *DetailModel) SetSize(width, height int) {
	resized := width != m.width
	m.width = width
	m.height = height

	// Re-render on width changes so word wrapping matches the new width.
	if resized {
		m.render()
	}
	m.scrollPosition = min(m.scrollPosition, m.maxScroll())
}

// render renders the recipe markdown for the current width and theme,
// keeping the scroll position (clamped to the new content height).
func (m *DetailModel) render() {
	if m.Recipe == nil || m.content == "" || m.width <= 0 {
		return
	}

	// Wrap inside DetailContent's horizontal padding so no rendered line is
	// wider than the content area (lipgloss would hard-wrap it otherwise).
	wrap := max(m.width-m.theme.DetailContent.GetHorizontalFrameSize(), 20)
	renderer, err := markdown.NewThemedRenderer(m.theme, wrap, glamour.WithEmoji())
	if err != nil {
		slog.Error("Failed to create markdown renderer", "error", err)
		return
	}
	content, err := renderer.Render(m.content)
	if err != nil {
		slog.Error("Failed to render recipe markdown", "error", err)
		return
	}
	// Normalise the trailing whitespace to exactly one blank line, so when
	// scrolled to the end the last line doesn't sit flush on the status bar.
	m.renderedContent = strings.TrimRight(content, "\n") + "\n"
	m.scrollPosition = min(m.scrollPosition, m.maxScroll())
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
	m.render()
}
