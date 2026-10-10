package chat

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/tmc/langchaingo/llms"

	"github.com/GarroshIcecream/yummy/internal/config"
	yllms "github.com/GarroshIcecream/yummy/internal/llms"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/tui/dialog"
	"github.com/GarroshIcecream/yummy/internal/tui/markdown"
	"github.com/GarroshIcecream/yummy/internal/tui/overlay"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

type toolCallEntry struct {
	Tool  string
	Input string
}

type ChatModel struct {
	// Configuration
	keyMap     config.ChatKeyMap
	theme      *themes.Theme
	chatConfig config.ChatConfig

	// UI components
	viewport         viewport.Model
	textarea         textarea.Model
	spinner          spinner.Model
	markdownRenderer *glamour.TermRenderer
	markdownWidth    int

	// LLM service
	ExecutorService *ExecutorService

	// UI state
	modelState         common.ModelState
	waitingForResponse bool
	width              int
	height             int
	sidebarWidth       int
	showSidebar        bool
	sidebarAuto        bool // when true, sidebar visibility follows the width threshold; a manual sidebar toggle (^l) pins it

	// Streaming state
	streamingResponse string
	isStreaming       bool
	generationID      uint64 // monotonically increasing ID to discard stale responses
	pendingToolCalls  []toolCallEntry

	// @-mention autocomplete
	mention mentionState

	// pendingUserInput holds the user message that was just submitted so it
	// can be rendered immediately, before the LLM executor adds it to memory.
	pendingUserInput string
}

func NewChatModel(executorService *ExecutorService, theme *themes.Theme) (*ChatModel, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	keymaps := cfg.Keymap.ToKeyMap().GetChatKeyMap()
	chatConfig := cfg.Chat
	windowWidth, windowHeight, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		windowWidth = chatConfig.UILayout.ViewportWidth
		windowHeight = chatConfig.UILayout.ViewportHeight
		slog.Error("Failed to get terminal size", "error", err)
	}

	// Calculate markdown width accounting for message formatting
	markdownWidth := max(windowWidth-chatConfig.UILayout.MarkdownPadding, chatConfig.UILayout.MinMarkdownWidth) // Reserve space for message formatting
	markdownRenderer, err := newThemedMarkdownRenderer(theme, markdownWidth)
	if err != nil {
		slog.Error("Error creating markdown renderer", "error", err)
		return nil, err
	}

	ta := textarea.New()
	ta.Placeholder = chatConfig.TextAreaPlaceholder
	ta.SetStyles(theme.TextAreaStyles())
	ta.Focus()

	ta.CharLimit = chatConfig.TextAreaMaxChar
	contentWidth := max(windowWidth-chatConfig.UILayout.ContentPadding, chatConfig.UILayout.MinContentWidth)
	ta.SetWidth(contentWidth)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	// Show the ">" prompt only on the first row; continuation rows get a blank
	// (width-matched) prompt so an empty composer reads as one field instead of
	// a stack of repeated arrows.
	ta.SetPromptFunc(themes.ComposerPromptWidth, themes.ComposerPrompt)

	// Calculate viewport height to fully utilize available terminal height
	viewportHeight := max(windowHeight-chatConfig.UILayout.TitleHeight-ta.Height(), chatConfig.UILayout.MinViewportHeight)

	vp := viewport.New(viewport.WithWidth(contentWidth), viewport.WithHeight(viewportHeight))
	vp.Style = lipgloss.NewStyle().Background(theme.BaseBg)

	s := spinner.New()
	s.Spinner = spinner.Dot

	chatModel := &ChatModel{
		keyMap:             keymaps,
		chatConfig:         chatConfig,
		textarea:           ta,
		viewport:           vp,
		spinner:            s,
		ExecutorService:    executorService,
		markdownRenderer:   markdownRenderer,
		markdownWidth:      markdownWidth,
		modelState:         common.ModelStateLoaded,
		sidebarWidth:       chatConfig.UILayout.SidebarWidth,
		showSidebar:        windowWidth >= chatConfig.UILayout.MinWidthForSidebar,
		sidebarAuto:        true,
		theme:              theme,
		waitingForResponse: false,
		isStreaming:        false,
		streamingResponse:  "",
	}

	return chatModel, nil
}

