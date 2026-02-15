package detail

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/GarroshIcecream/yummy/internal/config"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	utils "github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// chatResponseMsg is a local message returned when the LLM responds.
type chatResponseMsg struct {
	response string
	err      error
}

// timerTickMsg is sent every second while the cooking timer is running.
type timerTickMsg time.Time

// chatEntry represents a single message in the cooking chat history.
type chatEntry struct {
	role    string // "user" or "assistant"
	content string
}

type CookingModel struct {
	Recipe          *utils.RecipeRaw
	CurrentStep     int
	TotalSteps      int
	showIngredients bool
	theme           *themes.Theme
	keyMap          config.CookingKeyMap
	modelState      common.ModelState
	width           int
	height          int

	// Chat panel
	showChat         bool
	chatTextarea     textarea.Model
	chatViewport     viewport.Model
	chatSpinner      spinner.Model
	chatHistory      []chatEntry
	chatWaiting      bool
	llm              llms.Model
	ctx              context.Context
	markdownRenderer *glamour.TermRenderer

	// Cooking timer
	timerRunning   bool
	timerDone      bool
	timerTotal     time.Duration
	timerRemaining time.Duration
	timerDoneAt    time.Time // used to animate the "done" celebration
}

func NewCookingModel(theme *themes.Theme) (*CookingModel, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetCookingKeyMap()

	// Textarea for chat input
	ta := textarea.New()
	ta.Placeholder = "Ask about this step..."
	ta.CharLimit = 300
	ta.SetWidth(30)
	ta.SetHeight(2)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = theme.TextareaCursorLine
	ta.FocusedStyle.Base = theme.TextareaBase
	ta.BlurredStyle.Base = theme.TextareaBase
	ta.FocusedStyle.Placeholder = theme.TextareaPlaceholder
	ta.FocusedStyle.Text = theme.TextareaText
	ta.FocusedStyle.Prompt = theme.TextareaPrompt
	ta.Prompt = "› "
	ta.FocusedStyle.EndOfBuffer = theme.TextareaEndOfBuffer
	ta.Blur()

	// Viewport for chat messages
	vp := viewport.New(30, 10)

	// Spinner for loading state
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.Spinner

	// Markdown renderer for chat responses
	mdRenderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(26),
	)
	if err != nil {
		slog.Error("Failed to create markdown renderer for cooking chat", "error", err)
	}

	return &CookingModel{
		theme:            theme,
		keyMap:           keymaps,
		modelState:       common.ModelStateLoaded,
		chatTextarea:     ta,
		chatViewport:     vp,
		chatSpinner:      s,
		chatHistory:      []chatEntry{},
		ctx:              context.Background(),
		markdownRenderer: mdRenderer,
	}, nil
}

func (m *CookingModel) Init() tea.Cmd {
	return nil
}

