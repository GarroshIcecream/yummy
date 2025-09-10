package main_menu

import (
	"fmt"
	"strings"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuModel struct {
	cookbook    *db.CookBook
	items       []menuItem
	selected    int
	width       int
	height      int
	keyMap      KeyMap
	showHelp    bool
}

type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Help   key.Binding
	Quit   key.Binding
}

type menuItem struct {
	title       string
	description string
	state       ui.SessionState
	icon        string
}

func New(cookbook *db.CookBook) *MainMenuModel {
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

	return &MainMenuModel{
		cookbook: cookbook,
		items:    items,
		selected: 0,
		keyMap:   defaultKeyMap(),
		showHelp: false,
	}
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "move down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "h"),
			key.WithHelp("?/h", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (m *MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {

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
				
				// If the selected item is the detail state, we need to get a random recipe
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

	return m, tea.Sequence(cmds...)
}

func (m *MainMenuModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// Add decorative top border with purple theme
	topBorder := strings.Repeat("═", m.width-4)
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB")).
		Render("╔" + topBorder + "╗"))
	content.WriteString("\n")

	// Title
	title := m.renderTitle()
	content.WriteString(title)
	content.WriteString("\n\n")

	// Add decorative separator with purple theme
	separator := strings.Repeat("─", m.width-4)
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
		Background(lipgloss.Color("#2D1B3D")).
		Padding(1, 2).
		Margin(1, 0)

	return finalStyle.Render(content)
}

func (m *MainMenuModel) renderMenuItems() string {
	var items strings.Builder

	for i, item := range m.items {
		itemStyle := m.getItemStyle(i == m.selected)
		itemContent := m.renderMenuItem(item, i == m.selected)
		items.WriteString(itemStyle.Render(itemContent))
		items.WriteString("\n")
	}

	return items.String()
}

func (m *MainMenuModel) renderMenuItem(item menuItem, isSelected bool) string {
	var content strings.Builder

	// Enhanced bullet point with animation effect
	bullet := "○"
	if isSelected {
		bullet = "●"
	}

	bulletStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9370DB")).
		Bold(isSelected)

	content.WriteString(bulletStyle.Render(bullet))
	content.WriteString(" ")

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

func (m *MainMenuModel) getItemStyle(isSelected bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Padding(1, 3).
		Margin(0, 2)

	if isSelected {
		style = style.
			Background(lipgloss.Color("#4B0082")).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#9370DB")).
			Foreground(lipgloss.Color("#E6E6FA"))
	} else {
		style = style.
			Background(lipgloss.Color("#2D1B3D")).
			Foreground(lipgloss.Color("#B19CD9"))
	}

	return style
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
