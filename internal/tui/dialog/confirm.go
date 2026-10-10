package dialog

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	utils "github.com/GarroshIcecream/yummy/internal/utils"
)

// ConfirmDialogCmp asks the user to confirm a destructive action. On confirm it
// closes itself and emits onConfirm; cancel (n, esc, or the Cancel button)
// just closes. Cancel is focused by default so a stray enter is harmless.
type ConfirmDialogCmp struct {
	title        string
	message      string
	confirmLabel string
	onConfirm    tea.Msg
	confirmFocus bool

	theme  *themes.Theme
	width  int
	height int
}

// NewConfirmDialog creates a confirmation dialog for a destructive action.
func NewConfirmDialog(title, message, confirmLabel string, onConfirm tea.Msg, theme *themes.Theme) *ConfirmDialogCmp {
	return &ConfirmDialogCmp{
		title:        title,
		message:      message,
		confirmLabel: confirmLabel,
		onConfirm:    onConfirm,
		theme:        theme,
		width:        52,
		height:       10,
	}
}

func (c *ConfirmDialogCmp) Init() tea.Cmd {
	return nil
}

func (c *ConfirmDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return c, nil
	}
	switch strings.ToLower(keyMsg.String()) {
	case "left", "right", "tab", "shift+tab", "h", "l":
		c.confirmFocus = !c.confirmFocus
	case "y":
		return c, c.confirm()
	case "n":
		return c, messages.SendCloseModalViewMsg()
	case "enter":
		if c.confirmFocus {
			return c, c.confirm()
		}
		return c, messages.SendCloseModalViewMsg()
	}
	return c, nil
}

func (c *ConfirmDialogCmp) confirm() tea.Cmd {
	return tea.Sequence(messages.SendCloseModalViewMsg(), messages.CmdHandler(c.onConfirm))
}

func (c *ConfirmDialogCmp) View() tea.View {
	innerWidth := max(c.width-6, 20) // border (2) + padding (4)

	title := c.theme.DialogDangerTitle.Render(c.title)
	escHint := c.theme.CommandPaletteHelp.Render("esc")
	header := title + strings.Repeat(" ", max(innerWidth-lipgloss.Width(title)-lipgloss.Width(escHint), 1)) + escHint
	sep := c.theme.CommandPaletteShortcut.Render(strings.Repeat("─", innerWidth))

	body := c.theme.DialogUnselectedRow.Width(innerWidth).Render(c.message)

	confirmStyle, cancelStyle := c.theme.DialogDangerButtonBlurred, c.theme.DialogButtonFocused
	if c.confirmFocus {
		confirmStyle, cancelStyle = c.theme.DialogDangerButton, c.theme.DialogButton
	}
	buttons := confirmStyle.Render(c.confirmLabel) + "  " + cancelStyle.Render("Cancel")

	help := c.theme.CommandPaletteHelp.Render("y yes | n no | ←→ switch")

	content := lipgloss.JoinVertical(lipgloss.Left, header, sep, "", body, "", buttons, "", help)
	content = utils.FillBackground(content, c.theme.SurfaceBg, innerWidth)
	return tea.NewView(c.theme.CommandPaletteDialog.Width(c.width).Render(content))
}

func (c *ConfirmDialogCmp) SetSize(width, height int) {
	c.width = clampModalWidth(width, c.width, 30)
	c.height = clampModalHeight(height, c.height, 8)
}

func (c *ConfirmDialogCmp) GetSize() (int, int) {
	return c.width, c.height
}

func (c *ConfirmDialogCmp) GetModelState() common.ModelState {
	return common.ModelStateLoaded
}
