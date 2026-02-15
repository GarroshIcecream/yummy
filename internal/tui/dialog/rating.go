package dialog

import (
	"strings"

	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RatingDialogCmp is a small top-anchored modal for setting a recipe rating.
type RatingDialogCmp struct {
	recipeID uint
	cursor   int8
	theme    *themes.Theme
}

// NewRatingDialog creates a new rating dialog for the given recipe.
// currentRating is the recipe's existing rating (0 means unrated).
func NewRatingDialog(recipeID uint, currentRating int8, theme *themes.Theme) *RatingDialogCmp {
	cursor := currentRating
	if cursor == 0 {
		cursor = 3
	}

	return &RatingDialogCmp{
		recipeID: recipeID,
		cursor:   cursor,
		theme:    theme,
	}
}

func (r *RatingDialogCmp) Init() tea.Cmd {
	return nil
}

func (r *RatingDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return r, messages.SendCloseModalViewMsg()

		case "left", "h":
			if r.cursor > 0 {
				r.cursor--
			}

		case "right", "l":
			if r.cursor < 5 {
				r.cursor++
			}

		case "enter":
			return r, tea.Batch(
				messages.SendRatingSelectedMsg(r.recipeID, r.cursor),
				messages.SendCloseModalViewMsg(),
			)
		}
	}

	return r, nil
}

func (r *RatingDialogCmp) View() string {
	// Thick block border matching add-recipe-from-URL style
	dialogBorder := lipgloss.Border{
		Top:         "▀",
		Bottom:      "▄",
		Left:        "▌",
		Right:       "▐",
		TopLeft:     "▛",
		TopRight:    "▜",
		BottomLeft:  "▙",
		BottomRight: "▟",
	}

	// Build the single-line content: title · stars · hints
	var b strings.Builder

	b.WriteString(r.theme.RatingDialogTitle.Render("Rate"))
	b.WriteString("  ")

	for i := int8(1); i <= 5; i++ {
		if i <= r.cursor {
			b.WriteString(r.theme.RatingStarActive.Render("★"))
		} else {
			b.WriteString(r.theme.RatingStarInactive.Render("☆"))
		}
		if i < 5 {
			b.WriteString(" ")
		}
	}

	dot := r.theme.RatingDialogHelp.Render(" · ")
	b.WriteString("  ")
	b.WriteString(r.theme.RatingDialogHelp.Render("←→"))
	b.WriteString(dot)
	b.WriteString(r.theme.RatingDialogHelp.Render("enter"))
	b.WriteString(dot)
	b.WriteString(r.theme.RatingDialogHelp.Render("esc"))

	rendered := lipgloss.NewStyle().
		Border(dialogBorder).
		BorderForeground(r.theme.RatingStarActive.GetForeground()).
		Padding(0, 2).
		Render(b.String())

	return r.theme.RatingDialogContainer.Render(rendered)
}
