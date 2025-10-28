package status

import (
	"fmt"
	"strings"

	"github.com/GarroshIcecream/yummy/yummy/config"
	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	"github.com/GarroshIcecream/yummy/yummy/recipe"
	themes "github.com/GarroshIcecream/yummy/yummy/themes"
	"github.com/GarroshIcecream/yummy/yummy/tui/chat"
	"github.com/GarroshIcecream/yummy/yummy/tui/detail"
	yummy_list "github.com/GarroshIcecream/yummy/yummy/tui/list"
	"github.com/charmbracelet/lipgloss"
)

type StatusLine struct {
	width       int
	height      int
	linePadding int
	theme       themes.Theme
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

func New(theme *themes.Theme, config *config.StatusLineConfig) *StatusLine {
	return &StatusLine{
		width:       config.ContentWidth,
		height:      config.Height,
		linePadding: config.Padding,
		theme:       *theme,
	}
}

func (s *StatusLine) SetSize(width int, height int) {
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
	spaceWidth := s.width - leftWidth - rightWidth - s.linePadding

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
		info.FileName = consts.SessionStateMainMenu.GetStateName()
		info.FileInfo = "Ready"

	case consts.SessionStateList:
		info.Mode = consts.StatusModeList
		info.FileName = consts.SessionStateList.GetStateName()
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
		info.FileName = consts.SessionStateDetail.GetStateName()
		if detailModel, ok := currentModel.(*detail.DetailModel); ok {
			if detailModel.Recipe != nil {
				recipeName := detailModel.Recipe.Name
				recipeID := detailModel.Recipe.ID
				author := detailModel.Recipe.Author
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
		info.FileName = consts.SessionStateEdit.GetStateName()
		info.Modified = true
		info.FileInfo = "Modified"

	case consts.SessionStateChat:
		info.Mode = consts.StatusModeChat
		if chatModel, ok := currentModel.(*chat.ChatModel); ok {
			info.FileName = chatModel.ExecutorService.GetCurrentModelName()
		} else {
			info.FileName = ""
		}
		info.FileInfo = "Chat Mode"
	}

	return info
}
