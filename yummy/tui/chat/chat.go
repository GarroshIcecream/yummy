package chat

import (
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
	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/GarroshIcecream/yummy/yummy/utils"
)

type ChatModel struct {
	// Configuration
	keyMap     config.KeyMap
	theme      *themes.Theme
	chatConfig *config.ChatConfig

	// UI components
	viewport         viewport.Model
	textarea         textarea.Model
	spinner          spinner.Model
	markdownRenderer *glamour.TermRenderer

	// LLM service
	ExecutorService *ExecutorService

	// UI state
	modelState         consts.ModelState
	waitingForResponse bool
	width              int
	height             int
	sidebarWidth       int
	showSidebar        bool

	// Streaming state
	streamingResponse string
	isStreaming       bool
}

func New(executorService *ExecutorService, keymaps config.KeyMap, theme *themes.Theme, chatConfig *config.ChatConfig) *ChatModel {
	windowWidth, windowHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		windowWidth = chatConfig.ViewportWidth
		windowHeight = chatConfig.ViewportHeight
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
		return nil
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
	ta.SetHeight(chatConfig.TextAreaHeight)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	// Calculate viewport height with proper space allocation
	titleHeight := chatConfig.UILayout.TitleHeight
	inputHeight := chatConfig.UILayout.InputHeight
	borderPadding := chatConfig.UILayout.BorderPadding
	viewportHeight := windowHeight - titleHeight - inputHeight - borderPadding
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
		modelState:         consts.ModelStateLoaded,
		sidebarWidth:       chatConfig.SidebarWidth,
		showSidebar:        windowWidth >= chatConfig.MinWidthForSidebar,
		theme:              theme,
		waitingForResponse: false,
		isStreaming:        false,
		streamingResponse:  "",
	}

	return chatModel
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
		return m, tea.Batch(cmds...)

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

	case messages.LoadSessionMsg:
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
	var conversation strings.Builder
	conversationMessages, err := m.ExecutorService.GetMemoryConversation()
	if err != nil {
		slog.Error("Failed to get conversation", "error", err)
		return err
	}

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
			header = m.theme.UserMessage.Render("👤 You:")
		case llms.ChatMessageTypeAI:
			header = m.theme.AssistantMessage.Render("🤖 Assistant:")
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
		assistantHeader := m.theme.AssistantMessage.Render("🤖 Assistant:")
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
			thinkingContent := m.spinner.View() + " Thinking..."
			conversation.WriteString(thinkingContent)
		}
	}

	result := conversation.String()
	m.viewport.SetContent(result)
	m.viewport.GotoBottom()
	return nil
}

func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	m.showSidebar = width >= m.chatConfig.MinWidthForSidebar
	m.sidebarWidth = max(m.chatConfig.UILayout.MinSidebarWidth, min(m.chatConfig.UILayout.MaxSidebarWidth, width/m.chatConfig.UILayout.SidebarWidthRatio))

	contentWidth := width - m.chatConfig.UILayout.ContentPadding
	if m.showSidebar {
		contentWidth -= m.sidebarWidth
	}

	contentWidth = max(m.chatConfig.UILayout.MinContentWidth, contentWidth)
	viewportHeight := max(m.chatConfig.UILayout.MinViewportHeight, height-m.chatConfig.UILayout.TotalUIHeight)

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

func (m *ChatModel) GetModelState() consts.ModelState {
	return m.modelState
}

func (m *ChatModel) GetSessionState() consts.SessionState {
	return consts.SessionStateChat
}
