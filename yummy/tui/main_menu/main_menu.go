package main_menu

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuModel struct {
	cookbook   *db.CookBook
	items      []menuItem
	selected   int
	width      int
	height     int
	keyMap     config.KeyMap
	showHelp   bool
	spinner    spinner.Model
	isLoading  bool
	loadingMsg string
}

type menuItem struct {
	title       string
	description string
	state       ui.SessionState
	icon        string
}

func New(cookbook *db.CookBook, keymaps config.KeyMap) *MainMenuModel {
	items := []menuItem{
		{
			title:       "Browse Your Cookbook",
			description: "Explore your personal collection of saved recipes",
			state:       ui.SessionStateList,
			icon:        "📚",
		},
		{
			title:       "Discover Random Recipe",
			description: "Get inspired with a surprise recipe from the web",
			state:       ui.SessionStateDetail,
			icon:        "🎲",
		},
		{
			title:       "AI Cooking Assistant",
			description: "Chat with our AI for cooking tips and recipe advice",
			state:       ui.SessionStateChat,
			icon:        "🤖",
		},
	}

	// Initialize spinner with a nice style
	spinnerModel := spinner.New()
	spinnerModel.Spinner = spinner.Dot
	spinnerModel.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9370DB"))

	// Pick a random loading message once
	loadingMessages := []string{
		"🍳 Cooking up something delicious...",
		"✨ Adding flavor to your experience...",
		"🌟 Preparing your culinary journey...",
		"🎯 Almost ready to serve...",
	}
	randomItem, _ := rand.Int(rand.Reader, big.NewInt(int64(len(loadingMessages))))
	loadingMsg := loadingMessages[randomItem.Int64()]

	return &MainMenuModel{
		cookbook:   cookbook,
		items:      items,
		selected:   0,
		keyMap:     keymaps,
		showHelp:   false,
		spinner:    spinnerModel,
		isLoading:  true,
		loadingMsg: loadingMsg,
	}
}

func (m *MainMenuModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			if m.selected > 0 {
				m.selected--
			}

		case key.Matches(msg, m.keyMap.Down):
			if m.selected < len(m.items)-1 {
				m.selected++
			}

		case key.Matches(msg, m.keyMap.Enter):
			if len(m.items) > 0 {
				selectedItem := m.items[m.selected]
				cmds = append(cmds, ui.SendSessionStateMsg(selectedItem.state))

				if selectedItem.state == ui.SessionStateDetail {
					recipe, err := m.cookbook.RandomRecipe()
					if err == nil {
						cmds = append(cmds, ui.SendRecipeSelectedMsg(recipe.ID))
					}
				}
			}

		case key.Matches(msg, m.keyMap.Help):
			m.showHelp = !m.showHelp
		}
	}

	if m.width > 0 && m.height > 0 {
		m.isLoading = false
	}

	return m, tea.Sequence(cmds...)
}

func (m *MainMenuModel) View() string {
	if m.isLoading {
		return m.spinner.View() + " " + m.loadingMsg
	}

	var content strings.Builder

	// Add decorative top border with purple theme
	contentWidth := 80 // Fixed width for better centering
	topBorder := strings.Repeat("═", contentWidth)
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB")).
		Render("╔" + topBorder + "╗"))
	content.WriteString("\n")

	// Title
	title := m.renderTitle()
	content.WriteString(title)
	content.WriteString("\n\n")

	// Add decorative separator with purple theme
	separator := strings.Repeat("─", contentWidth)
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB")).
		Render("├" + separator + "┤"))
	content.WriteString("\n\n")

	// Welcome message with purple theme
	welcomeMsg := "🌟 Welcome to your culinary journey! Choose an option below to get started:"
	welcomeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDA0DD")).
		Italic(true).
		Padding(1, 0)
	content.WriteString(welcomeStyle.Render(welcomeMsg))
	content.WriteString("\n\n")

	// Menu items
	menuContent := m.renderMenuItems()
	content.WriteString(menuContent)

	// Help section
	if m.showHelp {
		content.WriteString("\n")
		content.WriteString(m.renderHelp())
	}

	// Add decorative bottom border with purple theme
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB")).
		Render("╚" + topBorder + "╝"))

	// Center the content with purple gradient styling
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Background(lipgloss.Color("#1A0B2E")).
		Padding(1, 2)

	return style.Render(content.String())
}

