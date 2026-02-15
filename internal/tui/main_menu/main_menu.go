package main_menu

import (
	"fmt"
	"strings"

	"github.com/GarroshIcecream/yummy/internal/config"
	consts "github.com/GarroshIcecream/yummy/internal/consts"
	db "github.com/GarroshIcecream/yummy/internal/db"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	"github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/tui/dialog"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuModel struct {
	// Configuration
	cookbook   *db.CookBook
	theme      *themes.Theme
	keyMap     config.MainMenuKeyMap
	config     config.MainMenuConfig
	modelState common.ModelState

	// UI components
	items    []menuItem
	selected int
	width    int
	height   int

	// Spinner
	spinner spinner.Model
}

type menuItem struct {
	title       string
	description string
	state       common.SessionState
	handler     func() tea.Cmd
	action      string // command palette action (for modal-based items)
}

func NewMainMenuModel(cookbook *db.CookBook, theme *themes.Theme) (*MainMenuModel, error) {
	cfg := config.GetGlobalConfig()
	if cfg == nil {
		return nil, fmt.Errorf("global config not set")
	}

	mainMenuConfig := cfg.MainMenu
	keymaps := cfg.Keymap.ToKeyMap().GetMainMenuKeyMap()
	items := []menuItem{
		{
			title:       "Browse Cookbook",
			description: "Explore your personal collection of saved recipes",
			state:       common.SessionStateList,
		},
		{
			title:       "Find Recipe",
			description: "Search and jump to a specific recipe by name",
			action:      dialog.ActionRecipeSelector,
		},
		{
			title:       "Add Recipe",
			description: "Import a new recipe from any URL",
			action:      dialog.ActionAddRecipe,
		},
		{
			title:       "Random Recipe",
			description: "Get inspired with a surprise pick from the web",
			state:       common.SessionStateDetail,
			handler:     func() tea.Cmd { return RandomRecipeCmd(cookbook) },
		},
		{
			title:       "AI Assistant",
			description: "Chat with AI for cooking tips and recipe ideas",
			state:       common.SessionStateChat,
		},
	}

	spinnerModel := spinner.New()
	spinnerModel.Spinner = spinner.Dot
	spinnerModel.Style = theme.Spinner

	return &MainMenuModel{
		cookbook:   cookbook,
		items:      items,
		selected:   0,
		keyMap:     keymaps,
		spinner:    spinnerModel,
		modelState: common.ModelStateLoading,
		theme:      theme,
		config:     mainMenuConfig,
	}, nil
}

func (m *MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m *MainMenuModel) Update(msg tea.Msg) (common.TUIModel, tea.Cmd) {
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
			} else {
				m.selected = len(m.items) - 1
			}

		case key.Matches(msg, m.keyMap.CursorDown):
			if m.selected < len(m.items)-1 {
				m.selected++
			} else {
				m.selected = 0
			}

		case key.Matches(msg, m.keyMap.Enter):
			if len(m.items) > 0 {
				selectedItem := m.items[m.selected]

				// Modal-based items use command palette action flow
				if selectedItem.action != "" {
					cmds = append(cmds, messages.SendCommandPaletteActionMsg(selectedItem.action))
					return m, tea.Batch(cmds...)
				}

				// State-based items navigate directly
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
		return m.spinner.View()
	}

	var content strings.Builder

	// Logo
	logo := m.theme.MainMenuLogo.Render(consts.MainMenuLogoText)
	content.WriteString(logo)
	content.WriteString("\n")

	// Subtitle
	subtitle := m.theme.MainMenuWelcome.Render("Your personal recipe manager")
	content.WriteString(subtitle)
	content.WriteString("\n\n")

	// Thin separator
	sepWidth := m.config.MainMenuContentWidth
	sep := m.theme.MainMenuSeparator.Render(strings.Repeat("─", sepWidth))
	content.WriteString(sep)
	content.WriteString("\n\n")

	// Menu items
	content.WriteString(m.renderMenuItems())

	// Bottom separator
	content.WriteString("\n\n")
	content.WriteString(sep)
	content.WriteString("\n\n")

	// Help line
	helpKeys := m.theme.MainMenuHelpKey
	helpDesc := m.theme.MainMenuHelpDesc
	enterHelp := m.keyMap.Enter.Help().Key
	quitHelp := m.keyMap.Quit.Help().Key
	help := helpKeys.Render("↑↓") + helpDesc.Render(" navigate  ") +
		helpKeys.Render(enterHelp) + helpDesc.Render(" select  ") +
		helpKeys.Render(quitHelp) + helpDesc.Render(" quit")
	content.WriteString(help)

	// Center everything
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
		itemStr := m.theme.RenderMainMenuItem(item.title, item.description, isSelected, m.config.MenuItemWidth)
		items.WriteString(itemStr)

		if i < len(m.items)-1 {
			items.WriteString("\n\n")
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

func (m *MainMenuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *MainMenuModel) GetSize() (width, height int) {
	return m.width, m.height
}

func (m *MainMenuModel) GetModelState() common.ModelState {
	return m.modelState
}

func (m *MainMenuModel) GetSessionState() common.SessionState {
	return common.SessionStateMainMenu
}

func (m *MainMenuModel) GetCurrentTheme() *themes.Theme {
	return m.theme
}

func (m *MainMenuModel) SetTheme(theme *themes.Theme) {
	m.theme = theme
}
