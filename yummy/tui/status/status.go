package status

import (
	"fmt"
	"strings"

	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/charmbracelet/lipgloss"
)

type StatusLine struct {
	width  int
	height int
}

// this sucks as we need to think of some other fields that are applicable to us
type StatusInfo struct {
	Mode        ui.StatusMode
	FileName    ui.StateNames
	FileInfo    string
	Position    string
	LineCount   int
	CurrentLine int
	Modified    bool
	ReadOnly    bool
}

func New(width, height int) *StatusLine {
	return &StatusLine{
		width:  width,
		height: height,
	}
}

func (s *StatusLine) SetSize(width, height int) {
	s.width = width
	s.height = height
}

func (s *StatusLine) Render(info StatusInfo) string {
	if s.width <= 0 {
		return ""
	}

	leftContent := s.renderLeftSide(info)
	rightContent := s.renderRightSide(info)

	leftStyled := styles.StatusLineLeftStyle.Render(leftContent)
	rightStyled := styles.StatusLineRightStyle.Render(rightContent)

	leftWidth := lipgloss.Width(leftStyled)
	rightWidth := lipgloss.Width(rightStyled)
	spaceWidth := s.width - leftWidth - rightWidth - ui.StatusLinePadding

	emptySpace := ""
	if spaceWidth > 0 {
		emptySpace = styles.StatusLineStyle.Width(spaceWidth).Render(strings.Repeat(" ", spaceWidth))
	}

	statusLine := leftStyled + emptySpace + rightStyled
	return styles.StatusLineStyle.Render(statusLine)
}

func (s *StatusLine) renderLeftSide(info StatusInfo) string {
	var parts []string

	if info.Mode != "" {
		modeText := styles.StatusLineModeStyle.Render(string(info.Mode))
		parts = append(parts, modeText)
	}

	// File name and status indicators
	if info.FileName != "" {
		fileName := info.FileName
		if info.Modified {
			fileName += " +"
		}
		if info.ReadOnly {
			fileName += " [RO]"
		}
		fileText := styles.StatusLineFileStyle.Render(string(fileName))
		parts = append(parts, fileText)
	}

	return strings.Join(parts, " ")
}

func (s *StatusLine) renderRightSide(info StatusInfo) string {
	var parts []string

	if info.Position != "" {
		positionText := styles.StatusLineInfoStyle.Render(info.Position)
		parts = append(parts, positionText)
	}

	if info.FileInfo != "" {
		fileInfoText := styles.StatusLineInfoStyle.Render(info.FileInfo)
		parts = append(parts, fileInfoText)
	}

	separator := styles.StatusLineSeparatorStyle.Render(" | ")
	return strings.Join(parts, separator)
}

func CreateStatusInfo(sessionState ui.SessionState, additionalInfo map[string]interface{}) StatusInfo {
	info := StatusInfo{}

	switch sessionState {
	case ui.SessionStateMainMenu:
		info.Mode = ui.StatusModeMenu
		info.FileName = ui.StateNameMainMenu
		info.FileInfo = "Ready"

	case ui.SessionStateList:
		info.Mode = ui.StatusModeList
		info.FileName = ui.StateNameList
		if selectedItem, ok := additionalInfo["selected_item"].(string); ok {
			info.FileName = ui.StateNames(selectedItem)
		} else {
			info.FileName = ui.StateNameList
		}
		if count, ok := additionalInfo["count"].(int); ok {
			info.FileInfo = fmt.Sprintf("%d recipes", count)
		} else {
			info.FileInfo = "Loading..."
		}

	case ui.SessionStateDetail:
		info.Mode = ui.StatusModeRecipe
		if recipeName, ok := additionalInfo["recipe_name"].(string); ok {
			info.FileName = ui.StateNames(recipeName)
		} else {
			info.FileName = ui.StateNameDetail
		}
		if scrollPos, ok := additionalInfo["scroll_pos"].(int); ok {
			if totalLines, ok := additionalInfo["total_lines"].(int); ok {
				info.Position = fmt.Sprintf("Line %d", scrollPos+1)
				info.CurrentLine = scrollPos + 1
				info.LineCount = totalLines
				info.FileInfo = fmt.Sprintf("%d lines", totalLines)
			}
		}

	case ui.SessionStateEdit:
		info.Mode = ui.StatusModeEdit
		if recipeName, ok := additionalInfo["recipe_name"].(string); ok {
			info.FileName = ui.StateNames(recipeName)
		} else {
			info.FileName = ui.StateNameEdit
		}
		info.Modified = true
		info.FileInfo = "Modified"

	case ui.SessionStateChat:
		info.Mode = ui.StatusModeChat
		info.FileName = ui.StateNameChat
		info.FileInfo = "Chat Mode"

	case ui.SessionStateStateSelector:
		info.Mode = ui.StatusModeStateSelector
		if stateSelected, ok := additionalInfo["state_selected"].(string); ok {
			info.FileName = ui.StateNames(stateSelected)
		} else {
			info.FileName = ui.StateNameMainMenu
		}
		info.FileInfo = "State Selector"
	}

	return info
}
