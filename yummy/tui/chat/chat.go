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

	"github.com/GarroshIcecream/yummy/yummy/config"
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/GarroshIcecream/yummy/yummy/tui/dialog"
	"github.com/GarroshIcecream/yummy/yummy/utils"
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
}

func New(executorService *ExecutorService, theme *themes.Theme) (*ChatModel, error) {
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
	markdownWidth := windowWidth - chatConfig.UILayout.MarkdownPadding // Reserve space for message formatting
	if markdownWidth < chatConfig.UILayout.MinMarkdownWidth {
		markdownWidth = chatConfig.UILayout.MinMarkdownWidth
	}

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
	contentWidth := windowWidth - chatConfig.UILayout.ContentPadding
	if contentWidth < chatConfig.UILayout.MinContentWidth {
		contentWidth = chatConfig.UILayout.MinContentWidth
	}
	ta.SetWidth(contentWidth)
	ta.SetHeight(chatConfig.UILayout.InputHeight)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	// Calculate viewport height to fully utilize available terminal height
	titleHeight := chatConfig.UILayout.TitleHeight
	inputHeight := ta.Height()
	viewportHeight := windowHeight - titleHeight - inputHeight
	if viewportHeight < chatConfig.UILayout.MinViewportHeight {
		viewportHeight = chatConfig.UILayout.MinViewportHeight
	}

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

func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case messages.GenerateResponseMsg:
		m.waitingForResponse = true
		m.isStreaming = true
		m.streamingResponse = ""
		cmds = append(cmds, m.SendGenerateResponseMsg(msg.UserInput))

	case messages.RenderConversationAsMarkdownMsg:
		err := m.RenderConversationAsMarkdown()
		if err != nil {
			slog.Error("Error rendering conversation", "error", err)
			return m, nil
		}

	case messages.ResponseMsg:
		m.waitingForResponse = false
		m.isStreaming = false
		m.streamingResponse = ""
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case messages.StreamingChunkMsg:
		m.streamingResponse += msg.Chunk
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case messages.SessionSelectedMsg:
		slog.Debug("Loading session", "sessionID", msg.SessionID)
		err := m.ExecutorService.LoadSession(msg.SessionID)
		if err != nil {
			slog.Error("Error loading session", "error", err)
			return m, nil
		}

		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.SessionSelector):
			currentSessionID := m.ExecutorService.GetSessionID()
			sessionSelectorDialog, err := dialog.NewSessionSelectorDialog(m.ExecutorService.sessionLog, m.theme, currentSessionID)
			if err != nil {
				slog.Error("Error creating session selector dialog", "error", err)
				return m, nil
			}
			cmds = append(cmds, messages.SendOpenModalViewMsg(sessionSelectorDialog, common.ModalTypeSessionSelector))

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
			cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())

		case key.Matches(msg, m.keyMap.Enter):
			userInput := strings.TrimSpace(m.textarea.Value())
			m.textarea.Reset()
			cmds = append(cmds, m.spinner.Tick)
			cmds = append(cmds, messages.SendGenerateResponseMsg(userInput))
			cmds = append(cmds, messages.SendRenderConversationAsMarkdownMsg())
			return m, tea.Batch(cmds...)
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ChatModel) View() string {
	title := m.theme.ChatTitle.Render("🍳 Cooking Assistant")

	if m.width > 0 {
		titleWidth := lipgloss.Width(title)
		if titleWidth < m.width {
			padding := (m.width - titleWidth) / 2
			title = lipgloss.NewStyle().MarginLeft(padding).Render(title)
		}
	}

	chat := m.theme.Chat.Render(m.viewport.View())
	input := m.textarea.View()
	mainContent := lipgloss.JoinVertical(lipgloss.Left, chat, input)

	var chatLayout string
	if m.showSidebar {
		sidebar := RenderSidebar(m.ExecutorService.sessionStats, *m.ExecutorService.ollamaStatus, m.ExecutorService, m.theme, m.sidebarWidth, m.viewport.Height)
		chatLayout = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
	} else {
		chatLayout = mainContent
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, chatLayout)
}

func (m *ChatModel) SendGenerateResponseMsg(userInput string) tea.Cmd {
	return func() tea.Msg {
		response, err := m.ExecutorService.GenerateResponse(userInput)
		if err != nil {
			slog.Error("Error generating response", "error", err)
			return messages.ResponseMsg{Response: ""} // Return empty response on error
		}
		return messages.ResponseMsg{Response: response}
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

	userNameFull := fmt.Sprintf("%s %s:", m.chatConfig.UserAvatar, m.chatConfig.UserName)
	assistantNameFull := fmt.Sprintf("%s %s:", m.chatConfig.AssistantAvatar, m.chatConfig.AssistantName)
	for i, message := range conversationMessages {
		role := message.GetType()
		content := message.GetContent()
		if role == llms.ChatMessageTypeSystem {
			continue
		}

		// Format based on message role with markdown rendering
		var header string
		switch role {
		case llms.ChatMessageTypeHuman:
			header = m.theme.UserMessage.Render(userNameFull)
		case llms.ChatMessageTypeAI:
			header = m.theme.AssistantMessage.Render(assistantNameFull)
		default:
			header = ""
		}

		// Try to render as markdown, fallback to plain text if it fails\
		conversation.WriteString(header + "\n")
		if rendered, err := m.markdownRenderer.Render(content); err == nil {
			conversation.WriteString(rendered)
		} else {
			wrappedText := utils.WrapTextToWidth(content, m.viewport.Width-4)
			conversation.WriteString(wrappedText)
		}

		if i < len(conversationMessages)-1 {
			conversation.WriteString("\n")
		}
	}

	// Add streaming response or thinking indicator if loading
	if m.waitingForResponse || m.isStreaming {
		conversation.WriteString("\n")
		assistantHeader := m.theme.AssistantMessage.Render(assistantNameFull)
		conversation.WriteString(assistantHeader + "\n\n")

		if m.isStreaming && m.streamingResponse != "" {
			// Show streaming response
			if rendered, err := m.markdownRenderer.Render(m.streamingResponse); err == nil {
				conversation.WriteString(rendered)
			} else {
				wrappedText := utils.WrapTextToWidth(m.streamingResponse, m.viewport.Width-4)
				conversation.WriteString(wrappedText)
			}
			// Add cursor to show it's still typing
			conversation.WriteString("▋")
		} else {
			// Show thinking indicator
			thinkingContent := m.spinner.View() + " " + m.chatConfig.AssistantThinkingMessage
			conversation.WriteString(thinkingContent)
		}
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
	desiredInputHeight := m.chatConfig.UILayout.InputHeight
	m.textarea.SetHeight(desiredInputHeight)
	viewportHeight := max(
		m.chatConfig.UILayout.MinViewportHeight,
		height-m.chatConfig.UILayout.TitleHeight-desiredInputHeight,
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
