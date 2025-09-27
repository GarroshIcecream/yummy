package main_menu

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/GarroshIcecream/yummy/yummy/config"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/GarroshIcecream/yummy/yummy/tui/styles"
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
	modelState ui.ModelState
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
	spinnerModel.Style = styles.MainMenuSpinnerStyle

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
		modelState: ui.ModelStateLoading,
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
		case key.Matches(msg, m.keyMap.CursorUp):
			if m.selected > 0 {
				m.selected--
			}

		case key.Matches(msg, m.keyMap.CursorDown):
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
		m.modelState = ui.ModelStateLoaded
	}

	return m, tea.Sequence(cmds...)
}

func (m *MainMenuModel) View() string {
	if m.modelState == ui.ModelStateLoading {
		return m.spinner.View() + " " + m.loadingMsg
	}

	var content strings.Builder

	content.WriteString(styles.GetMainMenuBorderTop(ui.MainMenuContentWidth))
	content.WriteString("\n")

	// Title
	title := styles.GetMainMenuTitle()
	content.WriteString(title)
	content.WriteString("\n\n")

	// Add decorative separator
	content.WriteString(styles.GetMainMenuSeparator(ui.MainMenuContentWidth))
	content.WriteString("\n\n")

	// Welcome message
	welcomeMsg := styles.GetMainMenuWelcomeMessage()
	content.WriteString(welcomeMsg)
	content.WriteString("\n\n")

	// Menu items
	menuContent := m.renderMenuItems()
	content.WriteString(menuContent)

	// Help section
	if m.showHelp {
		content.WriteString("\n")
		content.WriteString(m.renderHelp())
	}

	// Add decorative bottom border
	content.WriteString("\n")
	content.WriteString(styles.GetMainMenuBorderBottom(ui.MainMenuContentWidth))

	// Center the content with styling
	style := styles.MainMenuContainerStyle.
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(content.String())
}

func (m *MainMenuModel) renderMenuItems() string {
	var items strings.Builder

	for i, item := range m.items {
		isSelected := i == m.selected
		itemContentStyled := styles.RenderMainMenuItem(item.icon, item.title, item.description, isSelected, ui.MenuItemWidth)

		items.WriteString(itemContentStyled)

		if i < len(m.items)-1 {
			items.WriteString("\n\n")
		} else {
			items.WriteString("\n")
		}
	}

	return items.String()
}

func (m *MainMenuModel) renderHelp() string {
	return styles.RenderMainMenuHelp(
		m.keyMap.CursorUp.Help().Key,
		m.keyMap.CursorDown.Help().Key,
		m.keyMap.Enter.Help().Key,
		m.keyMap.Quit.Help().Key,
	)
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

func (m *MainMenuModel) GetModelState() ui.ModelState {
	return m.modelState
}
