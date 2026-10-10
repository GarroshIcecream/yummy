package themes

import (
	"image/color"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// HuhTheme returns a Huh form theme derived from the active app theme's
// palette, so forms (the recipe editor) match the rest of the UI.
func HuhTheme(theme *Theme) huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		if theme == nil {
			return huh.ThemeCharm(isDark)
		}
		p := theme.Palette
		fg := func(c color.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(c) }

		t := huh.ThemeBase(isDark)
		t.Form.Base = lipgloss.NewStyle().Background(p.Base)

		// The focused field shows a primary accent bar on its left; blurred
		// fields keep the same (hidden) border width so nothing shifts.
		t.Focused.Base = t.Focused.Base.BorderForeground(p.Primary).BorderBackground(p.Base)
		t.Focused.Card = t.Focused.Base
		t.Focused.Title = fg(p.Primary).Bold(true)
		t.Focused.NoteTitle = fg(p.Primary).Bold(true).MarginBottom(1)
		t.Focused.Directory = fg(p.Primary)
		t.Focused.Description = fg(p.Subtle)
		t.Focused.ErrorIndicator = fg(p.Error).SetString(" *")
		t.Focused.ErrorMessage = fg(p.Error)
		t.Focused.SelectSelector = fg(p.Primary).SetString("> ")
		t.Focused.NextIndicator = fg(p.Primary)
		t.Focused.PrevIndicator = fg(p.Primary)
		t.Focused.Option = fg(p.Text)
		t.Focused.MultiSelectSelector = fg(p.Primary).SetString("> ")
		t.Focused.SelectedOption = fg(p.Success)
		t.Focused.SelectedPrefix = fg(p.Success).SetString("[x] ")
		t.Focused.UnselectedOption = fg(p.Text)
		t.Focused.UnselectedPrefix = fg(p.Subtle).SetString("[ ] ")
		t.Focused.FocusedButton = lipgloss.NewStyle().
			Foreground(p.OnPrimary).
			Background(p.Primary).
			Bold(true).
			Padding(0, 2).
			MarginRight(1)
		t.Focused.BlurredButton = lipgloss.NewStyle().
			Foreground(p.Muted).
			Background(p.Selection).
			Padding(0, 2).
			MarginRight(1)
		t.Focused.Next = t.Focused.FocusedButton

		t.Focused.TextInput.Cursor = fg(p.Primary)
		t.Focused.TextInput.CursorText = fg(p.Base).Background(p.Primary)
		t.Focused.TextInput.Placeholder = fg(p.Subtle)
		t.Focused.TextInput.Prompt = fg(p.Primary)
		t.Focused.TextInput.Text = fg(p.Text)

		t.Blurred = t.Focused
		t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
		t.Blurred.Card = t.Blurred.Base
		t.Blurred.Title = fg(p.Muted)
		t.Blurred.NoteTitle = fg(p.Muted)
		t.Blurred.TextInput.Prompt = fg(p.Subtle)
		t.Blurred.TextInput.Text = fg(p.Muted)
		t.Blurred.NextIndicator = lipgloss.NewStyle()
		t.Blurred.PrevIndicator = lipgloss.NewStyle()

		t.Group.Title = t.Focused.Title
		t.Group.Description = t.Focused.Description

		t.Help.Ellipsis = fg(p.Subtle)
		t.Help.ShortKey = fg(p.Muted)
		t.Help.ShortDesc = fg(p.Subtle)
		t.Help.ShortSeparator = fg(p.Border)
		t.Help.FullKey = fg(p.Muted)
		t.Help.FullDesc = fg(p.Subtle)
		t.Help.FullSeparator = fg(p.Border)

		return t
	})
}
