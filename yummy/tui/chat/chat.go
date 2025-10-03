package chat

import (
	"fmt"
	"log"
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
	db "github.com/GarroshIcecream/yummy/yummy/db"
	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
	utils "github.com/GarroshIcecream/yummy/yummy/tui/utils"
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
	showSidebar      bool
	messageCount     int
	InputTokenCount  int
	OutputTokenCount int
	TotalTokenCount  int
	modelState       utils.ModelState
	sessionID        uint
}

func New(cookbook *db.CookBook, keymaps config.KeyMap) *ChatModel {
	llmService, err := NewLLMService(cookbook)
	if err != nil {
		log.Printf("Warning: LLM service initialization failed: %v", err)
	}

	windowWidth, windowHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		windowWidth = utils.DefaultViewportWidth
		windowHeight = utils.DefaultViewportHeight
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
	conversation = AppendMessage(conversation, llms.ChatMessageTypeAI, utils.WelcomeMessage)

	ta := textarea.New()
	ta.Placeholder = utils.TextAreaPlaceholder
	ta.Focus()

	ta.CharLimit = utils.TextAreaMaxChar
	contentWidth := windowWidth - 8
	if contentWidth < 20 {
		contentWidth = 20
	}
	ta.SetWidth(contentWidth)
	ta.SetHeight(utils.TextAreaHeight)
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

	// Create a new session for this chat
	sessionID, err := cookbook.CreateSession()
	if err != nil {
		log.Printf("Warning: Failed to create chat session: %v", err)
		sessionID = 0 // Use 0 as fallback - messages won't be saved but chat will still work
	} else {
		log.Printf("Created new chat session with ID: %d", sessionID)
	}

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
		modelState:       utils.ModelStateLoaded,
		sidebarWidth:     utils.SidebarWidth,
		showSidebar:      windowWidth >= utils.MinWidthForSidebar,
		messageCount:     0,
		InputTokenCount:  0,
		OutputTokenCount: 0,
		loading:          false,
		sessionID:        sessionID,
	}

	chatModel.updateViewportFromConversation()

	return chatModel
}