func (m *CookingModel) Update(msg tea.Msg) (common.TUIModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.EnterCookingModeMsg:
		m.Recipe = msg.Recipe
		m.CurrentStep = 0
		m.TotalSteps = len(msg.Recipe.Metadata.Instructions)
		m.modelState = common.ModelStateLoaded
		// Reset chat state for new recipe
		m.chatHistory = []chatEntry{}
		m.chatWaiting = false
		m.showChat = false
		m.chatTextarea.Reset()
		m.chatTextarea.Blur()
		// Reset timer state for new recipe
		m.timerRunning = false
		m.timerDone = false
		m.timerTotal = msg.Recipe.Metadata.TotalTime
		m.timerRemaining = msg.Recipe.Metadata.TotalTime

	case chatResponseMsg:
		m.chatWaiting = false
		if msg.err != nil {
			slog.Error("Cooking chat LLM error", "error", msg.err)
			m.chatHistory = append(m.chatHistory, chatEntry{
				role:    "assistant",
				content: "Sorry, I couldn't generate a response. Make sure Ollama is running.",
			})
		} else {
			m.chatHistory = append(m.chatHistory, chatEntry{
				role:    "assistant",
				content: msg.response,
			})
		}
		m.updateChatViewport()

	case spinner.TickMsg:
		if m.chatWaiting {
			var cmd tea.Cmd
			m.chatSpinner, cmd = m.chatSpinner.Update(msg)
			cmds = append(cmds, cmd)
			m.updateChatViewport()
		}

	case timerTickMsg:
		if m.timerRunning && !m.timerDone {
			m.timerRemaining -= time.Second
			if m.timerRemaining <= 0 {
				m.timerRemaining = 0
				m.timerDone = true
				m.timerRunning = false
				m.timerDoneAt = time.Now()
			} else {
				cmds = append(cmds, timerTick())
			}
		}

	case tea.KeyMsg:
		if m.showChat {
			// Chat panel is open — only handle keys that don't conflict
			// with typing in the textarea (ctrl-combos and esc).
			switch {
			case key.Matches(msg, m.keyMap.Back):
				// Close chat panel (esc only)
				m.showChat = false
				m.chatTextarea.Blur()
				return m, nil

			case key.Matches(msg, m.keyMap.ChatScrollUp):
				m.chatViewport.HalfPageUp()
				return m, nil

			case key.Matches(msg, m.keyMap.ChatScrollDown):
				m.chatViewport.HalfPageDown()
				return m, nil

			case key.Matches(msg, m.keyMap.Enter):
				if m.chatWaiting {
					return m, nil
				}
				userInput := strings.TrimSpace(m.chatTextarea.Value())
				if userInput == "" {
					return m, nil
				}
				m.chatTextarea.Reset()
				m.chatHistory = append(m.chatHistory, chatEntry{
					role:    "user",
					content: userInput,
				})
				m.chatWaiting = true
				m.updateChatViewport()
				cmds = append(cmds, m.chatSpinner.Tick)
				cmds = append(cmds, m.sendChatMessage(userInput))
				return m, tea.Batch(cmds...)

			default:
				// Pass remaining keys to textarea
				var cmd tea.Cmd
				m.chatTextarea, cmd = m.chatTextarea.Update(msg)
				cmds = append(cmds, cmd)
			}
		} else {
			// Chat panel is closed — normal cooking mode keys
			switch {
			case key.Matches(msg, m.keyMap.NextStep):
				if m.CurrentStep < m.TotalSteps-1 {
					m.CurrentStep++
				}
			case key.Matches(msg, m.keyMap.PrevStep):
				if m.CurrentStep > 0 {
					m.CurrentStep--
				}
			case key.Matches(msg, m.keyMap.ToggleIngredients):
				m.showIngredients = !m.showIngredients
			case key.Matches(msg, m.keyMap.ToggleChat):
				m.showChat = true
				m.chatTextarea.Focus()
				if err := m.initLLM(); err != nil {
					slog.Error("Failed to initialize LLM for cooking chat", "error", err)
				}
				return m, nil
			case key.Matches(msg, m.keyMap.ToggleTimer):
				if m.timerTotal > 0 && !m.timerDone {
					if m.timerRunning {
						m.timerRunning = false
					} else {
						m.timerRunning = true
						cmds = append(cmds, timerTick())
					}
				}
			case key.Matches(msg, m.keyMap.ResetTimer):
				if m.timerTotal > 0 {
					m.timerRunning = false
					m.timerDone = false
					m.timerRemaining = m.timerTotal
				}
			case key.Matches(msg, m.keyMap.Back):
				cmds = append(cmds, messages.SendSessionStateMsg(common.SessionStateDetail))
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// timerFunnyMessage returns a fun quip based on how much time remains.
func timerFunnyMessage(remaining, total time.Duration, done bool) string {
	if done {
		phrases := []string{
			"DING DING DING! Dinner is served!",
			"Time's up, chef! Bon appetit!",
			"Ring ring! Your food is calling!",
			"That's a wrap! Plating time!",
		}
		// Rotate through phrases based on second parity for a tiny animation
		return phrases[time.Now().Second()%len(phrases)]
	}

	if total <= 0 {
		return ""
	}
	pct := float64(remaining) / float64(total)

	switch {
	case pct > 0.90:
		return "Just started... deep breaths, chef"
	case pct > 0.75:
		return "Plenty of time. Maybe do a little dance?"
	case pct > 0.50:
		return "Halfway-ish. Smells good in here!"
	case pct > 0.35:
		return "Getting there... resist the urge to peek!"
	case pct > 0.20:
		return "Almost... almost... patience is a spice too"
	case pct > 0.10:
		return "Home stretch! Get those plates ready!"
	case pct > 0.05:
		return "Any second now... don't blink!"
	default:
		return "HOLD ON TO YOUR SPATULA!"
	}
}

// timerFoodEmoji returns a rotating food emoji for the timer animation.
func timerFoodEmoji(running bool) string {
	if !running {
		return "⏸"
	}
	frames := []string{"🍳", "🔥", "🍲", "🫕", "♨️", "🍳", "🔥", "🍲"}
	return frames[time.Now().Second()%len(frames)]
}

// formatCountdown formats a duration as MM:SS or HH:MM:SS.
func formatCountdown(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// renderTimer renders the cooking countdown timer widget.
func (m *CookingModel) renderTimer() string {
	if m.timerTotal <= 0 {
		return ""
	}

	emoji := timerFoodEmoji(m.timerRunning)
	countdown := formatCountdown(m.timerRemaining)
	funMsg := timerFunnyMessage(m.timerRemaining, m.timerTotal, m.timerDone)

	var timerLine string
	if m.timerDone {
		// Big celebration
		timerLine = m.theme.CookingTimerDone.Render(
			fmt.Sprintf("🎉 %s 🎉", countdown))
	} else if m.timerRunning {
		timerLine = m.theme.CookingTimerActive.Render(
			fmt.Sprintf("%s  %s", emoji, countdown))
	} else {
		// Paused
		timerLine = m.theme.CookingTimerLabel.Render(
			fmt.Sprintf("⏸  %s", countdown))
	}

	msgLine := m.theme.CookingTimerMessage.Render(funMsg)

	// Progress bar for the timer
	barWidth := 30
	elapsed := m.timerTotal - m.timerRemaining
	filledN := 0
	if m.timerTotal > 0 {
		filledN = int(float64(elapsed) / float64(m.timerTotal) * float64(barWidth))
	}
	if filledN > barWidth {
		filledN = barWidth
	}

	filledStyle := m.theme.CookingTimerBarFilled
	emptyStyle := m.theme.CookingTimerBarEmpty
	if m.timerDone {
		filledStyle = m.theme.CookingTimerBarCompleted
	}
	timerBar := filledStyle.Render(strings.Repeat("▓", filledN)) +
		emptyStyle.Render(strings.Repeat("░", barWidth-filledN))

	// Percentage label
	pct := 0
	if m.timerTotal > 0 {
		pct = int(float64(elapsed) / float64(m.timerTotal) * 100)
	}
	if pct > 100 {
		pct = 100
	}
	pctLabel := m.theme.CookingTimerActive.Render(fmt.Sprintf(" %d%%", pct))

	barLine := lipgloss.JoinHorizontal(lipgloss.Center, timerBar, pctLabel)

	return lipgloss.JoinVertical(lipgloss.Center,
		timerLine,
		barLine,
		msgLine,
	)
}

// timerTick returns a tea.Cmd that fires a timerTickMsg after one second.
func timerTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return timerTickMsg(t)
	})
}

// initLLM lazily initializes the Ollama LLM for cooking chat.
func (m *CookingModel) initLLM() error {
	if m.llm != nil {
		return nil
	}

	chatConfig := config.GetChatConfig()
	llm, err := ollama.New(
		ollama.WithModel(chatConfig.DefaultModel),
	)
	if err != nil {
		slog.Error("Failed to create Ollama LLM for cooking chat", "error", err)
		return fmt.Errorf("failed to create LLM: %w", err)
	}
	m.llm = llm
	return nil
}

// buildRecipeContext creates a text summary of the recipe for the LLM system prompt.
func (m *CookingModel) buildRecipeContext() string {
	if m.Recipe == nil {
		return ""
	}

	var ctx strings.Builder
	ctx.WriteString(fmt.Sprintf("Recipe: %s\n", m.Recipe.RecipeName))
	if m.Recipe.RecipeDescription != "" {
		ctx.WriteString(fmt.Sprintf("Description: %s\n", m.Recipe.RecipeDescription))
	}

	// Recipe metadata
	meta := m.Recipe.Metadata
	if meta.Author != "" {
		ctx.WriteString(fmt.Sprintf("Author: %s\n", meta.Author))
	}
	if meta.Quantity != "" {
		ctx.WriteString(fmt.Sprintf("Servings: %s\n", meta.Quantity))
	}
	if meta.PrepTime > 0 {
		ctx.WriteString(fmt.Sprintf("Prep Time: %s\n", meta.PrepTime))
	}
	if meta.CookTime > 0 {
		ctx.WriteString(fmt.Sprintf("Cook Time: %s\n", meta.CookTime))
	}
	if meta.TotalTime > 0 {
		ctx.WriteString(fmt.Sprintf("Total Time: %s\n", meta.TotalTime))
	}
	if meta.Rating > 0 {
		ctx.WriteString(fmt.Sprintf("Rating: %d/5\n", meta.Rating))
	}
	if len(meta.Categories) > 0 {
		ctx.WriteString(fmt.Sprintf("Categories: %s\n", strings.Join(meta.Categories, ", ")))
	}
	if meta.URL != "" {
		ctx.WriteString(fmt.Sprintf("Source URL: %s\n", meta.URL))
	}

	ctx.WriteString(fmt.Sprintf("\nCurrent Step (%d of %d):\n%s\n",
		m.CurrentStep+1, m.TotalSteps,
		m.Recipe.Metadata.Instructions[m.CurrentStep]))

	ctx.WriteString("\nIngredients:\n")
	for _, ing := range m.Recipe.Metadata.Ingredients {
		if ing.Amount != "" {
			line := ing.Amount
			if ing.Unit != "" {
				line += " " + ing.Unit
			}
			line += " " + ing.Name
			ctx.WriteString(fmt.Sprintf("- %s\n", line))
		} else {
			ctx.WriteString(fmt.Sprintf("- %s\n", ing.Name))
		}
	}

	ctx.WriteString("\nAll Steps:\n")
	for i, step := range m.Recipe.Metadata.Instructions {
		marker := "  "
		if i == m.CurrentStep {
			marker = "→ "
		}
		ctx.WriteString(fmt.Sprintf("%s%d. %s\n", marker, i+1, step))
	}
	return ctx.String()
}

// sendChatMessage creates a tea.Cmd that calls the Ollama LLM with recipe context.
func (m *CookingModel) sendChatMessage(userMessage string) tea.Cmd {
	llm := m.llm
	ctx := m.ctx
	recipeContext := m.buildRecipeContext()

	// Snapshot conversation history
	history := make([]chatEntry, len(m.chatHistory))
	copy(history, m.chatHistory)

	return func() tea.Msg {
		if llm == nil {
			return chatResponseMsg{err: fmt.Errorf("LLM not initialized")}
		}

		systemPrompt := fmt.Sprintf(
			"You are a concise and helpful cooking assistant. The user is actively cooking "+
				"a recipe and needs quick help with the current step. Keep your answers brief, "+
				"practical, and easy to follow while cooking.\n\n%s", recipeContext)

		msgs := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		}

		// Add conversation history (skip the last entry which is the current user message)
		for _, entry := range history {
			if entry.role == "user" {
				msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, entry.content))
			} else {
				msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeAI, entry.content))
			}
		}

		response, err := llm.GenerateContent(ctx, msgs)
		if err != nil {
			return chatResponseMsg{err: err}
		}

		if len(response.Choices) > 0 {
			return chatResponseMsg{response: response.Choices[0].Content}
		}

		return chatResponseMsg{response: "I couldn't generate a response."}
	}
}

