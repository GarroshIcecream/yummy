package chat

import (
	"log"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/tmc/langchaingo/llms"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
	"github.com/GarroshIcecream/yummy/yummy/tui/utils"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
)

type ChatModel struct {
	cookbook         *db.CookBook
	keyMap           config.KeyMap
	viewport         viewport.Model
	textarea         textarea.Model
	spinner          spinner.Model
	markdownRenderer *glamour.TermRenderer
	conversation     []llms.MessageContent
	llmService       *LLMService
	loading          bool
	windowWidth      int
	windowHeight     int
	width            int
	height           int
	sidebarWidth     int
	messageCount     int
	tokenCount       int
	ollamaStatus     OllamaServiceStatus
	modelState       ui.ModelState
}

func New(cookbook *db.CookBook, keymaps config.KeyMap) *ChatModel {
	llmService, err := NewLLMService()
	if err != nil {
		log.Printf("Warning: LLM service initialization failed: %v", err)
	}

	windowWidth, windowHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		windowWidth = ui.DefaultViewportWidth
		windowHeight = ui.DefaultViewportHeight
	}

	// Calculate markdown width accounting for message formatting
	markdownWidth := windowWidth - 8 // Reserve space for message formatting
	if markdownWidth < 20 {
		markdownWidth = 20
	}
	
	markdownRenderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(markdownWidth),
	)
	if err != nil {
		log.Printf("Error creating markdown renderer: %v", err)
	}

	conversation := []llms.MessageContent{}
	if llmService != nil {
		conversation = []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{llms.TextPart(
					llmService.GetSystemPrompt(),
				)},
			},
		}
		// Add a welcome message from the assistant
		conversation = AppendMessage(conversation, llms.ChatMessageTypeAI, ui.WelcomeMessage)
	}

	ta := textarea.New()
	ta.Placeholder = ui.TextAreaPlaceholder
	ta.Focus()

	ta.CharLimit = ui.TextAreaMaxChar
	contentWidth := windowWidth - 8
	if contentWidth < 20 {
		contentWidth = 20
	}
	ta.SetWidth(contentWidth)
	ta.SetHeight(ui.TextAreaHeight)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	// Calculate viewport height with proper space allocation
	titleHeight := 5
	inputHeight := 6
	borderPadding := 6
	viewportHeight := windowHeight - titleHeight - inputHeight - borderPadding
	if viewportHeight < 8 {
		viewportHeight = 8
	}

	vp := viewport.New(contentWidth, viewportHeight)
	vp.Style = styles.ChatStyle

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	// Initialize Ollama status
	ollamaStatus := GetOllamaServiceStatus()

	// Initialize the chat model
	chatModel := &ChatModel{
		cookbook:         cookbook,
		keyMap:           keymaps,
		textarea:         ta,
		viewport:         vp,
		spinner:          s,
		conversation:     conversation,
		llmService:       llmService,
		markdownRenderer: markdownRenderer,
		modelState:       ui.ModelStateLoaded,
		sidebarWidth:     ui.SidebarWidth,
		messageCount:     0,
		tokenCount:       0,
		ollamaStatus:     ollamaStatus,
		loading:          false,
	}

	// Set initial content in viewport
	if len(conversation) > 1 { // More than just system message
		chatModel.updateViewportFromConversation()
	}

	return chatModel
}

func (m *ChatModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *ChatModel) renderConversationAsMarkdown() string {
	var content strings.Builder
	for i, message := range m.conversation {
		// Skip the initial system prompt
		if i == 0 && message.Role == llms.ChatMessageTypeSystem {
			continue
		}

		// Get the text content from the message
		var messageText string
		for _, part := range message.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				messageText += textPart.Text
			}
		}

		// Format based on message role with markdown rendering
		var header string
		switch message.Role {
		case llms.ChatMessageTypeHuman:
			header = styles.UserMessageStyle.Render("👤 You:")
		case llms.ChatMessageTypeAI:
			header = styles.AssistantMessageStyle.Render("🤖 Assistant:")
		default:
			header = ""
		}

		// Try to render as markdown, fallback to plain text if it fails\
		content.WriteString(header + "\n")
		if rendered, err := m.markdownRenderer.Render(messageText); err == nil {
				content.WriteString(rendered)
		} else {
			wrappedText := m.wrapText(messageText, m.viewport.Width-4)
			content.WriteString(wrappedText)
		}

		if i < len(m.conversation)-1 {
			content.WriteString("\n")
		}
	}

	result := content.String()
	return result
}

// wrapText wraps text to fit within the specified width
func (m *ChatModel) wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}
	
	var lines []string
	var currentLine string
	
	for _, word := range words {
		// If adding this word would exceed the width, start a new line
		if len(currentLine)+len(word)+1 > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Word is longer than width, add it anyway
				lines = append(lines, word)
			}
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}
	
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	return strings.Join(lines, "\n")
}

