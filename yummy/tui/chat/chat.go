package chat

import (
	"fmt"
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
	ollamaStatus     map[string]interface{}
	modelState       ui.ModelState
}

func New(cookbook *db.CookBook, keymaps config.KeyMap) *ChatModel {
	llmService, err := NewLLMService()
	if err != nil {
		log.Printf("Warning: LLM service initialization failed: %v", err)
		// Continue without LLM service - user can still use the chat interface
		// but won't get AI responses until Ollama is properly configured
	}

	windowWidth, windowHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		windowWidth = ui.DefaultViewportWidth
		windowHeight = ui.DefaultViewportHeight
	}

	markdownRenderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(windowWidth),
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
		conversation = AppendMessage(conversation, llms.ChatMessageTypeAI, "Hello! I'm your cooking assistant. How can I help you today?")
	}

	ta := textarea.New()
	ta.Placeholder = ui.TextAreaPlaceholder
	ta.Focus()

	ta.CharLimit = ui.TextAreaMaxChar
	// Calculate content width accounting for chat border and padding
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
		loading:          false,
		modelState:       ui.ModelStateLoaded,
		sidebarWidth:     30, // Default sidebar width
		messageCount:     0,
		tokenCount:       0,
		ollamaStatus:     ollamaStatus,
	}

	// Set initial content in viewport
	if len(conversation) > 1 { // More than just system message
		chatModel.updateViewportFromConversation()
	}

	return chatModel
}

func (m *ChatModel) Init() tea.Cmd {
	return nil
}

// renderSidebar renders the sidebar with usage, model info, tools, and health status
func (m *ChatModel) renderSidebar() string {
	var sidebar strings.Builder

	// Model Information
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Model"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("• %s", ui.LlamaModel)))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("• Thinking On"))
	sidebar.WriteString("\n\n")

	// Usage Statistics
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Usage"))
	sidebar.WriteString("\n")
	// Calculate rough percentage based on message count
	usagePercent := (m.messageCount * 5) % 100 // Simple calculation for demo
	sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("%d%% (%dK) $%.2f", usagePercent, m.tokenCount/1000, float64(m.tokenCount)*0.0001)))
	sidebar.WriteString("\n\n")

	// Ollama Health Status
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Ollama Status"))
	sidebar.WriteString("\n")

	status := m.ollamaStatus
	if status["functional"].(bool) && status["model_available"].(bool) {
		sidebar.WriteString(styles.SidebarSuccessStyle.Render("✅ Service Healthy"))
	} else {
		sidebar.WriteString(styles.SidebarErrorStyle.Render("❌ Service Issues"))
		if errors, ok := status["errors"].([]string); ok && len(errors) > 0 {
			for _, err := range errors {
				sidebar.WriteString("\n")
				sidebar.WriteString(styles.SidebarErrorStyle.Render(fmt.Sprintf("  • %s", err)))
			}
		}
	}
	sidebar.WriteString("\n\n")

	// Available Tools
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Available Tools"))
	sidebar.WriteString("\n")
	if m.llmService != nil && m.llmService.toolManager != nil {
		tools := m.llmService.toolManager.GetTools()
		for _, tool := range tools {
			sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("• %s", tool.Name())))
			sidebar.WriteString("\n")
		}
	} else {
		sidebar.WriteString(styles.SidebarContentStyle.Render("• No tools available"))
	}
	sidebar.WriteString("\n\n")

	// Session Stats
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Session Stats"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("• Messages: %d", m.messageCount)))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("• Tokens: %d", m.tokenCount)))
	sidebar.WriteString("\n\n")

	// Controls
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Controls"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("• Enter: Send message"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("• Ctrl+C: Exit"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("• ↑/↓: Scroll messages"))

	// Create a dynamic sidebar style based on the current width
	sidebarStyle := styles.SidebarStyle.Width(m.sidebarWidth - 4)
	return sidebarStyle.Render(sidebar.String())
}