func (m *ChatModel) Init() tea.Cmd {
	return tea.Batch(m.textarea.Focus(), messages.SendRenderConversationAsMarkdownMsg())
}

func (m *ChatModel) Update(msg tea.Msg) (common.TUIModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// SetSize has already run (the manager resizes every model before
		// forwarding the message), so the viewport now has its new dimensions.
		// Re-render so the vertically-centered empty state stays centered.
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case messages.GenerateResponseMsg:
		// Save the compact display text (with @[Recipe] intact) to memory/DB.
		displayText := msg.DisplayInput
		if displayText == "" {
			displayText = msg.UserInput
		}
		// PrepareForGeneration resets the callback state (including replacing
		// the tool-call channel), so listeners must be started AFTER this call.
		if err := m.ExecutorService.PrepareForGeneration(displayText); err != nil {
			slog.Error("Error preparing for generation", "error", err)
			return m, nil
		}
		m.waitingForResponse = true
		m.isStreaming = true
		m.streamingResponse = ""
		genID := m.generationID
		// Start listeners on the fresh channels created by PrepareForGeneration.
		cmds = append(cmds, m.listenForStreamingChunks(genID))
		cmds = append(cmds, m.listenForToolCalls(genID))
		// Send the augmented prompt (with full recipe context) to the LLM,
		// along with the compact display text for memory cleanup.
		cmds = append(cmds, m.SendGenerateResponseMsg(msg.UserInput, displayText, genID))

	case messages.RenderConversationAsMarkdownMsg:
		err := m.RenderConversationAsMarkdown()
		if err != nil {
			slog.Error("Error rendering conversation", "error", err)
			return m, nil
		}

	case messages.ResponseMsg:
		// Ignore stale responses from a cancelled generation.
		if msg.GenerationID != m.generationID {
			return m, nil
		}
		m.waitingForResponse = false
		m.isStreaming = false
		m.streamingResponse = ""
		m.pendingUserInput = ""
		// Keep pendingToolCalls — they render between the last human/AI pair
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case messages.ToolCallMsg:
		if msg.GenerationID != m.generationID {
			return m, nil
		}
		m.pendingToolCalls = append(m.pendingToolCalls, toolCallEntry{Tool: msg.Tool, Input: msg.Input})
		cmds = append(cmds, m.listenForToolCalls(m.generationID))
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case messages.StreamingChunkMsg:
		// Ignore chunks from a cancelled generation.
		if msg.GenerationID != m.generationID {
			return m, nil
		}
		m.streamingResponse += msg.Chunk
		cmds = append(cmds, m.listenForStreamingChunks(m.generationID))
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case messages.SessionSelectedMsg:
		slog.Debug("Loading session", "sessionID", msg.SessionID)
		err := m.ExecutorService.LoadSession(msg.SessionID)
		if err != nil {
			slog.Error("Error loading session", "error", err)
			return m, nil
		}

		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case messages.ModelSelectedMsg:
		slog.Debug("Changing model", "model", msg.ModelName)
		err := m.ExecutorService.SetModelByName(msg.ModelName)
		if err != nil {
			slog.Error("Error changing model", "error", err)
			return m, nil
		}
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case tea.KeyPressMsg:
		// When the @-mention popup is active, intercept navigation keys.
		if m.mention.active {
			switch msg.Key().Code {
			case tea.KeyUp:
				m.mention.moveUp()
				return m, nil
			case tea.KeyDown:
				m.mention.moveDown()
				return m, nil
			case tea.KeyTab, tea.KeyEnter:
				m.acceptMention()
				return m, nil
			case tea.KeyEsc:
				m.mention.reset()
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, m.keyMap.SessionSelector):
			currentSessionID := m.ExecutorService.GetSessionID()
			sessionSelectorDialog, err := dialog.NewSessionSelectorDialog(m.ExecutorService.sessionLog, m.theme, currentSessionID)
			if err != nil {
				slog.Error("Error creating session selector dialog", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(sessionSelectorDialog, common.ModalTypeSessionSelector))

		case key.Matches(msg, m.keyMap.ModelSelector):
			currentModelName := m.ExecutorService.GetCurrentModelName()
			installedModels := m.ExecutorService.GetInstalledModels()
			modelSelectorDialog, err := dialog.NewModelSelectorDialog(installedModels, currentModelName, m.theme)
			if err != nil {
				slog.Error("Error creating model selector dialog", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(modelSelectorDialog, common.ModalTypeModelSelector))

		case key.Matches(msg, m.keyMap.ToggleSidebar):
			// Pin the choice (leave auto-by-width mode) and re-flow the panels
			// at the current size for the new content width.
			m.sidebarAuto = false
			m.showSidebar = !m.showSidebar
			m.SetSize(m.width, m.height)
			cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

		case key.Matches(msg, m.keyMap.NewSession):
			err := m.ExecutorService.ResetSession()
			if err != nil {
				slog.Error("Error resetting session", "error", err)
				return m, nil
			}

			m.waitingForResponse = false
			m.isStreaming = false
			m.streamingResponse = ""
			m.pendingToolCalls = nil
			m.pendingUserInput = ""
			cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

		case key.Matches(msg, m.keyMap.Enter):
			userInput := strings.TrimSpace(m.textarea.Value())
			if userInput == "" {
				return m, nil
			}

			// If a response is currently streaming, cancel it before
			// starting a new generation.
			if m.isStreaming || m.waitingForResponse {
				m.ExecutorService.CancelStreaming()
				m.isStreaming = false
				m.waitingForResponse = false
				m.streamingResponse = ""
				m.pendingToolCalls = nil
				m.pendingUserInput = ""
			}

			// Resolve @[RecipeName] mentions to inject recipe context
			augmented := resolveMentions(userInput, m.ExecutorService)

			m.generationID++
			m.textarea.Reset()
			m.mention.reset()
			m.pendingToolCalls = nil // clear stale tool calls from previous turn
			m.pendingUserInput = userInput
			cmds = append(cmds, m.spinner.Tick)
			cmds = append(cmds, messages.SendGenerateResponseMsg(augmented, userInput))
			cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())
			return m, tea.Batch(cmds...)
		}
	}

	// Don't forward scroll-related messages (arrow keys, mouse wheel) to the
	// textarea — those should only reach the viewport for scrolling the
	// conversation history.
	forwardToTextarea := true
	switch typedMsg := msg.(type) {
	case tea.KeyPressMsg:
		if typedMsg.Key().Code == tea.KeyUp || typedMsg.Key().Code == tea.KeyDown {
			forwardToTextarea = false
		}
	case tea.MouseWheelMsg, tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseMotionMsg:
		forwardToTextarea = false
	}
	if forwardToTextarea {
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// After updating the textarea, re-evaluate the @-mention state.
	if _, isKey := msg.(tea.KeyPressMsg); isKey {
		m.mention.updateMention(m.textarea.Value(), len(m.textarea.Value()), m.ExecutorService)
	}

	return m, tea.Batch(cmds...)
}

// Busy reports whether the assistant is producing an answer.
func (m *ChatModel) Busy() (bool, string) {
	if m.waitingForResponse || m.isStreaming {
		return true, "Assistant is answering"
	}
	return false, ""
}

func (m *ChatModel) View() string {
	contentWidth := m.viewport.Width()
	chatBody := m.viewport.View()

	// @-mention autocomplete popup. Float it over the bottom of the
	// conversation viewport (just above the input panel) rather than inserting
	// it into the vertical stack — stacking grew the layout past the terminal
	// height and pushed both the popup and the composer off-screen.
	if mentionPopup := viewMention(&m.mention, m.theme, contentWidth); mentionPopup != "" {
		chatBody = overlay.Composite(mentionPopup, chatBody, overlay.Left, overlay.Bottom, 1, 0)
	}

	// Conversation panel: a rounded frame whose top border carries the brand
	// title. The composer sits in its own frame below.
	chatPanel := renderTitledPanel(m.theme, chatBody, m.theme.ChatPanelTitle.Render("Yummy"), "", contentWidth, m.theme.BaseBg)

	// The composer sits on the same dark canvas as the conversation, separated
	// only by its border. The textarea does not paint its full width on its own
	// (empty lines and short lines leave bare cells), so fill the view before
	// framing it — otherwise the field falls through to the terminal background
	// and reads as ragged.
	composer := utils.FillBackground(m.textarea.View(), m.theme.BaseBg, contentWidth)
	inputPanel := renderTitledPanel(m.theme, composer, "", "", contentWidth, m.theme.BaseBg)
	mainContent := lipgloss.JoinVertical(lipgloss.Left, chatPanel, inputPanel)

	if m.showSidebar {
		// The sidebar spans the full height of the left column (conversation
		// + composer), so both columns end on the same line.
		sidebarBody := lipgloss.Height(mainContent) - 2 // minus its top/bottom border
		sidebar := RenderSidebar(m.ExecutorService, m.keyMap, m.theme, m.sidebarWidth, sidebarBody)
		gap := lipgloss.NewStyle().MarginLeft(2).Render(sidebar)
		combined := lipgloss.JoinHorizontal(lipgloss.Top, mainContent, gap)
		// The 2-column gutter (margins are always transparent in lipgloss)
		// is bare cells; paint it with the base canvas color.
		return utils.FillBackground(combined, m.theme.BaseBg, m.width)
	}

	return mainContent
}

func (m *ChatModel) acceptMention() {
	replacement, _ := m.mention.accept()
	if replacement == "" {
		return
	}

	val := m.textarea.Value()

	// Find the last @ that started the mention
	atIdx := strings.LastIndex(val, "@")
	if atIdx < 0 {
		return
	}

	// Replace from @ to the current cursor position with the completed mention
	newVal := val[:atIdx] + replacement + " "
	m.textarea.Reset()
	m.textarea.SetValue(newVal)

	// Move cursor to end
	for i := 0; i < len(newVal); i++ {
		m.textarea.CursorEnd()
	}
}

// listenForStreamingChunks returns a tea.Cmd that blocks on the streaming
// channel and delivers the next chunk as a StreamingChunkMsg to the Bubble Tea
// runtime. Call it once to start listening; each StreamingChunkMsg handler
// should call it again to keep the pipeline going.
func (m *ChatModel) listenForStreamingChunks(genID uint64) tea.Cmd {
	ch := m.ExecutorService.GetStreamCh()
	done := m.ExecutorService.GetDoneCh()
	return func() tea.Msg {
		select {
		case chunk, ok := <-ch:
			if !ok {
				return nil
			}
			return messages.StreamingChunkMsg{Chunk: chunk, GenerationID: genID}
		case <-done:
			// Channels were replaced (cancel or new generation); stop listening.
			return nil
		}
	}
}

func (m *ChatModel) listenForToolCalls(genID uint64) tea.Cmd {
	ch := m.ExecutorService.GetToolCallCh()
	done := m.ExecutorService.GetDoneCh()
	return func() tea.Msg {
		select {
		case ev, ok := <-ch:
			if !ok {
				return nil
			}
			return messages.ToolCallMsg{Tool: ev.Tool, Input: ev.Input, GenerationID: genID}
		case <-done:
			// Channels were replaced (cancel or new generation); stop listening.
			return nil
		}
	}
}

func (m *ChatModel) SendGenerateResponseMsg(promptInput, displayInput string, genID uint64) tea.Cmd {
	return func() tea.Msg {
		response, err := m.ExecutorService.GenerateResponse(promptInput, displayInput)
		if err != nil {
			slog.Error("Error generating response", "error", err)
			return messages.ResponseMsg{Response: "", GenerationID: genID}
		}
		return messages.ResponseMsg{Response: response, GenerationID: genID}
	}
}

// RenderConversationAsMarkdown lays the conversation out as a compact thread:
//
//	> You
//
//	  question…
//
//
//	< Yummy
//
//	  -> Searching cookbook  "Baklava"
//	  answer…
//
// Each turn is a one-line header, a blank line, then its body indented under
// it; turns are separated by two blank lines. Tool calls belong to the assistant turn
// they produced, so they render inside it, before the answer.
func (m *ChatModel) RenderConversationAsMarkdown() error {
	conversationMessages, err := m.ExecutorService.GetMemoryConversation()
	if err != nil {
		slog.Error("Failed to get conversation", "error", err)
		return err
	}

	generating := m.waitingForResponse || m.isStreaming

	// The assistant reply to the latest question is where this turn's tool
	// calls are shown once the answer has arrived.
	lastHuman := -1
	for i, message := range conversationMessages {
		if message.GetType() == llms.ChatMessageTypeHuman {
			lastHuman = i
		}
	}

	var turns []string
	aiIndex := 0
	for i, message := range conversationMessages {
		switch message.GetType() {
		case llms.ChatMessageTypeHuman:
			turns = append(turns, m.renderTurn(true, m.renderMarkdown(message.GetContent())))
		case llms.ChatMessageTypeAI:
			// Each answer keeps the tool calls that produced it (recorded by
			// the executor and persisted with the message).
			calls := m.ExecutorService.ToolCallsFor(aiIndex)
			aiIndex++
			var tools string
			if len(calls) > 0 {
				tools = m.renderToolCalls(toToolCallEntries(calls))
			} else if i > lastHuman && !generating {
				tools = m.renderToolCalls(m.pendingToolCalls)
			}
			turns = append(turns, m.renderTurn(false, joinBlocks(tools, m.renderMarkdown(message.GetContent()))))
		}
	}

	// Show the user message immediately (before the executor adds it to memory)
	if m.pendingUserInput != "" {
		turns = append(turns, m.renderTurn(true, m.renderMarkdown(m.pendingUserInput)))
	}

	// The in-progress assistant turn: tool calls so far, then the streamed
	// answer or a thinking indicator.
	if generating {
		var reply string
		if m.isStreaming && m.streamingResponse != "" {
			reply = m.renderMarkdown(m.streamingResponse) + m.theme.Spinner.Render("▋")
		} else {
			reply = "  " + m.theme.Spinner.Render(m.spinner.View()) +
				m.theme.ChatEmptyState.Render(" "+m.chatConfig.AssistantThinkingMessage)
		}
		turns = append(turns, m.renderTurn(false, joinBlocks(m.renderToolCalls(m.pendingToolCalls), reply)))
	}

	var conversation string
	if len(turns) == 0 {
		conversation = m.renderEmptyState()
	} else {
		// Two blank lines between turns (vs one under each header) keeps a
		// header visibly grouped with its own body.
		conversation = strings.Join(turns, "\n\n\n")
	}

	// Only auto-scroll to bottom if user was already at the bottom
	wasAtBottom := m.viewport.AtBottom()
	m.viewport.SetContent(utils.FillBackground(conversation, m.theme.BaseBg, m.viewport.Width()))
	if wasAtBottom {
		m.viewport.GotoBottom()
	}
	return nil
}

// renderTurn renders a message header ("> You" / "< Yummy") above its body.
func (m *ChatModel) renderTurn(user bool, body string) string {
	header := m.theme.AssistantMessage.Render("< " + m.chatConfig.AssistantName)
	if user {
		header = m.theme.UserMessage.Render("> " + m.chatConfig.UserName)
	}
	if body == "" {
		return header
	}
	return header + "\n\n" + body
}

// renderMarkdown renders message text with the themed renderer (indented two
// columns by its document margin) and strips glamour's surrounding blank
// lines, so the thread's spacing is controlled here rather than by glamour.
func (m *ChatModel) renderMarkdown(content string) string {
	rendered, err := m.markdownRenderer.Render(content)
	if err != nil {
		rendered = utils.WrapTextToWidth(content, m.viewport.Width()-4)
	}
	return HighlightMentions(trimBlankLines(rendered), m.theme)
}

// renderToolCalls lists a turn's tool calls as compact activity lines.
func (m *ChatModel) renderToolCalls(calls []toolCallEntry) string {
	if len(calls) == 0 {
		return ""
	}
	lines := make([]string, 0, len(calls))
	for _, tc := range calls {
		label, input := friendlyToolCall(tc.Tool, tc.Input)
		line := "  " + m.theme.ToolCallLabel.Render("-> "+label)
		if input != "" {
			line += m.theme.ToolCallInput.Render("  " + input)
		}
		lines = append(lines, ansi.Truncate(line, m.viewport.Width(), "…"))
	}
	return strings.Join(lines, "\n")
}

func toToolCallEntries(events []yllms.ToolCallEvent) []toolCallEntry {
	out := make([]toolCallEntry, len(events))
	for i, ev := range events {
		out[i] = toolCallEntry{Tool: ev.Tool, Input: ev.Input}
	}
	return out
}

// friendlyToolCall turns a raw tool name and input into readable activity
// text, e.g. searchRecipeByName("Baklava") -> Searching cookbook  "Baklava".
func friendlyToolCall(tool, input string) (string, string) {
	input = strings.Trim(strings.TrimSpace(input), `"`)
	switch tool {
	case "searchRecipeByName":
		return "Searching cookbook", quoteIfSet(input)
	case "getRecipeById":
		if input != "" {
			input = "#" + strings.TrimPrefix(input, "#")
		}
		return "Opening recipe", input
	case "listRecipes":
		return "Listing recipes", input
	case "findRecipesByIngredients":
		return "Finding recipes with", input
	case "getCookbookOverview":
		return "Reading cookbook overview", ""
	}
	// Fallback: split camelCase into words ("getShoppingList" -> "Get shopping list").
	var words []string
	start := 0
	for i := 1; i < len(tool); i++ {
		if tool[i] >= 'A' && tool[i] <= 'Z' {
			words = append(words, strings.ToLower(tool[start:i]))
			start = i
		}
	}
	words = append(words, strings.ToLower(tool[start:]))
	label := strings.Join(words, " ")
	if label != "" {
		label = strings.ToUpper(label[:1]) + label[1:]
	}
	return label, quoteIfSet(input)
}

func quoteIfSet(s string) string {
	if s == "" {
		return ""
	}
	return `"` + s + `"`
}

// joinBlocks joins non-empty blocks with a newline.
func joinBlocks(blocks ...string) string {
	var out []string
	for _, b := range blocks {
		if b != "" {
			out = append(out, b)
		}
	}
	return strings.Join(out, "\n")
}

// trimBlankLines drops leading and trailing lines that are visually empty
// (glamour pads its output with blank, space-filled lines).
func trimBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	blank := func(l string) bool { return strings.TrimSpace(ansi.Strip(l)) == "" }
	start, end := 0, len(lines)
	for start < end && blank(lines[start]) {
		start++
	}
	for end > start && blank(lines[end-1]) {
		end--
	}
	return strings.Join(lines[start:end], "\n")
}

