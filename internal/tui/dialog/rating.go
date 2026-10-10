package dialog

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/utils"
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
	case tea.KeyPressMsg:
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

func (r *RatingDialogCmp) View() tea.View {

	// Build the single-line content: title | stars | hints
	var b strings.Builder

	b.WriteString(r.theme.RatingDialogTitle.Render("Rate recipe"))
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

	dot := r.theme.RatingDialogHelp.Render(" | ")
	b.WriteString("  ")
	b.WriteString(r.theme.RatingDialogHelp.Render("←→"))
	b.WriteString(dot)
	b.WriteString(r.theme.RatingDialogHelp.Render("enter"))
	b.WriteString(dot)
	b.WriteString(r.theme.RatingDialogHelp.Render("esc"))

	// Fill the interior with the elevated surface so the segments (title, stars,
	// hints separated by foreground-only spaces) sit on a solid card instead of
	// leaving bare cells that show the canvas behind.
	line := utils.FillBackground(b.String(), r.theme.SurfaceBg, lipgloss.Width(b.String()))
	rendered := r.theme.RatingDialogBox.Render(line)

	return tea.NewView(r.theme.RatingDialogContainer.Render(rendered))
}
