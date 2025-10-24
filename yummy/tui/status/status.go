package status

import (
	"fmt"
	"strings"

	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	"github.com/GarroshIcecream/yummy/yummy/recipe"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/GarroshIcecream/yummy/yummy/tui/detail"
	yummy_list "github.com/GarroshIcecream/yummy/yummy/tui/list"
	"github.com/charmbracelet/lipgloss"
)

type StatusLine struct {
	width  int
	height int
	theme  themes.Theme
}

// this sucks as we need to think of some other fields that are applicable to us
type StatusInfo struct {
	Mode        consts.StatusMode
	FileName    string
	FileInfo    string
	Position    string
	LineCount   int
	CurrentLine int
	Modified    bool
	ReadOnly    bool
}

func New(theme *themes.Theme) *StatusLine {
	return &StatusLine{
		width:  consts.MainMenuContentWidth,
		height: consts.StatusLineHeight,
		theme:  *theme,
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

	leftStyled := s.theme.StatusLineLeft.Render(leftContent)
	rightStyled := s.theme.StatusLineRight.Render(rightContent)

	leftWidth := lipgloss.Width(leftStyled)
	rightWidth := lipgloss.Width(rightStyled)
	spaceWidth := s.width - leftWidth - rightWidth - consts.StatusLinePadding

	emptySpace := ""
	if spaceWidth > 0 {
		emptySpace = s.theme.StatusLine.Width(spaceWidth).Render(strings.Repeat(" ", spaceWidth))
	}

	statusLine := leftStyled + emptySpace + rightStyled
	return s.theme.StatusLine.Render(statusLine)
}

func (s *StatusLine) renderLeftSide(info StatusInfo) string {
	var parts []string

	if info.Mode != "" {
		modeText := s.theme.StatusLineMode.Render(string(info.Mode))
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
		fileText := s.theme.StatusLineFile.Render(string(fileName))
		parts = append(parts, fileText)
	}

	return strings.Join(parts, " ")
}

func (s *StatusLine) renderRightSide(info StatusInfo) string {
	var parts []string

	if info.Position != "" {
		positionText := s.theme.StatusLineInfo.Render(info.Position)
		parts = append(parts, positionText)
	}

	if info.FileInfo != "" {
		fileInfoText := s.theme.StatusLineInfo.Render(info.FileInfo)
		parts = append(parts, fileInfoText)
	}

	separator := s.theme.StatusLineSeparator.Render(" | ")
	return strings.Join(parts, separator)
}

func CreateStatusInfo(currentModel common.TUIModel) StatusInfo {
	info := StatusInfo{}

	// Add specific information based on current session state
	switch currentModel.GetSessionState() {
	case consts.SessionStateMainMenu:
		info.Mode = consts.StatusModeMenu
		info.FileName = string(consts.StateNameMainMenu)
		info.FileInfo = "Ready"

	case consts.SessionStateList:
		info.Mode = consts.StatusModeList
		info.FileName = string(consts.StateNameList)
		if listModel, ok := currentModel.(*yummy_list.ListModel); ok {
			count := len(listModel.RecipeList.Items())
			selectedItem := listModel.RecipeList.SelectedItem()
			info.FileInfo = fmt.Sprintf("%d recipes", count)
			if selectedItem != nil {
				if recipeItem, ok := selectedItem.(recipe.RecipeWithDescription); ok {
					info.FileName = recipeItem.Title()
				}
			} else {
				info.FileName = ""
			}
		}

	case consts.SessionStateDetail:
		info.Mode = consts.StatusModeRecipe
		info.FileName = string(consts.StateNameDetail)
		if detailModel, ok := currentModel.(*detail.DetailModel); ok {
			if detailModel.CurrentRecipe != nil {
				recipeName := detailModel.CurrentRecipe.Name
				recipeID := detailModel.CurrentRecipe.ID
				author := detailModel.CurrentRecipe.Author
				if author != "" {
					author = fmt.Sprintf("(by %s)", author)
				}
				info.FileName = strings.Join([]string{fmt.Sprintf("(#%d)", recipeID), recipeName, author}, " ")
			} else {
				info.FileName = ""
			}
			// Add scroll position info
			scrollPos := detailModel.GetScrollPosition()
			totalLines := detailModel.GetContentHeight()
			info.Position = fmt.Sprintf("Line %d", scrollPos+1)
			info.CurrentLine = scrollPos + 1
			info.LineCount = totalLines
			info.FileInfo = fmt.Sprintf("%d lines", totalLines)
		}

	case consts.SessionStateEdit:
		info.Mode = consts.StatusModeEdit
		info.FileName = string(consts.StateNameEdit)
		info.Modified = true
		info.FileInfo = "Modified"

	case consts.SessionStateChat:
		info.Mode = consts.StatusModeChat
		info.FileName = string(consts.StateNameChat)
		info.FileInfo = "Chat Mode"

	case consts.SessionStateSessionSelector:
		info.Mode = consts.StatusModeSessionSelector
		info.FileName = "Session Selector"
		info.FileInfo = "Select Chat Session"
	}

	return info
}