// renderEmptyState builds the welcome shown when the conversation has no
// messages yet. It pairs a title and hint with a few example prompts so the
// blank canvas reads as an invitation rather than dead space, and centers the
// whole block in the viewport both horizontally and vertically.
func (m *ChatModel) renderEmptyState() string {
	icon := m.theme.AssistantMessage.Render("🍳")
	title := m.theme.ChatEmptyStateTitle.Render("What are we cooking today?")
	hint := m.theme.ChatEmptyState.Render("Ask about recipes, ingredients, and techniques.")

	// Example prompts give the empty canvas a concrete starting point. The
	// accent prompt marker ties them to the composer's ">" prompt below.
	promptStyle := lipgloss.NewStyle().Foreground(m.theme.AssistantMessage.GetForeground())
	suggestStyle := lipgloss.NewStyle().Foreground(m.theme.ToolCallInput.GetForeground())

	suggestions := []string{
		"Find a quick weeknight dinner",
		"What can I make with chicken and rice?",
		"Type @ to pull in one of your saved recipes",
	}
	rows := []string{m.theme.ChatEmptyState.Render("TRY ASKING")}
	for _, s := range suggestions {
		rows = append(rows, promptStyle.Render("> ")+suggestStyle.Render(s))
	}
	suggestBlock := lipgloss.JoinVertical(lipgloss.Left, rows...)

	emptyText := lipgloss.JoinVertical(
		lipgloss.Center,
		icon,
		"",
		title,
		hint,
		"",
		"",
		suggestBlock,
	)

	verticalPad := max((m.viewport.Height()-lipgloss.Height(emptyText))/2, 0)
	return lipgloss.NewStyle().
		Width(m.viewport.Width()).
		Align(lipgloss.Center).
		PaddingTop(verticalPad).
		Render(emptyText)
}