// updateChatViewport rebuilds the chat viewport content from history.
func (m *CookingModel) updateChatViewport() {
	innerWidth := m.chatViewport.Width - 2
	if innerWidth < 10 {
		innerWidth = 10
	}

	userLabel := m.theme.CookingChatUserLabel.Render("You")
	assistantLabel := m.theme.CookingChatAssistantLabel.Render("Chef")

	sepLine := m.theme.MessageSeparator.
		Render(strings.Repeat("─", max(innerWidth-2, 10)))

	var conv strings.Builder

	if len(m.chatHistory) == 0 && !m.chatWaiting {
		hint := m.theme.CookingChatEmpty.
			Width(innerWidth).
			Align(lipgloss.Center).
			Render("Ask anything about the recipe or current step...")
		conv.WriteString("\n\n" + hint)
	}

	for i, entry := range m.chatHistory {
		if i > 0 {
			conv.WriteString("\n" + sepLine + "\n\n")
		}
		if entry.role == "user" {
			conv.WriteString(userLabel + "\n")
			wrapped := utils.WrapTextToWidth(entry.content, innerWidth)
			conv.WriteString(wrapped + "\n")
		} else {
			conv.WriteString(assistantLabel + "\n")
			if m.markdownRenderer != nil {
				if rendered, err := m.markdownRenderer.Render(entry.content); err == nil {
					conv.WriteString(rendered)
				} else {
					wrapped := utils.WrapTextToWidth(entry.content, innerWidth)
					conv.WriteString(wrapped + "\n")
				}
			} else {
				wrapped := utils.WrapTextToWidth(entry.content, innerWidth)
				conv.WriteString(wrapped + "\n")
			}
		}
	}

	if m.chatWaiting {
		if len(m.chatHistory) > 0 {
			conv.WriteString("\n" + sepLine + "\n\n")
		}
		conv.WriteString(assistantLabel + "\n\n")
		conv.WriteString(m.chatSpinner.View() + " Thinking...\n")
	}

	wasAtBottom := m.chatViewport.AtBottom()
	m.chatViewport.SetContent(conv.String())
	if wasAtBottom || m.chatWaiting {
		m.chatViewport.GotoBottom()
	}
}