// updateViewportFromConversation renders the conversation and updates the viewport
func (m *ChatModel) updateViewportFromConversation() {
	content := m.renderConversationAsMarkdown()

	if m.loading {
		if content != "" {
			content += "\n"
		}

		assistantHeader := styles.AssistantMessageStyle.Render("🤖 Assistant:")
		content += assistantHeader + "\n"		
		spinnerText := m.spinner.View() + " Thinking..."
		content += spinnerText
	}

	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

// DisplayUserMessage adds a user message to the chat display
func (m *ChatModel) DisplayUserMessage(userInput string) {
	// Add user message to conversation
	m.conversation = AppendMessage(m.conversation, llms.ChatMessageTypeHuman, userInput)

	m.messageCount++
	m.tokenCount += len(userInput) / 4

	// Update viewport from conversation
	m.updateViewportFromConversation()
	m.textarea.Reset()
}

// generateResponseCommand creates a command that generates a response in the background
func (m *ChatModel) generateResponseCommand() tea.Cmd {
	return func() tea.Msg {
		response := m.llmService.GenerateResponse(m.conversation)
		return ui.ResponseMsg{
			Content: response.Response,
			Error:   response.Error,
		}
	}
}

// Update handles all messages and updates the model
func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case ui.ResponseMsg:
		if msg.Error != nil {
			m.messageCount++
		}
		m.loading = false
		m.conversation = AppendMessage(m.conversation, llms.ChatMessageTypeAI, msg.Content)
		m.updateViewportFromConversation()

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
		m.updateViewportFromConversation()
	
	case tea.WindowSizeMsg:
		// Store window dimensions
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Calculate sidebar width (30% of total width, minimum 25, maximum 40)
		// But ensure we don't take too much space on small screens
		if msg.Width < 80 {
			m.sidebarWidth = 25 // Fixed width for small screens
		} else {
			m.sidebarWidth = msg.Width * 30 / 100
			if m.sidebarWidth < 25 {
				m.sidebarWidth = 25
			} else if m.sidebarWidth > 40 {
				m.sidebarWidth = 40
			}
		}

		// Calculate available space for components
		titleHeight := 3   // Space for title with margins
		inputHeight := 6   // Space for textarea with margins and borders
		borderPadding := 4 // Space for chat border and padding

		// Calculate viewport height dynamically
		viewportHeight := msg.Height - titleHeight - inputHeight - borderPadding
		if viewportHeight < 8 {
			viewportHeight = 8 // Minimum viewport height
		}

		// Calculate content width accounting for sidebar and chat border/padding
		// Sidebar + 2 chars border + 2 chars padding on each side = sidebar + 8 chars total
		contentWidth := msg.Width - m.sidebarWidth - 8
		if contentWidth < 20 {
			contentWidth = 20 // Minimum content width
		}

		// Set viewport dimensions with better scrolling
		m.viewport.Width = contentWidth
		m.viewport.Height = viewportHeight
		m.viewport.YPosition = 0 // Reset scroll position

		// Set textarea width to match viewport width
		m.textarea.SetWidth(contentWidth)

		// Update markdown renderer word wrap to match content width
		// Account for additional message styling (headers, borders, etc.)
		markdownWidth := contentWidth - 8 // Reserve space for message formatting
		if markdownWidth > 0 {
			m.markdownRenderer, _ = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(markdownWidth),
			)
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			userInput := strings.TrimSpace(m.textarea.Value())
			if userInput != "" {
				m.DisplayUserMessage(userInput)
				m.loading = true
				if m.llmService == nil {
					cmds = append(cmds, utils.CmdHandler(ui.SendEmptyResponseMsg()))
				} else {
					cmds = append(cmds, m.generateResponseCommand())
				}
			}
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m *ChatModel) View() string {
	// Create responsive title
	title := styles.ChatTitleStyle.Render("🍳 Cooking Assistant")

	// Center the title if we have window width information
	if m.windowWidth > 0 {
		titleWidth := lipgloss.Width(title)
		if titleWidth < m.windowWidth {
			padding := (m.windowWidth - titleWidth) / 2
			title = lipgloss.NewStyle().MarginLeft(padding).Render(title)
		}
	}

	sidebar := RenderSidebar(m.messageCount, m.tokenCount, m.ollamaStatus, m.llmService, m.sidebarWidth, m.viewport.Height)
	chat := styles.ChatStyle.Render(m.viewport.View())
	input := m.textarea.View()
	mainContent := lipgloss.JoinVertical(lipgloss.Left, chat, input)
	chatLayout := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
	layout := lipgloss.JoinVertical(lipgloss.Left, title, chatLayout)

	return layout
}

// SetSize sets the width and height of the model
func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.windowWidth = width
	m.windowHeight = height

	// Calculate sidebar width (30% of total width, minimum 25, maximum 40)
	// But ensure we don't take too much space on small screens
	if width < 80 {
		m.sidebarWidth = 25 // Fixed width for small screens
	} else {
		m.sidebarWidth = width * 30 / 100
		if m.sidebarWidth < 25 {
			m.sidebarWidth = 25
		} else if m.sidebarWidth > 40 {
			m.sidebarWidth = 40
		}
	}

	// Calculate available space for components (same logic as WindowSizeMsg)
	titleHeight := 3   // Space for title with margins
	inputHeight := 6   // Space for textarea with margins and borders
	borderPadding := 4 // Space for chat border and padding

	// Calculate viewport height dynamically
	viewportHeight := height - titleHeight - inputHeight - borderPadding
	if viewportHeight < 8 {
		viewportHeight = 8 // Minimum viewport height
	}

	// Calculate content width accounting for sidebar and chat border/padding
	contentWidth := width - m.sidebarWidth - 8
	if contentWidth < 20 {
		contentWidth = 20 // Minimum content width
	}

	// Update viewport and textarea sizes with consistent calculations
	m.viewport.Width = contentWidth
	m.viewport.Height = viewportHeight
	m.viewport.YPosition = 0 // Reset scroll position
	m.textarea.SetWidth(contentWidth)

	// Update markdown renderer word wrap
	markdownWidth := contentWidth - 8
	if markdownWidth > 0 {
		m.markdownRenderer, _ = glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(markdownWidth),
		)
	}
}

// GetSize returns the current width and height of the model
func (m *ChatModel) GetSize() (width, height int) {
	return m.width, m.height
}

func (m *ChatModel) GetModelState() ui.ModelState {
	return m.modelState
}