func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Auto mode follows the width threshold; a manual sidebar toggle (^l) pins visibility.
	if m.sidebarAuto {
		m.showSidebar = width >= m.chatConfig.UILayout.MinWidthForSidebar
	}
	m.sidebarWidth = max(m.chatConfig.UILayout.MinSidebarWidth, min(m.chatConfig.UILayout.MaxSidebarWidth, width/3))

	// Each panel costs 2 border + 2 horizontal padding columns; the sidebar is
	// offset by a 2-column gap. Size the conversation body to fill what's left.
	const panelChrome = 4
	const sidebarGap = 2
	contentWidth := width - panelChrome
	if m.showSidebar {
		contentWidth -= m.sidebarWidth + sidebarGap
	}

	contentWidth = max(m.chatConfig.UILayout.MinContentWidth, contentWidth)
	m.textarea.SetHeight(3)
	viewportHeight := max(
		m.chatConfig.UILayout.MinViewportHeight,
		height-7,
	)

	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(viewportHeight)
	m.viewport.YPosition = 0
	m.textarea.SetWidth(contentWidth)

	if contentWidth > m.chatConfig.UILayout.MinMarkdownWidthForRenderer {
		m.markdownWidth = contentWidth - m.chatConfig.UILayout.MarkdownPadding
		m.markdownRenderer, _ = newThemedMarkdownRenderer(m.theme, m.markdownWidth)
	}
}

func (m *ChatModel) GetSize() (width, height int) {
	return m.width, m.height
}

func (m *ChatModel) GetModelState() common.ModelState {
	return m.modelState
}

func (m *ChatModel) GetSessionState() common.SessionState {
	return common.SessionStateChat
}

func (m *ChatModel) GetCurrentTheme() *themes.Theme {
	return m.theme
}

func (m *ChatModel) SetTheme(theme *themes.Theme) {
	m.theme = theme
	m.viewport.Style = lipgloss.NewStyle().Background(theme.BaseBg)
	m.textarea.SetStyles(theme.TextAreaStyles())
	if r, err := newThemedMarkdownRenderer(theme, m.markdownWidth); err == nil {
		m.markdownRenderer = r
		if err := m.RenderConversationAsMarkdown(); err != nil {
			slog.Error("Failed to re-render conversation after theme change", "error", err)
		}
	}
}

// newThemedMarkdownRenderer creates the app's themed markdown renderer. The
// StyleConfig now lives in the shared markdown package so the chat, detail and
// cooking views all render with one palette.
func newThemedMarkdownRenderer(theme *themes.Theme, wordWrap int) (*glamour.TermRenderer, error) {
	return markdown.NewThemedRenderer(theme, wordWrap)
}