func (m *CookingModel) View() string {
	if m.Recipe == nil || m.TotalSteps == 0 {
		empty := m.theme.CookingNoRecipe.Render("No recipe loaded")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, empty)
	}

	mainPanel := m.renderMainPanel()

	if m.showChat {
		chatWidth := max(m.width*2/5, 32)
		chatPanel := m.renderChatPanel(chatWidth, m.height-2)
		return lipgloss.JoinHorizontal(lipgloss.Top, mainPanel, chatPanel)
	}

	if m.showIngredients {
		sidebarWidth := min(m.width/3, 30)
		sidebar := m.renderIngredientsSidebar(sidebarWidth, m.height-2)
		return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainPanel)
	}

	return mainPanel
}

func (m *CookingModel) renderMainPanel() string {
	contentWidth := m.width
	if m.showIngredients && !m.showChat {
		contentWidth -= min(m.width/3, 30)
	}
	if m.showChat {
		contentWidth -= max(m.width*2/5, 32)
	}

	// Recipe name
	name := m.theme.CookingRecipeName.Render(m.Recipe.RecipeName)

	// Step counter
	counter := m.theme.CookingStepCounter.Render(
		fmt.Sprintf("Step %d of %d", m.CurrentStep+1, m.TotalSteps))

	// Progress bar
	progressWidth := min(contentWidth-8, 40)
	if progressWidth < 5 {
		progressWidth = 5
	}
	filled := (m.CurrentStep + 1) * progressWidth / m.TotalSteps
	bar := m.theme.CookingProgressFilled.Render(strings.Repeat("━", filled)) +
		m.theme.CookingProgressUnfilled.Render(strings.Repeat("━", progressWidth-filled))

	// Instruction text (with ingredient names highlighted)
	instruction := m.Recipe.Metadata.Instructions[m.CurrentStep]
	ingredientHighlight := m.theme.CookingIngredientHighlight
	highlightedInstruction := utils.HighlightIngredientsWithStyle(
		instruction, m.Recipe.Metadata.Ingredients, ingredientHighlight,
	)
	wrappedWidth := min(contentWidth-16, 60)
	if wrappedWidth < 20 {
		wrappedWidth = 20
	}
	instructionText := m.theme.CookingInstruction.
		Width(wrappedWidth).
		Align(lipgloss.Center).
		Render(highlightedInstruction)

	// Navigation hints
	var navParts []string
	if m.CurrentStep > 0 {
		navParts = append(navParts,
			m.theme.CookingNavArrow.Render("←")+
				m.theme.CookingNavHint.Render(" prev"))
	} else {
		navParts = append(navParts, m.theme.CookingNavHint.Render("     "))
	}

	navParts = append(navParts, m.theme.CookingNavHint.Render("    "))

	if m.CurrentStep < m.TotalSteps-1 {
		navParts = append(navParts,
			m.theme.CookingNavHint.Render("next ")+
				m.theme.CookingNavArrow.Render("→"))
	} else {
		navParts = append(navParts, m.theme.CookingNavHint.Render("     "))
	}
	nav := strings.Join(navParts, "")

	// Help line
	helpKeys := m.theme.CookingHelpKey
	helpDesc := m.theme.CookingNavHint

	ingredientKey := m.keyMap.ToggleIngredients.Help().Key
	timerKey := m.keyMap.ToggleTimer.Help().Key
	resetTimerKey := m.keyMap.ResetTimer.Help().Key
	chatKey := m.keyMap.ToggleChat.Help().Key
	backKey := m.keyMap.Back.Help().Key

	var helpLine string
	if m.showChat {
		helpLine = helpKeys.Render(backKey) + helpDesc.Render(" close chat")
	} else {
		var parts []string
		if m.showIngredients {
			parts = append(parts, helpKeys.Render(ingredientKey)+helpDesc.Render(" hide ingredients"))
		} else {
			parts = append(parts, helpKeys.Render(ingredientKey)+helpDesc.Render(" ingredients"))
		}
		if m.timerTotal > 0 {
			if m.timerRunning {
				parts = append(parts, helpKeys.Render(timerKey)+helpDesc.Render(" pause"))
			} else {
				parts = append(parts, helpKeys.Render(timerKey)+helpDesc.Render(" start"))
			}
			parts = append(parts, helpKeys.Render(resetTimerKey)+helpDesc.Render(" reset"))
		}
		parts = append(parts, helpKeys.Render(chatKey)+helpDesc.Render(" ask AI"))
		parts = append(parts, helpKeys.Render(backKey)+helpDesc.Render(" back"))
		helpLine = strings.Join(parts, "  ")
	}

	// Cooking timer widget (only if recipe has a total time)
	timerWidget := m.renderTimer()

	// Compose vertically centered
	parts := []string{
		name,
		"",
		counter,
		bar,
	}
	if timerWidget != "" {
		parts = append(parts, "", timerWidget)
	}
	parts = append(parts,
		"",
		instructionText,
		"",
		nav,
		"",
		helpLine,
	)
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)

	return lipgloss.Place(contentWidth, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *CookingModel) renderChatPanel(width, height int) string {
	innerWidth := width - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

	// Recreate the markdown renderer if the width changed
	mdWidth := max(innerWidth-4, 12)
	if m.markdownRenderer == nil || m.chatViewport.Width != innerWidth {
		if r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(mdWidth),
		); err == nil {
			m.markdownRenderer = r
		}
	}

	title := m.theme.CookingChatTitle.Render("  Ask about this step")
	titleSep := m.theme.SeparatorLine.
		Render(strings.Repeat("─", innerWidth))

	// Calculate heights: title + sep + blank = 3, separator + textarea = 4
	titleHeight := 3
	inputHeight := 4
	viewportHeight := height - titleHeight - inputHeight
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	// Update viewport/textarea dimensions
	m.chatViewport.Width = innerWidth
	m.chatViewport.Height = viewportHeight
	m.chatTextarea.SetWidth(innerWidth)
	m.updateChatViewport()

	vpView := m.chatViewport.View()

	inputSep := m.theme.SeparatorLine.
		Render(strings.Repeat("─", innerWidth))

	input := m.chatTextarea.View()

	panel := lipgloss.JoinVertical(lipgloss.Left,
		title,
		titleSep,
		"",
		vpView,
		inputSep,
		input,
	)

	return m.theme.CookingChatPanel.
		Width(width).
		Height(height).
		Render(panel)
}

