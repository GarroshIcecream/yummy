package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Main Menu Styles - Purple theme with golden accents

// Border and Layout Styles
var (
	MainMenuBorderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9370DB"))

	MainMenuContainerStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1A0B2E")).
				Padding(1, 2)

	MainMenuSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9370DB"))

	MainMenuWelcomeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#DDA0DD")).
				Italic(true).
				Padding(1, 0)
)

// Logo and Title Styles
var (
	MainMenuLogoStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#B19CD9")).
				Bold(true)

	MainMenuSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#DDA0DD")).
				Italic(true).
				Padding(1, 0)

	MainMenuTitleBorderStyle = lipgloss.NewStyle().
					Border(lipgloss.DoubleBorder()).
					BorderForeground(lipgloss.Color("#9370DB")).
					Align(lipgloss.Center).
					Padding(1, 2).
					Margin(1, 0)
)

// Menu Item Styles
var (
	MainMenuSelectedArrowStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFD700")).
					Bold(true)

	MainMenuSelectedItemStyle = lipgloss.NewStyle().
					Border(lipgloss.DoubleBorder()).
					BorderForeground(lipgloss.Color("#FFD700"))

	MainMenuUnselectedItemStyle = lipgloss.NewStyle().
					Border(lipgloss.NormalBorder()).
					BorderForeground(lipgloss.Color("#9370DB"))

	// Icon and Text Styles for Menu Items
	MainMenuSelectedIconStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFD700")).
					Bold(true)

	MainMenuUnselectedIconStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#DDA0DD"))

	MainMenuSelectedTitleStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FFD700"))

	MainMenuUnselectedTitleStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#E6E6FA"))

	MainMenuSelectedDescStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFA500")).
					Italic(true).
					PaddingLeft(4).
					PaddingBottom(1)

	MainMenuUnselectedDescStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#B19CD9")).
					Italic(true).
					PaddingLeft(4).
					PaddingBottom(1)
)

// Help Section Styles
var (
	MainMenuHelpHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9370DB")).
				Bold(true).
				PaddingBottom(1)

	MainMenuHelpContentStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#DDA0DD")).
					PaddingLeft(2)

	MainMenuHelpBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("#9370DB")).
				Background(lipgloss.Color("#2D1B3D")).
				Padding(1, 2).
				Margin(1, 0)
)

// Spinner Style
var MainMenuSpinnerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#9370DB"))

// Helper functions for main menu rendering

// GetBorderTop creates the top border for the main menu
func GetMainMenuBorderTop(width int) string {
	topBorder := strings.Repeat("═", width)
	return MainMenuBorderStyle.Render("╔" + topBorder + "╗")
}

// GetBorderBottom creates the bottom border for the main menu
func GetMainMenuBorderBottom(width int) string {
	topBorder := strings.Repeat("═", width)
	return MainMenuBorderStyle.Render("╚" + topBorder + "╝")
}

// GetSeparator creates a separator line for the main menu
func GetMainMenuSeparator(width int) string {
	separator := strings.Repeat("─", width)
	return MainMenuSeparatorStyle.Render("├" + separator + "┤")
}

// GetMainMenuLogo returns the styled ASCII logo
func GetMainMenuLogo() string {
	logo := `
    ██╗   ██╗██╗   ██╗███╗   ███╗███╗   ███╗██╗   ██╗
    ╚██╗ ██╔╝██║   ██║████╗ ████║████╗ ████║╚██╗ ██╔╝
     ╚████╔╝ ██║   ██║██╔████╔██║██╔████╔██║ ╚████╔╝ 
      ╚██╔╝  ██║   ██║██║╚██╔╝██║██║╚██╔╝██║  ╚██╔╝  
       ██║   ╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║   ██║   
       ╚═╝    ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝   ╚═╝`
	return MainMenuLogoStyle.Render(logo)
}

// GetMainMenuSubtitle returns the styled subtitle
func GetMainMenuSubtitle() string {
	subtitle := "🍳 Your Personal Culinary Companion 🍳"
	return MainMenuSubtitleStyle.Render(subtitle)
}

// GetMainMenuTitle returns the complete styled title with logo and subtitle
func GetMainMenuTitle() string {
	logo := GetMainMenuLogo()
	subtitle := GetMainMenuSubtitle()
	content := logo + "\n" + subtitle
	return MainMenuTitleBorderStyle.Render(content)
}

// GetMainMenuWelcomeMessage returns the styled welcome message
func GetMainMenuWelcomeMessage() string {
	welcomeMsg := "🌟 Welcome to your culinary journey! Choose an option below to get started:"
	return MainMenuWelcomeStyle.Render(welcomeMsg)
}

// RenderMenuItem renders a single menu item with appropriate styling
func RenderMainMenuItem(icon, title, description string, isSelected bool, width int) string {
	var content strings.Builder

	// Icon styling
	var iconStyle lipgloss.Style
	if isSelected {
		iconStyle = MainMenuSelectedIconStyle
	} else {
		iconStyle = MainMenuUnselectedIconStyle
	}
	content.WriteString(iconStyle.Render(icon))
	content.WriteString(" ")

	// Title styling
	var titleStyle lipgloss.Style
	if isSelected {
		titleStyle = MainMenuSelectedTitleStyle
	} else {
		titleStyle = MainMenuUnselectedTitleStyle
	}
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n")

	// Description styling
	var descStyle lipgloss.Style
	if isSelected {
		descStyle = MainMenuSelectedDescStyle
	} else {
		descStyle = MainMenuUnselectedDescStyle
	}
	content.WriteString(descStyle.Render(description))

	// Apply border and width styling
	var itemStyle lipgloss.Style
	if isSelected {
		itemStyle = MainMenuSelectedItemStyle.Width(width).Align(lipgloss.Center)
		arrow := MainMenuSelectedArrowStyle.Render("▶")
		contentStr := arrow + " " + content.String()
		return itemStyle.Render(contentStr)
	} else {
		itemStyle = MainMenuUnselectedItemStyle.Width(width).Align(lipgloss.Center)
		contentStr := "  " + content.String()
		return itemStyle.Render(contentStr)
	}
}

// RenderMainMenuHelp renders the help section
func RenderMainMenuHelp(upKey, downKey, enterKey, quitKey string) string {
	helpText := "🎮 Navigation Controls"
	helpHeader := MainMenuHelpHeaderStyle.Render(helpText)

	helpContent := MainMenuHelpContentStyle.Render(upKey) + " • " +
		MainMenuHelpContentStyle.Render(downKey) + " • " +
		MainMenuHelpContentStyle.Render(enterKey) + " • " +
		MainMenuHelpContentStyle.Render(quitKey)

	helpText = helpHeader + "\n" + helpContent
	return MainMenuHelpBorderStyle.Render(helpText)
}
