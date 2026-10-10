package detail

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/GarroshIcecream/yummy/internal/config"
	"github.com/GarroshIcecream/yummy/internal/llms"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/tui/markdown"
	utils "github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/memory"
)

// cookingStreamChunkMsg delivers a streaming chunk from the LLM to the cooking chat.
type cookingStreamChunkMsg struct {
	chunk string
}

// cookingResponseDoneMsg is sent when the LLM generation is complete.
type cookingResponseDoneMsg struct {
	response string
	err      error
}

// cookingToolCallMsg delivers a tool call event from the agent to the cooking UI.
type cookingToolCallMsg struct {
	tool  string
	input string
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
	markdownRenderer *glamour.TermRenderer

	// LLM service and agent
	ai                *llms.LLMService
	executor          *agents.Executor
	chatMemory        *memory.ConversationBuffer
	isStreaming       bool
	streamingResponse string
	pendingToolCalls  []cookingToolCallMsg

	// Cooking timer
	timerRunning   bool
	timerDone      bool
	timerTotal     time.Duration
	timerRemaining time.Duration
	timerDoneAt    time.Time // used to animate the "done" celebration
}

func NewCookingModel(theme *themes.Theme, ai *llms.LLMService) (*CookingModel, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetCookingKeyMap()

	// Textarea for chat input
	ta := textarea.New()
	ta.Placeholder = "Ask about this step..."
	ta.SetStyles(theme.TextAreaStyles())
	ta.CharLimit = 300
	ta.SetWidth(30)
	ta.SetHeight(2)
	ta.ShowLineNumbers = false
	// Match the main chat composer: the ">" prompt only on the first row, blank
	// continuation rows, so an empty input isn't a stack of repeated arrows.
	ta.SetPromptFunc(themes.ComposerPromptWidth, themes.ComposerPrompt)
	ta.Blur()

	// Viewport for chat messages. Its background matches the chat panel so the
	// area below the messages stays on the same surface.
	vp := viewport.New(viewport.WithWidth(30), viewport.WithHeight(10))
	vp.Style = lipgloss.NewStyle().Background(theme.BaseBg)

	// Spinner for loading state
	s := spinner.New()
	s.Spinner = spinner.Dot

	// Markdown renderer for chat responses — the shared themed renderer so the
	// cooking chat uses the same palette as the main chat and recipe detail.
	mdRenderer, err := markdown.NewThemedRenderer(theme, 26)
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
		markdownRenderer: mdRenderer,
		ai:               ai,
	}, nil
}

func (m *CookingModel) Init() tea.Cmd {
	return nil
}

// initAgent creates a fresh agent executor with the cooking-specific system prompt.
func (m *CookingModel) initAgent() {
	systemPrompt := m.buildSystemPrompt()

	m.chatMemory = memory.NewConversationBuffer(
		memory.WithInputKey("input"),
		memory.WithOutputKey("output"),
	)

	if err := m.ai.AddSystemPrompt(m.chatMemory, systemPrompt); err != nil {
		slog.Error("Failed to add cooking system prompt to memory", "error", err)
	}

	m.executor = m.ai.NewAgent(m.chatMemory)
}

// buildSystemPrompt creates the cooking-specific system prompt with recipe context.
func (m *CookingModel) buildSystemPrompt() string {
	recipeContext := m.buildRecipeContext()
	return fmt.Sprintf(
		"You are a concise and helpful cooking assistant. The user is actively cooking "+
			"a recipe and needs quick help with the current step. Keep your answers brief, "+
			"practical, and easy to follow while cooking.\n\n"+
			"You also have access to tools to search for other recipes in the cookbook if needed.\n\n%s", recipeContext)
}

