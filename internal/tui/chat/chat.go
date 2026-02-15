package chat

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmc/langchaingo/llms"

	"github.com/GarroshIcecream/yummy/internal/config"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/tui/dialog"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

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

	// LLM service
	ExecutorService *ExecutorService

	// UI state
	modelState         common.ModelState
	waitingForResponse bool
	width              int
	height             int
	sidebarWidth       int
	showSidebar        bool

	// Streaming state
	streamingResponse string
	isStreaming       bool
	generationID      uint64 // monotonically increasing ID to discard stale responses

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
	windowWidth, windowHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		windowWidth = chatConfig.UILayout.ViewportWidth
		windowHeight = chatConfig.UILayout.ViewportHeight
		slog.Error("Failed to get terminal size", "error", err)
	}

	// Calculate markdown width accounting for message formatting
	markdownWidth := max(windowWidth-chatConfig.UILayout.MarkdownPadding, chatConfig.UILayout.MinMarkdownWidth) // Reserve space for message formatting
	markdownRenderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(markdownWidth),
	)
	if err != nil {
		slog.Error("Error creating markdown renderer", "error", err)
		return nil, err
	}

	ta := textarea.New()
	ta.Placeholder = chatConfig.TextAreaPlaceholder
	ta.Focus()

	ta.CharLimit = chatConfig.TextAreaMaxChar
	contentWidth := max(windowWidth-chatConfig.UILayout.ContentPadding, chatConfig.UILayout.MinContentWidth)
	ta.SetWidth(contentWidth)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = theme.TextareaCursorLine
	ta.ShowLineNumbers = false

	// Clean input styling
	ta.FocusedStyle.Base = theme.TextareaBase
	ta.BlurredStyle.Base = theme.TextareaBase
	ta.FocusedStyle.Placeholder = theme.TextareaPlaceholder
	ta.FocusedStyle.Text = theme.TextareaText
	ta.FocusedStyle.Prompt = theme.TextareaPrompt
	ta.Prompt = "› "
	ta.FocusedStyle.EndOfBuffer = theme.TextareaEndOfBuffer

	// Calculate viewport height to fully utilize available terminal height
	viewportHeight := max(windowHeight-chatConfig.UILayout.TitleHeight-ta.Height(), chatConfig.UILayout.MinViewportHeight)

	vp := viewport.New(contentWidth, viewportHeight)
	vp.Style = theme.Chat

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.Spinner

	chatModel := &ChatModel{
		keyMap:             keymaps,
		chatConfig:         chatConfig,
		textarea:           ta,
		viewport:           vp,
		spinner:            s,
		ExecutorService:    executorService,
		markdownRenderer:   markdownRenderer,
		modelState:         common.ModelStateLoaded,
		sidebarWidth:       chatConfig.UILayout.SidebarWidth,
		showSidebar:        windowWidth >= chatConfig.UILayout.MinWidthForSidebar,
		theme:              theme,
		waitingForResponse: false,
		isStreaming:        false,
		streamingResponse:  "",
	}

	return chatModel, nil
}

func (m *ChatModel) Init() tea.Cmd {
	return nil
}

