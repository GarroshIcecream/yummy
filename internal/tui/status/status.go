package status

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/GarroshIcecream/yummy/internal/config"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/tui/chat"
	"github.com/GarroshIcecream/yummy/internal/tui/detail"
	yummy_list "github.com/GarroshIcecream/yummy/internal/tui/list"
	"github.com/GarroshIcecream/yummy/internal/utils"
	"github.com/charmbracelet/x/ansi"
)

type StatusLine struct {
	width       int
	height      int
	linePadding int
	theme       themes.Theme
}

// this sucks as we need to think of some other fields that are applicable to us
type StatusInfo struct {
	Mode        common.StatusMode
	Description string
	ModeInfo    string
	Position    string
	LineCount   int
	CurrentLine int
	Modified    bool
	ReadOnly    bool
}

func NewStatusLine(theme *themes.Theme) *StatusLine {
	cfg := config.GetStatusLineConfig()
	return &StatusLine{
		width:       cfg.ContentWidth,
		height:      cfg.Height,
		linePadding: cfg.Padding,
		theme:       *theme,
	}
}

func (s *StatusLine) SetSize(width int, height int) {
	s.width = width
	s.height = height
}

func (s *StatusLine) SetTheme(theme *themes.Theme) {
	s.theme = *theme
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
	if spaceWidth < 1 {
		// Too long for the bar (long recipe names): truncate the left side so
		// the bar never wraps onto a second line.
		leftStyled = ansi.Truncate(leftStyled, max(leftWidth+spaceWidth-1, 0), "…")
		spaceWidth = s.width - lipgloss.Width(leftStyled) - rightWidth - s.linePadding
	}

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
	if info.Description != "" {
		fileName := info.Description
		if info.Modified {
			fileName += " +"
		}
		if info.ReadOnly {
			fileName += " [RO]"
		}
		fileText := s.theme.StatusLineFile.Render(string(fileName))
		parts = append(parts, fileText)
	}

	// The joiner must carry the bar's background too, or the cell between the
	// mode badge and the description falls through to the canvas color.
	return strings.Join(parts, s.theme.StatusLineFile.UnsetPadding().Render(" "))
}

func (s *StatusLine) renderRightSide(info StatusInfo) string {
	var parts []string

	if info.Position != "" {
		positionText := s.theme.StatusLineInfo.Render(info.Position)
		parts = append(parts, positionText)
	}

	if info.ModeInfo != "" {
		ModeInfoText := s.theme.StatusLineInfo.Render(info.ModeInfo)
		parts = append(parts, ModeInfoText)
	}

	separator := s.theme.StatusLineSeparator.Render(" | ")
	return strings.Join(parts, separator)
}

func CreateStatusInfo(currentModel common.TUIModel) StatusInfo {
	info := StatusInfo{}

	// Add specific information based on current session state
	switch currentModel.GetSessionState() {
	case common.SessionStateMainMenu:
		info.Mode = common.StatusModeMenu
		info.Description = common.SessionStateMainMenu.GetStateName()
		info.ModeInfo = "Ready"

	case common.SessionStateList:
		info.Mode = common.StatusModeList
		info.Description = common.SessionStateList.GetStateName()
		if listModel, ok := currentModel.(*yummy_list.ListModel); ok {
			count := len(listModel.RecipeList.Items())
			selectedItem := listModel.RecipeList.SelectedItem()
			info.ModeInfo = fmt.Sprintf("%d recipes", count)
			if selectedItem != nil {
				if recipeItem, ok := selectedItem.(utils.RecipeRaw); ok {
					info.Description = recipeItem.Title()
				}
			} else {
				info.Description = ""
			}
		}

	case common.SessionStateDetail:
		info.Mode = common.StatusModeRecipe
		info.Description = common.SessionStateDetail.GetStateName()
		if detailModel, ok := currentModel.(*detail.DetailModel); ok {
			if detailModel.Recipe != nil {
				recipeName := detailModel.Recipe.RecipeName
				recipeID := detailModel.Recipe.RecipeID
				author := detailModel.Recipe.Metadata.Author
				if author != "" {
					author = fmt.Sprintf("(by %s)", author)
				}
				info.Description = strings.Join([]string{fmt.Sprintf("(#%d)", recipeID), recipeName, author}, " ")
			} else {
				info.Description = ""
			}
			// Scroll position: visible line range and a vim-style position.
			scrollPos := detailModel.GetScrollPosition()
			totalLines := detailModel.GetContentHeight()
			visible := detailModel.GetViewportHeight()
			last := min(scrollPos+visible, totalLines)
			info.CurrentLine = scrollPos + 1
			info.LineCount = totalLines
			if totalLines > 0 {
				info.Position = fmt.Sprintf("%d–%d of %d", min(scrollPos+1, totalLines), last, totalLines)
				info.ModeInfo = scrollPercent(scrollPos, last, totalLines)
			}
		}

	case common.SessionStateEdit:
		info.Mode = common.StatusModeEdit
		info.Description = common.SessionStateEdit.GetStateName()
		info.Modified = true
		info.ModeInfo = "Modified"

	case common.SessionStateChat:
		info.Mode = common.StatusModeChat
		if chatModel, ok := currentModel.(*chat.ChatModel); ok {
			modelName := chatModel.ExecutorService.GetCurrentModelName()
			info.ModeInfo = chatModel.ExecutorService.GetProviderName() + " | " + modelName

			// Get session summary if available
			summary, err := chatModel.ExecutorService.GetSessionSummary()
			if err == nil {
				if len(summary) > 60 {
					summary = summary[:57] + "..."
				}
				info.Description = summary
			}
		} else {
			info.Description = ""
		}

	case common.SessionStateCooking:
		info.Mode = common.StatusModeCooking
		if cookingModel, ok := currentModel.(*detail.CookingModel); ok {
			if cookingModel.Recipe != nil {
				info.Description = cookingModel.Recipe.RecipeName
				info.Position = fmt.Sprintf("Step %d/%d", cookingModel.CurrentStep+1, cookingModel.TotalSteps)
				info.ModeInfo = "Cooking Mode"
			}
		}
	}

	return info
}

// scrollPercent describes a scroll position the way vim does: "All" when the
// whole document fits, "Top"/"Bot" at either end, otherwise a percentage.
func scrollPercent(first, last, total int) string {
	switch {
	case first == 0 && last >= total:
		return "All"
	case first == 0:
		return "Top"
	case last >= total:
		return "Bot"
	default:
		return fmt.Sprintf("%d%%", first*100/max(total-(last-first), 1))
	}
}
