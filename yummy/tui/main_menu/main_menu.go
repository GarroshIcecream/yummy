package main_menu

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/GarroshIcecream/yummy/yummy/config"
	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	"github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuModel struct {
	// Configuration
	cookbook   *db.CookBook
	theme      *themes.Theme
	keyMap     config.KeyMap
	config     *config.MainMenuConfig
	modelState common.ModelState

	// UI components
	items    []menuItem
	selected int
	width    int
	height   int

	// Spinner
	spinner    spinner.Model
	loadingMsg string
}

type menuItem struct {
	title       string
	description string
	state       common.SessionState
	handler     func() tea.Cmd
	icon        string
}

func New(cookbook *db.CookBook, keymaps config.KeyMap, theme *themes.Theme, config *config.MainMenuConfig) *MainMenuModel {
	items := []menuItem{
		{
			title:       "Browse Your Cookbook",
			description: "Explore your personal collection of saved recipes",
			state:       common.SessionStateList,
			icon:        "📚",
			handler:     nil,
		},
		{
			title:       "Discover Random Recipe",
			description: "Get inspired with a surprise recipe from the web",
			state:       common.SessionStateDetail,
			icon:        "🎲",
			handler:     func() tea.Cmd { return RandomRecipeCmd(cookbook) },
		},
		{
			title:       "AI Cooking Assistant",
			description: "Chat with our AI for cooking tips and recipe advice",
			state:       common.SessionStateChat,
			icon:        "🤖",
			handler:     nil,
		},
	}

	// Initialize spinner with a nice style
	spinnerModel := spinner.New()
	spinnerModel.Spinner = spinner.Dot
	spinnerModel.Style = theme.Spinner

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
		spinner:    spinnerModel,
		modelState: common.ModelStateLoading,
		loadingMsg: loadingMsg,
		theme:      theme,
		config:     config,
	}
}

func (m *MainMenuModel) Init() tea.Cmd {
	return nil
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
				cmds = append(cmds, messages.SendSessionStateMsg(selectedItem.state))
				if selectedItem.handler != nil {
					cmds = append(cmds, selectedItem.handler())
				}
			}
		}
	}

	if m.width > 0 && m.height > 0 {
		m.modelState = common.ModelStateLoaded
	}

	return m, tea.Sequence(cmds...)
}

func (m *MainMenuModel) View() string {
	if m.modelState == common.ModelStateLoading {
		return m.spinner.View() + " " + m.loadingMsg
	}

	var content strings.Builder

	content.WriteString(m.theme.GetMainMenuBorderTop(m.config.MainMenuContentWidth))
	content.WriteString("\n")

	// Title
	title := consts.MainMenuLogoText
	content.WriteString(title)
	content.WriteString("\n\n")

	// Add decorative separator
	content.WriteString(m.theme.GetMainMenuSeparator(m.config.MainMenuContentWidth))
	content.WriteString("\n\n")

	// Welcome message
	welcomeMsg := m.config.MainMenuWelcomeText
	content.WriteString(welcomeMsg)
	content.WriteString("\n\n")

	// Menu items
	menuContent := m.renderMenuItems()
	content.WriteString(menuContent)

	// Add decorative bottom border
	content.WriteString("\n")
	content.WriteString(m.theme.GetMainMenuBorderBottom(m.config.MainMenuContentWidth))

	// Center the content with styling
	style := m.theme.MainMenuContainer.
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(content.String())
}

func (m *MainMenuModel) renderMenuItems() string {
	var items strings.Builder

	for i, item := range m.items {
		isSelected := i == m.selected
		itemContentStyled := m.theme.RenderMainMenuItem(item.icon, item.title, item.description, isSelected, m.config.MenuItemWidth)

		items.WriteString(itemContentStyled)

		if i < len(m.items)-1 {
			items.WriteString("\n\n")
		} else {
			items.WriteString("\n")
		}
	}

	return items.String()
}

func RandomRecipeCmd(cookbook *db.CookBook) tea.Cmd {
	recipe, err := cookbook.RandomRecipe()
	if err == nil {
		return messages.SendRecipeSelectedMsg(recipe.ID)
	}
	return nil
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

func (m *MainMenuModel) GetModelState() common.ModelState {
	return m.modelState
}

func (m *MainMenuModel) GetSessionState() common.SessionState {
	return common.SessionStateMainMenu
}