func (m *CookingModel) renderIngredientsSidebar(width, height int) string {
	innerWidth := width - 4
	if innerWidth < 8 {
		innerWidth = 8
	}

	var sidebar strings.Builder

	// Title
	sidebar.WriteString(m.theme.CookingSidebarTitle.Render("🧾 Ingredients"))
	sidebar.WriteString("\n")

	// Count indicator
	countStr := fmt.Sprintf("   %d items", len(m.Recipe.Metadata.Ingredients))
	sidebar.WriteString(m.theme.CookingIngredientDetail.Render(countStr))
	sidebar.WriteString("\n")

	// Separator line
	sep := m.theme.SeparatorLine.
		Render(strings.Repeat("─", innerWidth))
	sidebar.WriteString(sep)
	sidebar.WriteString("\n\n")

	for _, ing := range m.Recipe.Metadata.Ingredients {
		// Bullet prefix
		bullet := m.theme.CookingNavHint.Render("  • ")

		// Amount + unit in accent style
		var amountPart string
		if ing.Amount != "" {
			amt := ing.Amount
			if ing.Unit != "" {
				amt += " " + ing.Unit
			}
			amountPart = m.theme.CookingIngredientAmount.Render(amt) + " "
		}

		// Ingredient name
		namePart := m.theme.CookingIngredient.Render(ing.Name)

		line := bullet + amountPart + namePart

		sidebar.WriteString(line)
		sidebar.WriteString("\n")

		// Details in italic if present
		if ing.Details != "" {
			detail := "    " + m.theme.CookingIngredientDetail.Render("("+ing.Details+")")
			sidebar.WriteString(detail)
			sidebar.WriteString("\n")
		}
	}

	return m.theme.CookingSidebar.
		Width(width).
		Height(height).
		Render(sidebar.String())
}

// IsChatOpen returns whether the cooking chat panel is currently open.
func (m *CookingModel) IsChatOpen() bool {
	return m.showChat
}

func (m *CookingModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *CookingModel) GetSize() (width, height int) {
	return m.width, m.height
}

func (m *CookingModel) GetModelState() common.ModelState {
	return m.modelState
}

func (m *CookingModel) GetSessionState() common.SessionState {
	return common.SessionStateCooking
}

func (m *CookingModel) GetCurrentTheme() *themes.Theme {
	return m.theme
}

func (m *CookingModel) SetTheme(theme *themes.Theme) {
	m.theme = theme
}