func (m *CookingModel) Update(msg tea.Msg) (common.TUIModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.EnterCookingModeMsg:
		// Cancel any in-flight generation from a previous recipe
		if m.chatWaiting || m.isStreaming {
			m.ai.CancelStreaming()
		}
		m.Recipe = msg.Recipe
		m.CurrentStep = 0
		m.TotalSteps = len(msg.Recipe.Metadata.Instructions)
		m.modelState = common.ModelStateLoaded
		m.chatHistory = []chatEntry{}
		m.chatWaiting = false
		m.isStreaming = false
		m.streamingResponse = ""
		m.pendingToolCalls = nil
		m.showChat = false
		m.chatTextarea.Reset()
		m.chatTextarea.Blur()
		m.timerRunning = false
		m.timerDone = false
		m.timerTotal = msg.Recipe.Metadata.TotalTime
		m.timerRemaining = msg.Recipe.Metadata.TotalTime
		m.initAgent()

	case cookingToolCallMsg:
		m.pendingToolCalls = append(m.pendingToolCalls, msg)
		m.updateChatViewport()
		cmds = append(cmds, m.listenForToolCalls())

	case cookingStreamChunkMsg:
		m.streamingResponse += msg.chunk
		m.updateChatViewport()
		cmds = append(cmds, m.listenForStreamingChunks())

	case cookingResponseDoneMsg:
		m.chatWaiting = false
		m.isStreaming = false
		// Persist tool calls into chat history before the response
		for _, tc := range m.pendingToolCalls {
			m.chatHistory = append(m.chatHistory, chatEntry{
				role:    "tool",
				content: tc.tool + ": " + tc.input,
			})
		}
		m.pendingToolCalls = nil
		if msg.err != nil {
			slog.Error("Cooking chat LLM error", "error", msg.err)
			m.chatHistory = append(m.chatHistory, chatEntry{
				role:    "assistant",
				content: "Sorry, I couldn't generate a response. Make sure Ollama is running.",
			})
		} else {
			// Use the streaming response if we have it, otherwise fall back to the final result
			response := m.streamingResponse
			if response == "" {
				response = msg.response
			}
			m.chatHistory = append(m.chatHistory, chatEntry{
				role:    "assistant",
				content: response,
			})
		}
		m.streamingResponse = ""
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

	case tea.KeyPressMsg:
		if m.showChat {
			// Chat panel is open — only handle keys that don't conflict
			// with typing in the textarea (ctrl-combos and esc).
			switch {
			case key.Matches(msg, m.keyMap.Back):
				if m.chatWaiting || m.isStreaming {
					m.ai.CancelStreaming()
					m.chatWaiting = false
					m.isStreaming = false
					m.streamingResponse = ""
					m.pendingToolCalls = nil
				}
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
				m.isStreaming = true
				m.streamingResponse = ""
				m.updateChatViewport()
				cmds = append(cmds, m.chatSpinner.Tick)
				// ResetCallbackState replaces the streaming and tool-call
				// channels, so it must run before the listeners capture them.
				m.ai.ResetCallbackState()
				cmds = append(cmds, m.sendChatMessage(userInput))
				cmds = append(cmds, m.listenForStreamingChunks())
				cmds = append(cmds, m.listenForToolCalls())
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

// buildRecipeContext creates a text summary of the recipe for the LLM system prompt.
func (m *CookingModel) buildRecipeContext() string {
	if m.Recipe == nil {
		return ""
	}

	var ctx strings.Builder
	fmt.Fprintf(&ctx, "Recipe: %s\n", m.Recipe.RecipeName)
	if m.Recipe.RecipeDescription != "" {
		fmt.Fprintf(&ctx, "Description: %s\n", m.Recipe.RecipeDescription)
	}

	// Recipe metadata
	meta := m.Recipe.Metadata
	if meta.Author != "" {
		fmt.Fprintf(&ctx, "Author: %s\n", meta.Author)
	}
	if meta.Quantity != "" {
		fmt.Fprintf(&ctx, "Servings: %s\n", meta.Quantity)
	}
	if meta.PrepTime > 0 {
		fmt.Fprintf(&ctx, "Prep Time: %s\n", meta.PrepTime)
	}
	if meta.CookTime > 0 {
		fmt.Fprintf(&ctx, "Cook Time: %s\n", meta.CookTime)
	}
	if meta.TotalTime > 0 {
		fmt.Fprintf(&ctx, "Total Time: %s\n", meta.TotalTime)
	}
	if meta.Rating > 0 {
		fmt.Fprintf(&ctx, "Rating: %d/5\n", meta.Rating)
	}
	if len(meta.Categories) > 0 {
		fmt.Fprintf(&ctx, "Categories: %s\n", strings.Join(meta.Categories, ", "))
	}
	if meta.URL != "" {
		fmt.Fprintf(&ctx, "Source URL: %s\n", meta.URL)
	}

	fmt.Fprintf(&ctx, "\nCurrent Step (%d of %d):\n%s\n",
		m.CurrentStep+1, m.TotalSteps,
		m.Recipe.Metadata.Instructions[m.CurrentStep])

	ctx.WriteString("\nIngredients:\n")
	for _, ing := range m.Recipe.Metadata.Ingredients {
		if ing.Amount != "" {
			line := ing.Amount
			if ing.Unit != "" {
				line += " " + ing.Unit
			}
			line += " " + ing.Name
			fmt.Fprintf(&ctx, "- %s\n", line)
		} else {
			fmt.Fprintf(&ctx, "- %s\n", ing.Name)
		}
	}

	ctx.WriteString("\nAll Steps:\n")
	for i, step := range m.Recipe.Metadata.Instructions {
		marker := "  "
		if i == m.CurrentStep {
			marker = "→ "
		}
		fmt.Fprintf(&ctx, "%s%d. %s\n", marker, i+1, step)
	}
	return ctx.String()
}

// sendChatMessage creates a tea.Cmd that runs the agent executor with the user message.
func (m *CookingModel) sendChatMessage(userMessage string) tea.Cmd {
	executor := m.executor
	ai := m.ai

	return func() tea.Msg {
		if executor == nil {
			return cookingResponseDoneMsg{err: fmt.Errorf("agent not initialized")}
		}

		ctx := ai.GetContext()
		result, err := chains.Run(ctx, executor, userMessage)
		if err != nil {
			return cookingResponseDoneMsg{err: err}
		}

		return cookingResponseDoneMsg{response: result}
	}
}

// listenForStreamingChunks returns a tea.Cmd that blocks on the streaming
// channel and delivers the next chunk as a cookingStreamChunkMsg.
func (m *CookingModel) listenForStreamingChunks() tea.Cmd {
	ch := m.ai.GetStreamCh()
	done := m.ai.GetDoneCh()
	return func() tea.Msg {
		select {
		case chunk, ok := <-ch:
			if !ok {
				return nil
			}
			return cookingStreamChunkMsg{chunk: chunk}
		case <-done:
			// Channels were replaced (cancel or new generation); stop listening.
			return nil
		}
	}
}

func (m *CookingModel) listenForToolCalls() tea.Cmd {
	ch := m.ai.GetToolCallCh()
	done := m.ai.GetDoneCh()
	return func() tea.Msg {
		select {
		case ev, ok := <-ch:
			if !ok {
				return nil
			}
			return cookingToolCallMsg{tool: ev.Tool, input: ev.Input}
		case <-done:
			// Channels were replaced (cancel or new generation); stop listening.
			return nil
		}
	}
}

// updateChatViewport rebuilds the chat viewport content from history.
func (m *CookingModel) updateChatViewport() {
	innerWidth := max(m.chatViewport.Width()-2, 10)

	userLabel := m.theme.CookingChatUserLabel.Render("You")
	assistantLabel := m.theme.CookingChatAssistantLabel.Render("Chef")

	sepLine := m.theme.MessageSeparator.
		Background(m.theme.BaseBg).
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
		switch entry.role {
		case "user":
			conv.WriteString(userLabel + "\n")
			wrapped := utils.WrapTextToWidth(entry.content, innerWidth)
			conv.WriteString(wrapped + "\n")
		case "tool":
			toolLabel := m.theme.ToolCallLabel.Render("🔧 " + entry.content)
			conv.WriteString(toolLabel + "\n")
		default:
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

	// Show tool call messages
	for _, tc := range m.pendingToolCalls {
		if len(m.chatHistory) > 0 || conv.Len() > 0 {
			conv.WriteString("\n" + sepLine + "\n\n")
		}
		toolLabel := m.theme.ToolCallLabel.Render("🔧 " + tc.tool)
		conv.WriteString(toolLabel + "\n")
		if tc.input != "" {
			toolInput := m.theme.ToolCallInput.Render("  " + tc.input)
			conv.WriteString(toolInput + "\n")
		}
	}

	// Show streaming response or thinking spinner
	hasContent := len(m.chatHistory) > 0 || len(m.pendingToolCalls) > 0
	if m.isStreaming && m.streamingResponse != "" {
		if hasContent {
			conv.WriteString("\n" + sepLine + "\n\n")
		}
		conv.WriteString(assistantLabel + "\n")
		if m.markdownRenderer != nil {
			if rendered, err := m.markdownRenderer.Render(m.streamingResponse); err == nil {
				conv.WriteString(rendered)
			} else {
				wrapped := utils.WrapTextToWidth(m.streamingResponse, innerWidth)
				conv.WriteString(wrapped + "\n")
			}
		} else {
			wrapped := utils.WrapTextToWidth(m.streamingResponse, innerWidth)
			conv.WriteString(wrapped + "\n")
		}
		conv.WriteString("▋")
	} else if m.chatWaiting {
		if hasContent {
			conv.WriteString("\n" + sepLine + "\n\n")
		}
		conv.WriteString(assistantLabel + "\n\n")
		conv.WriteString(m.theme.Spinner.Background(m.theme.BaseBg).Render(m.chatSpinner.View()) + " Thinking...\n")
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
		chatPanel := m.renderChatPanel(chatWidth, m.height)
		return lipgloss.JoinHorizontal(lipgloss.Top, mainPanel, chatPanel)
	}

	if m.showIngredients {
		sidebarWidth := min(m.width/3, 30)
		sidebar := m.renderIngredientsSidebar(sidebarWidth, m.height)
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
	progressWidth := max(min(contentWidth-8, 40), 5)
	filled := (m.CurrentStep + 1) * progressWidth / m.TotalSteps
	bar := m.theme.CookingProgressFilled.Render(strings.Repeat("━", filled)) +
		m.theme.CookingProgressUnfilled.Render(strings.Repeat("━", progressWidth-filled))

	// Instruction text (with ingredient names highlighted)
	instruction := utils.NormalizeSymbols(m.Recipe.Metadata.Instructions[m.CurrentStep])
	ingredientHighlight := m.theme.CookingIngredientHighlight
	highlightedInstruction := utils.HighlightIngredientsWithStyle(
		instruction, m.Recipe.Metadata.Ingredients, ingredientHighlight,
	)
	wrappedWidth := max(min(contentWidth-16, 60), 20)
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
	innerWidth := max(width-4, 10)

	// Recreate the markdown renderer if the width changed
	mdWidth := max(innerWidth-4, 12)
	if m.markdownRenderer == nil || m.chatViewport.Width() != innerWidth {
		if r, err := markdown.NewThemedRenderer(m.theme, mdWidth); err == nil {
			m.markdownRenderer = r
		}
	}

	title := m.theme.CookingChatTitle.Render("Ask about this step")
	titleSep := m.theme.SeparatorLine.
		Background(m.theme.BaseBg).
		Render(strings.Repeat("─", innerWidth))

	// Calculate heights: blank + title + sep + blank = 4, separator + textarea = 4
	titleHeight := 4
	inputHeight := 4
	viewportHeight := max(height-titleHeight-inputHeight, 3)

	// Update viewport/textarea dimensions
	m.chatViewport.SetWidth(innerWidth)
	m.chatViewport.SetHeight(viewportHeight)
	m.chatTextarea.SetWidth(innerWidth)
	m.updateChatViewport()

	vpView := m.chatViewport.View()

	inputSep := m.theme.SeparatorLine.
		Background(m.theme.BaseBg).
		Render(strings.Repeat("─", innerWidth))

	input := m.chatTextarea.View()

	panel := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		titleSep,
		"",
		vpView,
		inputSep,
		input,
	)

	// Re-establish the panel surface across every line so foreground-only
	// elements (title, separators) and after-reset cells don't fall through
	// to the terminal background.
	panel = utils.FillBackground(panel, m.theme.BaseBg, innerWidth)

	return m.theme.CookingChatPanel.
		Width(width).
		Height(height).
		Render(panel)
}

func (m *CookingModel) renderIngredientsSidebar(width, height int) string {
	innerWidth := max(width-4, 8)

	var sidebar strings.Builder

	// Title with count
	sidebar.WriteString("\n")
	sidebar.WriteString(m.theme.CookingSidebarTitle.Render("Ingredients"))
	sidebar.WriteString(m.theme.CookingIngredientDetail.Render(fmt.Sprintf("  %d", len(m.Recipe.Metadata.Ingredients))))
	sidebar.WriteString("\n")

	sep := m.theme.SeparatorLine.Render(strings.Repeat("─", innerWidth))
	sidebar.WriteString(sep)
	sidebar.WriteString("\n")

	// hanging wraps text to the sidebar width, prefixing the first line with
	// first and every continuation line with an equally wide indent, so long
	// ingredients stay aligned under their own text instead of the bullet.
	const indent = "  "
	hanging := func(first, text string) string {
		wrapped := lipgloss.NewStyle().Width(max(innerWidth-lipgloss.Width(first), 4)).Render(text)
		lines := strings.Split(wrapped, "\n")
		pad := strings.Repeat(" ", lipgloss.Width(first))
		for i := range lines {
			if i == 0 {
				lines[i] = first + lines[i]
			} else {
				lines[i] = pad + lines[i]
			}
		}
		return strings.Join(lines, "\n")
	}

	currentGroup := ""
	for _, ing := range m.Recipe.Metadata.Ingredients {
		if ing.Group != currentGroup {
			currentGroup = ing.Group
			if group := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(currentGroup), ":")); group != "" {
				sidebar.WriteString("\n\n")
				sidebar.WriteString(m.theme.CookingSidebarTitle.Render(group))
			}
		}

		// Amount + unit in accent style, then the name
		var amountPart string
		if ing.Amount != "" {
			amt := ing.Amount
			if ing.Unit != "" {
				amt += " " + ing.Unit
			}
			amountPart = m.theme.CookingIngredientAmount.Render(amt) + " "
		}
		namePart := m.theme.CookingIngredient.Render(ing.Name)

		sidebar.WriteString("\n")
		sidebar.WriteString(hanging(m.theme.CookingNavHint.Render(">")+" ", amountPart+namePart))

		// Details in italic if present, aligned under the name
		if details := utils.CleanIngredientDetails(ing.Details); details != "" {
			sidebar.WriteString("\n")
			sidebar.WriteString(hanging(indent, m.theme.CookingIngredientDetail.Render(details)))
		}
	}

	// Re-establish the surface across every line so foreground-only ingredient
	// rows and after-reset cells don't fall through to the terminal background.
	body := utils.FillBackground(sidebar.String(), m.theme.BaseBg, innerWidth)
	return m.theme.CookingSidebar.
		Width(width).
		Height(height).
		Render(body)
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
	// Drop the renderer so the chat panel rebuilds it with the new palette.
	m.markdownRenderer = nil
	m.chatTextarea.SetStyles(theme.TextAreaStyles())
	m.chatViewport.Style = lipgloss.NewStyle().Background(theme.BaseBg)
}