func (m *ChatModel) Update(msg tea.Msg) (common.TUIModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case messages.GenerateResponseMsg:
		// Save the compact display text (with @[Recipe] intact) to memory/DB.
		displayText := msg.DisplayInput
		if displayText == "" {
			displayText = msg.UserInput
		}
		if err := m.ExecutorService.PrepareForGeneration(displayText); err != nil {
			slog.Error("Error preparing for generation", "error", err)
			return m, nil
		}
		m.waitingForResponse = true
		m.isStreaming = true
		m.streamingResponse = ""
		genID := m.generationID
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
		err := m.ExecutorService.SetModelByName(msg.ModelName, m.ExecutorService.ollamaStatus)
		if err != nil {
			slog.Error("Error changing model", "error", err)
			return m, nil
		}
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case tea.KeyMsg:
		// When the @-mention popup is active, intercept navigation keys.
		if m.mention.active {
			switch msg.Type {
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
			installedModels := m.ExecutorService.ollamaStatus.InstalledModels
			modelSelectorDialog, err := dialog.NewModelSelectorDialog(installedModels, currentModelName, m.theme)
			if err != nil {
				slog.Error("Error creating model selector dialog", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(modelSelectorDialog, common.ModalTypeModelSelector))

		case key.Matches(msg, m.keyMap.NewSession):
			err := m.ExecutorService.ResetSession()
			if err != nil {
				slog.Error("Error resetting session", "error", err)
				return m, nil
			}

			// Reset streaming state when resetting session
			m.waitingForResponse = false
			m.isStreaming = false
			m.streamingResponse = ""
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
				m.pendingUserInput = ""
			}

			// Resolve @[RecipeName] mentions to inject recipe context
			augmented := resolveMentions(userInput, m.ExecutorService)

			m.generationID++
			m.textarea.Reset()
			m.mention.reset()
			m.pendingUserInput = userInput
			genID := m.generationID
			cmds = append(cmds, m.spinner.Tick)
			cmds = append(cmds, messages.SendGenerateResponseMsg(augmented, userInput))
			cmds = append(cmds, m.listenForStreamingChunks(genID))
			cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())
			return m, tea.Batch(cmds...)
		}
	}

	// Don't forward scroll-related messages (arrow keys, mouse wheel) to the
	// textarea — those should only reach the viewport for scrolling the
	// conversation history.
	forwardToTextarea := true
	switch typedMsg := msg.(type) {
	case tea.KeyMsg:
		if typedMsg.Type == tea.KeyUp || typedMsg.Type == tea.KeyDown {
			forwardToTextarea = false
		}
	case tea.MouseMsg:
		forwardToTextarea = false
	}
	if forwardToTextarea {
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// After updating the textarea, re-evaluate the @-mention state.
	if _, isKey := msg.(tea.KeyMsg); isKey {
		m.mention.updateMention(m.textarea.Value(), len(m.textarea.Value()), m.ExecutorService)
	}

	return m, tea.Batch(cmds...)
}

func (m *ChatModel) View() string {
	chat := m.theme.Chat.Render(m.viewport.View())

	// Input separator + input
	sepWidth := m.viewport.Width
	if sepWidth < 10 {
		sepWidth = 10
	}
	inputSep := m.theme.SeparatorLine.
		PaddingLeft(1).
		Render(strings.Repeat("─", sepWidth))

	// @-mention autocomplete popup (rendered between separator and textarea)
	mentionPopup := viewMention(&m.mention, m.theme, m.viewport.Width)

	input := m.textarea.View()
	inputArea := lipgloss.NewStyle().MarginBottom(1).Render(input)

	var parts []string
	parts = append(parts, chat, inputSep)
	if mentionPopup != "" {
		parts = append(parts, mentionPopup)
	}
	parts = append(parts, inputArea)
	mainContent := lipgloss.JoinVertical(lipgloss.Left, parts...)

	if m.showSidebar {
		sidebar := RenderSidebar(m.ExecutorService.sessionStats, *m.ExecutorService.ollamaStatus, m.ExecutorService, m.theme, m.sidebarWidth, m.viewport.Height+3)
		return lipgloss.JoinHorizontal(lipgloss.Top, mainContent, sidebar)
	}

	return mainContent
}

// acceptMention replaces the partial @query in the textarea with the
// completed @[RecipeName] reference.
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
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			return nil
		}
		return messages.StreamingChunkMsg{Chunk: chunk, GenerationID: genID}
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