func (m *ChatModel) renderConversationAsMarkdown() string {
	var content strings.Builder

	log.Printf("renderConversationAsMarkdown: conversation length = %d", len(m.conversation))

	for i, message := range m.conversation {
		// Skip the initial system prompt
		if i == 0 && message.Role == llms.ChatMessageTypeSystem {
			log.Printf("Skipping system message at index %d", i)
			continue
		}

		// Get the text content from the message
		var messageText string
		for _, part := range message.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				messageText += textPart.Text
			}
		}

		log.Printf("Rendering message %d: role=%s, text=%s", i, message.Role, messageText)

		// Format based on message role with simple text
		switch message.Role {
		case llms.ChatMessageTypeHuman:
			// User message - simple text
			userHeader := styles.UserMessageStyle.Render("👤 You:")
			userContent := styles.UserContentStyle.Render(messageText)
			content.WriteString(userHeader + "\n" + userContent)
		case llms.ChatMessageTypeAI:
			// Assistant message - simple text
			assistantHeader := styles.AssistantMessageStyle.Render("🤖 Assistant:")
			assistantContent := styles.AssistantContentStyle.Render(messageText)
			content.WriteString(assistantHeader + "\n" + assistantContent)
		}

		// Add spacing between messages for better readability
		if i < len(m.conversation)-1 {
			content.WriteString("\n\n")
		}
	}

	result := content.String()
	log.Printf("renderConversationAsMarkdown: result length = %d", len(result))
	return result
}

// updateViewportFromConversation renders the conversation and updates the viewport
func (m *ChatModel) updateViewportFromConversation() {
	content := m.renderConversationAsMarkdown()

	// Add spinner if loading with simple text
	if m.loading {
		if content != "" {
			content += "\n\n"
		}
		assistantHeader := styles.AssistantMessageStyle.Render("🤖 Assistant:")
		loadingContent := styles.AssistantContentStyle.Render(m.spinner.View() + " Thinking...")
		content += assistantHeader + "\n" + loadingContent
	}

	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

// DisplayUserMessage adds a user message to the chat display
func (m *ChatModel) DisplayUserMessage(userInput string) {
	// Add user message to conversation
	m.conversation = AppendMessage(m.conversation, llms.ChatMessageTypeHuman, userInput)

	// Update message count
	m.messageCount++
	// Rough token estimation (1 token ≈ 4 characters)
	m.tokenCount += len(userInput) / 4

	// Debug: Log the conversation
	log.Printf("DisplayUserMessage: conversation length = %d", len(m.conversation))
	log.Printf("DisplayUserMessage: user input = %s", userInput)

	// Update viewport from conversation
	m.updateViewportFromConversation()
	m.textarea.Reset()
}

// Update handles all messages and updates the model
func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
		rsCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.spinner, spCmd = m.spinner.Update(msg)

	// Don't update viewport here - let individual message handlers do it

	switch msg := msg.(type) {
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

	case ResponseMsg:
		m.loading = false
		if msg.Error != nil {
			log.Printf("Error generating response: %v", msg.Error)
			// Add the error message to the conversation so the user can see it
			m.conversation = AppendMessage(m.conversation, llms.ChatMessageTypeAI, msg.Content)
			m.messageCount++
			m.tokenCount += len(msg.Content) / 4
			m.updateViewportFromConversation()
			return m, nil
		}
		m.conversation = AppendMessage(m.conversation, llms.ChatMessageTypeAI, msg.Content)
		m.messageCount++
		m.tokenCount += len(msg.Content) / 4
		m.updateViewportFromConversation()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			userInput := strings.TrimSpace(m.textarea.Value())
			if userInput == "" {
				return m, nil
			}

			m.DisplayUserMessage(userInput)
			m.loading = true

			// create command to generate response with tools
			rsCmd = func() tea.Msg {
				if m.llmService == nil {
					return ResponseMsg{
						Content: "LLM service is not available. Please ensure Ollama is installed and the required model is downloaded. See the error message above for setup instructions.",
						Error:   fmt.Errorf("llm service not initialized"),
					}
				}
				return m.llmService.GenerateResponse(m.conversation)
			}
		}
	}

	m.updateViewportFromConversation()

	return m, tea.Batch(tiCmd, vpCmd, spCmd, rsCmd)
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

	// Render sidebar
	sidebar := m.renderSidebar()

	// Render chat viewport with proper styling
	chat := styles.ChatStyle.Render(m.viewport.View())

	// Render input with consistent spacing
	input := m.textarea.View()

	// Create the main content area (chat + input)
	mainContent := lipgloss.JoinVertical(lipgloss.Left, chat, input)

	// Create the layout with sidebar and main content
	layout := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)

	// Return with title and layout
	return fmt.Sprintf("\n%s\n\n%s", title, layout)
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