func (m *ChatModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *ChatModel) renderConversationAsMarkdown() string {
	var content strings.Builder
	for i, message := range m.conversation {
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

	// Add thinking indicator if loading
	if m.loading {
		content.WriteString("\n")
		thinkingHeader := styles.AssistantMessageStyle.Render("🤖 Assistant:")
		thinkingContent := m.spinner.View() + " Thinking..."
		content.WriteString(thinkingHeader + "\n\n")
		content.WriteString(thinkingContent)
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
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

// saveMessageToDatabase saves a message to the database
// This method handles both user messages and AI responses, storing them in the SessionMessage table
// with appropriate role, model information, and token counts for analytics
func (m *ChatModel) saveResponseToDatabase(msg utils.ResponseMsg) {
	if m.sessionID == 0 {
		log.Printf("No session ID available, skipping message save")
		return
	}

	err := m.cookbook.SaveSessionMessage(
		m.sessionID,
		msg.Response,
		llms.ChatMessageTypeAI,
		m.llmService.modelName,
		msg.Response,
		msg.PromptTokens,
		msg.CompletionTokens,
		msg.TotalTokens,
	)
	if err != nil {
		log.Printf("Failed to save message to database: %v", err)
	}
}

func (m *ChatModel) saveUserMessageToDatabase(userInput string) {
	err := m.cookbook.SaveSessionMessage(
		m.sessionID,
		userInput,
		llms.ChatMessageTypeHuman,
		m.llmService.modelName,
		userInput,
		0,
		0,
		0,
	)
	if err != nil {
		log.Printf("Failed to save user message to database: %v", err)
	}
}

// DisplayUserMessage adds a user message to the chat display
func (m *ChatModel) ProcessUserMessage(userInput string) {
	m.conversation = AppendMessage(m.conversation, llms.ChatMessageTypeHuman, userInput)
	m.updateViewportFromConversation()
	m.textarea.Reset()

	m.saveUserMessageToDatabase(userInput)
}

func (m *ChatModel) ProcessResponse(response utils.ResponseMsg) {
	m.conversation = AppendMessage(m.conversation, llms.ChatMessageTypeAI, response.Response)
	m.updateViewportFromConversation()
	m.viewport.GotoBottom()

	m.saveResponseToDatabase(response)
	m.InputTokenCount += response.PromptTokens
	m.OutputTokenCount += response.CompletionTokens
	m.TotalTokenCount += response.TotalTokens
}

// Update handles all messages and updates the model
func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case utils.GenerateResponseMsg:
		if m.llmService == nil {
			cmds = append(cmds, utils.SendEmptyResponseMsg())
		} else {
			response := m.llmService.GenerateResponse(m.conversation)
			cmds = append(cmds, utils.SendResponseMsg(response))
		}
		return m, tea.Batch(cmds...)

	case utils.ResponseMsg:
		m.loading = false
		if msg.Error != nil {
			log.Printf("Error generating response: %v", msg.Error)
		} else {
			m.messageCount++
			m.ProcessResponse(msg)
		}

	case utils.SessionSelectedMsg:
		cmds = append(cmds, m.loadSession(msg.SessionID))
		return m, tea.Batch(cmds...)

	case utils.LoadSessionMsg:
		if msg.Err != nil {
			log.Printf("Error loading session: %v", msg.Err)
		} else {
			dbMessages := make([]db.SessionMessage, len(msg.Messages))
			for i, uiMsg := range msg.Messages {
				dbMessages[i] = db.SessionMessage{
					SessionID:    uiMsg.SessionID,
					Message:      uiMsg.Message,
					Role:         uiMsg.Role,
					ModelName:    uiMsg.ModelName,
					Content:      uiMsg.Content,
					InputTokens:  uiMsg.InputTokens,
					OutputTokens: uiMsg.OutputTokens,
					TotalTokens:  uiMsg.TotalTokens,
				}
			}
			m.loadSessionMessages(dbMessages)
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.WindowSizeMsg:
		// Store window dimensions
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Determine if sidebar should be shown based on width
		m.showSidebar = msg.Width >= utils.MinWidthForSidebar

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
		var contentWidth int
		if m.showSidebar {
			contentWidth = msg.Width - m.sidebarWidth - 8
		} else {
			// No sidebar, use full width minus border/padding
			contentWidth = msg.Width - 8
		}
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
		switch {
		case key.Matches(msg, m.keyMap.Enter):
			userInput := strings.TrimSpace(m.textarea.Value())
			if userInput != "" {
				m.loading = true
				m.messageCount++
				m.ProcessUserMessage(userInput)
				cmds = append(cmds, utils.SendGenerateResponseMsg())
				cmds = append(cmds, m.spinner.Tick)
				return m, tea.Batch(cmds...)
			}
		case key.Matches(msg, m.keyMap.SessionSelector):
			cmds = append(cmds, tea.Sequence(
				utils.SendSessionStateMsg(utils.SessionStateSessionSelector),
				utils.SendLoadSessionsMsg(),
			))
			return m, tea.Batch(cmds...)
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

	chat := styles.ChatStyle.Render(m.viewport.View())
	input := m.textarea.View()
	mainContent := lipgloss.JoinVertical(lipgloss.Left, chat, input)

	var chatLayout string
	if m.showSidebar {
		sidebar := RenderSidebar(m.messageCount, m.TotalTokenCount, *m.llmService.ollamaStatus, m.llmService, m.sidebarWidth, m.viewport.Height)
		chatLayout = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainContent)
	} else {
		chatLayout = mainContent
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, chatLayout)
}

// SetSize sets the width and height of the model
func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.windowWidth = width
	m.windowHeight = height

	// Determine if sidebar should be shown based on width
	m.showSidebar = width >= utils.MinWidthForSidebar

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
	var contentWidth int
	if m.showSidebar {
		contentWidth = width - m.sidebarWidth - 8
	} else {
		// No sidebar, use full width minus border/padding
		contentWidth = width - 8
	}
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

func (m *ChatModel) GetModelState() utils.ModelState {
	return m.modelState
}

// GetSessionMessages retrieves all messages for the current session
func (m *ChatModel) GetSessionMessages() ([]db.SessionMessage, error) {
	if m.sessionID == 0 {
		return nil, fmt.Errorf("no session ID available")
	}

	sessionMessages, err := m.cookbook.GetSessionMessages(m.sessionID)
	if err != nil {
		return nil, err
	}
	return sessionMessages, nil
}

// GetSessionID returns the current session ID
func (m *ChatModel) GetSessionID() uint {
	return m.sessionID
}

// IsSessionActive returns true if the session is properly initialized
func (m *ChatModel) IsSessionActive() bool {
	return m.sessionID > 0
}

// GetDatabaseSessionStats returns statistics from the database for the current session
func (m *ChatModel) GetDatabaseSessionStats() (db.SessionStats, error) {
	if m.sessionID == 0 {
		return db.SessionStats{}, fmt.Errorf("no session ID available")
	}
	return m.cookbook.GetSessionStats(m.sessionID)
}

// loadSession loads a session asynchronously
func (m *ChatModel) loadSession(sessionID uint) tea.Cmd {
	return func() tea.Msg {
		dbMessages, err := m.cookbook.GetSessionMessages(sessionID)
		if err != nil {
			return utils.LoadSessionMsg{
				SessionID: sessionID,
				Messages:  []utils.SessionMessage{},
				Err:       err,
			}
		}

		// Convert db.SessionMessage to utils.SessionMessage
		uiMessages := make([]utils.SessionMessage, len(dbMessages))
		for i, dbMsg := range dbMessages {
			uiMessages[i] = utils.SessionMessage{
				SessionID:    dbMsg.SessionID,
				Message:      dbMsg.Message,
				Role:         dbMsg.Role,
				ModelName:    dbMsg.ModelName,
				Content:      dbMsg.Content,
				InputTokens:  dbMsg.InputTokens,
				OutputTokens: dbMsg.OutputTokens,
				TotalTokens:  dbMsg.TotalTokens,
			}
		}

		return utils.LoadSessionMsg{
			SessionID: sessionID,
			Messages:  uiMessages,
			Err:       nil,
		}
	}
}

// loadSessionMessages loads messages from a session into the conversation
func (m *ChatModel) loadSessionMessages(messages []db.SessionMessage) {
	// Clear current conversation
	m.conversation = []llms.MessageContent{}

	// Add welcome message
	m.conversation = AppendMessage(m.conversation, llms.ChatMessageTypeAI, utils.WelcomeMessage)

	// Load messages from the session
	for _, msg := range messages {
		role := llms.ChatMessageType(msg.Role)
		m.conversation = AppendMessage(m.conversation, role, msg.Content)
	}

	// Update session ID
	m.sessionID = messages[0].SessionID

	// Reset counters
	m.messageCount = len(messages)
	m.InputTokenCount = 0
	m.OutputTokenCount = 0
	m.TotalTokenCount = 0

	// Calculate token counts from loaded messages
	for _, msg := range messages {
		m.InputTokenCount += msg.InputTokens
		m.OutputTokenCount += msg.OutputTokens
		m.TotalTokenCount += msg.TotalTokens
	}

	// Update viewport with loaded conversation
	m.updateViewportFromConversation()
}
