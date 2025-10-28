package themes

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a complete theme with all styling components
type Theme struct {
	Name string

	// Core styles
	Title         lipgloss.Style
	Info          lipgloss.Style
	Error         lipgloss.Style
	Header        lipgloss.Style
	Ingredient    lipgloss.Style
	Doc           lipgloss.Style
	DetailContent lipgloss.Style
	DetailHeader  lipgloss.Style
	DetailFooter  lipgloss.Style
	ScrollBar     lipgloss.Style
	Loading       lipgloss.Style
	Instruction   lipgloss.Style

	// Status line styles
	Warning             lipgloss.Style
	Success             lipgloss.Style
	Help                lipgloss.Style
	StatusLine          lipgloss.Style
	StatusLineLeft      lipgloss.Style
	StatusLineRight     lipgloss.Style
	StatusLineMode      lipgloss.Style
	StatusLineFile      lipgloss.Style
	StatusLineInfo      lipgloss.Style
	StatusLineSeparator lipgloss.Style

	// List styles
	ListStyles     list.Styles
	DelegateStyles list.DefaultItemStyles

	// Chat styles
	ChatTitle        lipgloss.Style
	Chat             lipgloss.Style
	Sidebar          lipgloss.Style
	SidebarHeader    lipgloss.Style
	SidebarSection   lipgloss.Style
	SidebarContent   lipgloss.Style
	SidebarSuccess   lipgloss.Style
	SidebarError     lipgloss.Style
	UserMessage      lipgloss.Style
	UserContent      lipgloss.Style
	AssistantMessage lipgloss.Style
	AssistantContent lipgloss.Style
	User             lipgloss.Style
	Assistant        lipgloss.Style
	Spinner          lipgloss.Style

	// Main menu styles
	MainMenuBorder          lipgloss.Style
	MainMenuContainer       lipgloss.Style
	MainMenuSeparator       lipgloss.Style
	MainMenuWelcome         lipgloss.Style
	MainMenuLogo            lipgloss.Style
	MainMenuSubtitle        lipgloss.Style
	MainMenuTitleBorder     lipgloss.Style
	MainMenuSelectedArrow   lipgloss.Style
	MainMenuSelectedItem    lipgloss.Style
	MainMenuUnselectedItem  lipgloss.Style
	MainMenuSelectedIcon    lipgloss.Style
	MainMenuUnselectedIcon  lipgloss.Style
	MainMenuSelectedTitle   lipgloss.Style
	MainMenuUnselectedTitle lipgloss.Style
	MainMenuSelectedDesc    lipgloss.Style
	MainMenuUnselectedDesc  lipgloss.Style
	MainMenuHelpHeader      lipgloss.Style
	MainMenuHelpContent     lipgloss.Style
	MainMenuHelpBorder      lipgloss.Style
	MainMenuSpinner         lipgloss.Style

	// State selector styles
	StateSelectorContainer         lipgloss.Style
	StateSelectorDialog            lipgloss.Style
	StateSelectorTitle             lipgloss.Style
	StateSelectorHelp              lipgloss.Style
	StateSelectorItem              lipgloss.Style
	StateSelectorSelectedItem      lipgloss.Style
	StateSelectorIndicator         lipgloss.Style
	StateSelectorSelectedIndicator lipgloss.Style

	// Session selector styles
	SessionSelectorTitle      lipgloss.Style
	SessionSelectorPagination lipgloss.Style
	SessionSelectorHelp       lipgloss.Style
}

// Helper methods for backward compatibility and convenience
func (t *Theme) GetMainMenuBorderTop(width int) string {
	topBorder := strings.Repeat("═", width)
	return t.MainMenuBorder.Render("╔" + topBorder + "╗")
}

func (t *Theme) GetMainMenuBorderBottom(width int) string {
	bottomBorder := strings.Repeat("═", width)
	return t.MainMenuBorder.Render("╚" + bottomBorder + "╝")
}

func (t *Theme) GetMainMenuSeparator(width int) string {
	separator := strings.Repeat("─", width)
	return t.MainMenuSeparator.Render("├" + separator + "┤")
}

func (t *Theme) RenderMainMenuItem(icon, title, description string, isSelected bool, width int) string {
	var content strings.Builder

	// Icon styling
	var iconStyle lipgloss.Style
	if isSelected {
		iconStyle = t.MainMenuSelectedIcon
	} else {
		iconStyle = t.MainMenuUnselectedIcon
	}
	content.WriteString(iconStyle.Render(icon))
	content.WriteString(" ")

	// Title styling
	var titleStyle lipgloss.Style
	if isSelected {
		titleStyle = t.MainMenuSelectedTitle
	} else {
		titleStyle = t.MainMenuUnselectedTitle
	}
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n")

	// Description styling
	var descStyle lipgloss.Style
	if isSelected {
		descStyle = t.MainMenuSelectedDesc
	} else {
		descStyle = t.MainMenuUnselectedDesc
	}
	content.WriteString(descStyle.Render(description))

	// Apply border and width styling
	var itemStyle lipgloss.Style
	if isSelected {
		itemStyle = t.MainMenuSelectedItem.Width(width).Align(lipgloss.Center)
		arrow := t.MainMenuSelectedArrow.Render("▶")
		contentStr := arrow + " " + content.String()
		return itemStyle.Render(contentStr)
	} else {
		itemStyle = t.MainMenuUnselectedItem.Width(width).Align(lipgloss.Center)
		contentStr := "  " + content.String()
		return itemStyle.Render(contentStr)
	}
}

func (t *Theme) RenderMainMenuHelp(upKey, downKey, enterKey, quitKey string) string {
	helpText := "🎮 Navigation Controls"
	helpHeader := t.MainMenuHelpHeader.Render(helpText)

	helpContent := t.MainMenuHelpContent.Render(upKey) + " • " +
		t.MainMenuHelpContent.Render(downKey) + " • " +
		t.MainMenuHelpContent.Render(enterKey) + " • " +
		t.MainMenuHelpContent.Render(quitKey)

	helpText = helpHeader + "\n" + helpContent
	return t.MainMenuHelpBorder.Render(helpText)
}

func (t *Theme) GetStateItemStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return t.StateSelectorSelectedItem
	}
	return t.StateSelectorItem
}

func (t *Theme) GetIndicatorStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return t.StateSelectorSelectedIndicator
	}
	return t.StateSelectorIndicator
}