func (m *ChatModel) RenderConversationAsMarkdown() error {
	// Check if user is at the bottom before updating
	var conversation strings.Builder
	conversationMessages, err := m.ExecutorService.GetMemoryConversation()
	if err != nil {
		slog.Error("Failed to get conversation", "error", err)
		return err
	}

	sepLine := m.theme.MessageSeparator.Render(
		strings.Repeat("─", max(m.viewport.Width-4, 20)))

	userLabel := m.chatConfig.UserName
	assistantLabel := m.chatConfig.AssistantName
	msgCount := 0

	for _, message := range conversationMessages {
		role := message.GetType()
		content := message.GetContent()
		if role == llms.ChatMessageTypeSystem {
			continue
		}

		if msgCount > 0 {
			conversation.WriteString("\n" + sepLine + "\n\n")
		}

		var header string
		switch role {
		case llms.ChatMessageTypeHuman:
			header = m.theme.UserMessage.Render(userLabel)
		case llms.ChatMessageTypeAI:
			header = m.theme.AssistantMessage.Render(assistantLabel)
		default:
			header = ""
		}

		conversation.WriteString(header + "\n")
		var msgContent string
		if rendered, err := m.markdownRenderer.Render(content); err == nil {
			msgContent = rendered
		} else {
			msgContent = utils.WrapTextToWidth(content, m.viewport.Width-4)
		}
		conversation.WriteString(HighlightMentions(msgContent, m.theme))

		msgCount++
	}

	// Show the user message immediately (before the executor adds it to memory)
	if m.pendingUserInput != "" {
		if msgCount > 0 {
			conversation.WriteString("\n" + sepLine + "\n\n")
		}
		conversation.WriteString(m.theme.UserMessage.Render(userLabel) + "\n")
		var pendingContent string
		if rendered, err := m.markdownRenderer.Render(m.pendingUserInput); err == nil {
			pendingContent = rendered
		} else {
			pendingContent = utils.WrapTextToWidth(m.pendingUserInput, m.viewport.Width-4)
		}
		conversation.WriteString(HighlightMentions(pendingContent, m.theme))
		msgCount++
	}

	// Streaming or thinking indicator
	if m.waitingForResponse || m.isStreaming {
		if msgCount > 0 {
			conversation.WriteString("\n" + sepLine + "\n\n")
		}
		conversation.WriteString(m.theme.AssistantMessage.Render(assistantLabel) + "\n\n")

		if m.isStreaming && m.streamingResponse != "" {
			var streamContent string
			if rendered, err := m.markdownRenderer.Render(m.streamingResponse); err == nil {
				streamContent = rendered
			} else {
				streamContent = utils.WrapTextToWidth(m.streamingResponse, m.viewport.Width-4)
			}
			conversation.WriteString(HighlightMentions(streamContent, m.theme))
			conversation.WriteString("▋")
		} else {
			conversation.WriteString(m.spinner.View() + " " + m.chatConfig.AssistantThinkingMessage)
		}
	}

	// Empty state — center vertically in viewport
	if msgCount == 0 && !m.waitingForResponse && !m.isStreaming {
		emptyText := m.theme.ChatEmptyState.
			Render("Start a conversation...")
		verticalPad := max(m.viewport.Height/2-1, 0)
		centered := lipgloss.NewStyle().
			Width(m.viewport.Width).
			Align(lipgloss.Center).
			PaddingTop(verticalPad).
			Render(emptyText)
		conversation.WriteString(centered)
	}

	// Only auto-scroll to bottom if user was already at the bottom
	wasAtBottom := m.viewport.AtBottom()
	m.viewport.SetContent(conversation.String())
	if wasAtBottom {
		m.viewport.GotoBottom()
	}
	return nil
}

func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	m.showSidebar = width >= m.chatConfig.UILayout.MinWidthForSidebar
	m.sidebarWidth = max(m.chatConfig.UILayout.MinSidebarWidth, min(m.chatConfig.UILayout.MaxSidebarWidth, width/3))

	contentWidth := width - m.chatConfig.UILayout.ContentPadding
	if m.showSidebar {
		contentWidth -= m.sidebarWidth
	}

	contentWidth = max(m.chatConfig.UILayout.MinContentWidth, contentWidth)
	m.textarea.SetHeight(3)
	viewportHeight := max(
		m.chatConfig.UILayout.MinViewportHeight,
		height-7,
	)

	m.viewport.Width = contentWidth
	m.viewport.Height = viewportHeight
	m.viewport.YPosition = 0
	m.textarea.SetWidth(contentWidth)

	if contentWidth > m.chatConfig.UILayout.MinMarkdownWidthForRenderer {
		m.markdownRenderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(contentWidth-m.chatConfig.UILayout.MarkdownPadding),
		)
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
}
