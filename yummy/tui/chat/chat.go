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

	db "github.com/GarroshIcecream/yummy/yummy/db"
	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
)

type ChatModel struct {
	cookbook         *db.CookBook
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
}

func New(cookbook *db.CookBook) *ChatModel {
	llmService, err := NewLLMService()
	if err != nil {
		log.Fatal(err)
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

	conversation := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(
				llmService.GetSystemPrompt(),
			)},
		},
	}

	ta := textarea.New()
	ta.Placeholder = ui.TextAreaPlaceholder
	ta.Focus()

	ta.CharLimit = ui.TextAreaMaxChar
	ta.SetWidth(windowWidth)
	ta.SetHeight(ui.TextAreaHeight)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(windowWidth, windowHeight)
	vp.Style = styles.ChatStyle

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	return &ChatModel{
		cookbook:         cookbook,
		textarea:         ta,
		viewport:         vp,
		spinner:          s,
		conversation:     conversation,
		llmService:       llmService,
		markdownRenderer: markdownRenderer,
		loading:          false,
	}
}

func (m *ChatModel) Init() tea.Cmd {
	return nil
}

func (m *ChatModel) renderConversationAsMarkdown() string {
	var markdown strings.Builder

	for i, message := range m.conversation {
		// Skip the initial system prompt
		if i == 0 && message.Role == llms.ChatMessageTypeSystem {
			continue
		}

		// Add spacing between messages
		if i > 0 {
			markdown.WriteString("\n\n")
		}

		// Get the text content from the message
		var content string
		for _, part := range message.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				content += textPart.Text
			}
		}

		// Format based on message role
		switch message.Role {
		case llms.ChatMessageTypeHuman:
			markdown.WriteString("## 👤 You\n\n" + content)
		case llms.ChatMessageTypeAI:
			markdown.WriteString("## 🤖 Assistant\n\n" + content)
		}
	}

	renderedContent, err := m.markdownRenderer.Render(markdown.String())
	if err != nil {
		log.Printf("Error rendering conversation markdown: %v", err)
		return markdown.String()
	}

	return renderedContent
}

// updateViewportFromConversation renders the conversation and updates the viewport
func (m *ChatModel) updateViewportFromConversation() {
	markdownContent := m.renderConversationAsMarkdown()

	// Add spinner if loading
	if m.loading {
		if markdownContent != "" {
			markdownContent += "\n\n"
		}
		markdownContent += "🤖 Assistant\n\n" + m.spinner.View() + " Thinking..."
	}

	m.viewport.SetContent(markdownContent)
	m.viewport.GotoBottom()
}

// DisplayUserMessage adds a user message to the chat display
func (m *ChatModel) DisplayUserMessage(userInput string) {
	// Add user message to conversation
	m.conversation = AppendMessage(m.conversation, llms.ChatMessageTypeHuman, userInput)

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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Store window dimensions
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Calculate available space for components
		titleHeight := 3   // Space for title with margins
		inputHeight := 5   // Space for textarea with margins
		borderPadding := 4 // Space for borders and padding

		// Calculate viewport height dynamically
		viewportHeight := msg.Height - titleHeight - inputHeight - borderPadding
		if viewportHeight < 5 {
			viewportHeight = 5 // Minimum viewport height
		}

		// Set viewport dimensions
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = viewportHeight

		// Set textarea width
		m.textarea.SetWidth(msg.Width - 4)

		// Update markdown renderer word wrap to match viewport width
		// Account for borders, padding, and message styling (about 12 chars total)
		contentWidth := msg.Width - 12
		if contentWidth > 0 {
			m.markdownRenderer, _ = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(contentWidth),
			)
		}

	case ResponseMsg:
		m.loading = false
		if msg.Error != nil {
			log.Printf("Error generating response: %v", msg.Error)
			return m, nil
		}
		m.conversation = AppendMessage(m.conversation, llms.ChatMessageTypeAI, msg.Content)
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

	chat := styles.ChatStyle.Render(m.viewport.View())
	input := m.textarea.View()

	return fmt.Sprintf("\n%s\n\n%s\n%s", title, chat, input)
}

// SetSize sets the width and height of the model
func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.windowWidth = width
	m.windowHeight = height
	
	// Update viewport and textarea sizes
	m.viewport.Width = width
	m.viewport.Height = height - 4 // Reserve space for input
	
	m.textarea.SetWidth(width)
}

// GetSize returns the current width and height of the model
func (m *ChatModel) GetSize() (width, height int) {
	return m.width, m.height
}
