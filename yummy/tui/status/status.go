package status

import (
	"fmt"
	"strings"

	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
	utils "github.com/GarroshIcecream/yummy/yummy/tui/utils"
	"github.com/charmbracelet/lipgloss"
)

type StatusLine struct {
	width  int
	height int
}

// this sucks as we need to think of some other fields that are applicable to us
type StatusInfo struct {
	Mode        utils.StatusMode
	FileName    utils.StateNames
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
	spaceWidth := s.width - leftWidth - rightWidth - utils.StatusLinePadding

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

func CreateStatusInfo(sessionState utils.SessionState, additionalInfo map[string]interface{}) StatusInfo {
	info := StatusInfo{}

	switch sessionState {
	case utils.SessionStateMainMenu:
		info.Mode = utils.StatusModeMenu
		info.FileName = utils.StateNameMainMenu
		info.FileInfo = "Ready"

	case utils.SessionStateList:
		info.Mode = utils.StatusModeList
		info.FileName = utils.StateNameList
		if selectedItem, ok := additionalInfo["selected_item"].(string); ok {
			info.FileName = utils.StateNames(selectedItem)
		} else {
			info.FileName = utils.StateNameList
		}
		if count, ok := additionalInfo["count"].(int); ok {
			info.FileInfo = fmt.Sprintf("%d recipes", count)
		} else {
			info.FileInfo = "Loading..."
		}

	case utils.SessionStateDetail:
		info.Mode = utils.StatusModeRecipe
		if recipeName, ok := additionalInfo["recipe_name"].(string); ok {
			info.FileName = utils.StateNames(recipeName)
		} else {
			info.FileName = utils.StateNameDetail
		}
		if scrollPos, ok := additionalInfo["scroll_pos"].(int); ok {
			if totalLines, ok := additionalInfo["total_lines"].(int); ok {
				info.Position = fmt.Sprintf("Line %d", scrollPos+1)
				info.CurrentLine = scrollPos + 1
				info.LineCount = totalLines
				info.FileInfo = fmt.Sprintf("%d lines", totalLines)
			}
		}

	case utils.SessionStateEdit:
		info.Mode = utils.StatusModeEdit
		if recipeName, ok := additionalInfo["recipe_name"].(string); ok {
			info.FileName = utils.StateNames(recipeName)
		} else {
			info.FileName = utils.StateNameEdit
		}
		info.Modified = true
		info.FileInfo = "Modified"

	case utils.SessionStateChat:
		info.Mode = utils.StatusModeChat
		info.FileName = utils.StateNameChat
		info.FileInfo = "Chat Mode"

	case utils.SessionStateStateSelector:
		info.Mode = utils.StatusModeStateSelector
		if stateSelected, ok := additionalInfo["state_selected"].(string); ok {
			info.FileName = utils.StateNames(stateSelected)
		} else {
			info.FileName = utils.StateNameMainMenu
		}
		info.FileInfo = "State Selector"

	case utils.SessionStateSessionSelector:
		info.Mode = utils.StatusModeSessionSelector
		info.FileName = "Session Selector"
		info.FileInfo = "Select Chat Session"
	}

	return info
}