func (m *MainMenuModel) renderTitle() string {
	// ASCII Logo
	logo := `
    ██╗   ██╗██╗   ██╗███╗   ███╗███╗   ███╗██╗   ██╗
    ╚██╗ ██╔╝██║   ██║████╗ ████║████╗ ████║╚██╗ ██╔╝
     ╚████╔╝ ██║   ██║██╔████╔██║██╔████╔██║ ╚████╔╝ 
      ╚██╔╝  ██║   ██║██║╚██╔╝██║██║╚██╔╝██║  ╚██╔╝  
       ██║   ╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║   ██║   
       ╚═╝    ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝   ╚═╝`

	// Logo styling with purple gradient effect
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B19CD9")).
		Bold(true)

	// Subtitle
	subtitle := "🍳 Your Personal Culinary Companion 🍳"
	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDA0DD")).
		Italic(true).
		Padding(1, 0)

	// Combine all elements
	content := logoStyle.Render(logo) + "\n" + subtitleStyle.Render(subtitle)

	// Add decorative border around the entire title with purple theme
	finalStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#9370DB")).
		Align(lipgloss.Center).
		Padding(1, 2).
		Margin(1, 0)

	return finalStyle.Render(content)
}

func (m *MainMenuModel) renderMenuItems() string {
	var items strings.Builder

	for i, item := range m.items {
		isSelected := i == m.selected
		itemContent := m.renderMenuItem(item, isSelected)
		if isSelected {
			itemContent = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("#9370DB")).
				Render(itemContent)
		}
		items.WriteString(itemContent)
		items.WriteString("\n")
	}

	return items.String()
}

func (m *MainMenuModel) renderMenuItem(item menuItem, isSelected bool) string {
	var content strings.Builder

	// Enhanced icon with purple glow effect
	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDA0DD")).
		Bold(isSelected)

	content.WriteString(iconStyle.Render(item.icon))
	content.WriteString(" ")

	// Enhanced title with purple typography
	titleStyle := lipgloss.NewStyle().
		Bold(isSelected).
		Foreground(lipgloss.Color("#E6E6FA"))

	content.WriteString(titleStyle.Render(item.title))
	content.WriteString("\n")

	// Enhanced description with purple theme
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B19CD9")).
		Italic(true).
		PaddingLeft(4)

	content.WriteString(descStyle.Render(item.description))

	return content.String()
}

func (m *MainMenuModel) renderHelp() string {
	// Enhanced help section with purple theme
	helpHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB")).
		Bold(true).
		PaddingBottom(1)

	helpContentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DDA0DD")).
		PaddingLeft(2)

	helpBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#9370DB")).
		Background(lipgloss.Color("#2D1B3D")).
		Padding(1, 2).
		Margin(1, 0)

	helpText := fmt.Sprintf(
		"%s\n%s • %s • %s • %s",
		helpHeaderStyle.Render("🎮 Navigation Controls"),
		helpContentStyle.Render(m.keyMap.Up.Help().Key),
		helpContentStyle.Render(m.keyMap.Down.Help().Key),
		helpContentStyle.Render(m.keyMap.Enter.Help().Key),
		helpContentStyle.Render(m.keyMap.Quit.Help().Key),
	)

	return helpBorderStyle.Render(helpText)
}

// SetSize sets the width and height of the model
func (m *MainMenuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetSize returns the current width and height of the model
func (m *MainMenuModel) GetSize() (width, height int) {
	return m.width, m.height
}
