package chat

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"charm.land/bubbles/v2/key"
	"github.com/GarroshIcecream/yummy/internal/config"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/GarroshIcecream/yummy/internal/utils"
)

// RenderSidebar draws the session panel shown beside the conversation: the
// current session's facts and the chat key bindings. It mirrors the chat
// panel's rounded frame (via renderTitledPanel).
//
// sidebarWidth is the panel's total width; contentHeight is the body height,
// chosen by the caller so the panel spans the conversation and composer.
func RenderSidebar(executorService *ExecutorService, keyMap config.ChatKeyMap, theme *themes.Theme, sidebarWidth int, contentHeight int) string {
	contentWidth := max(10, sidebarWidth-4)

	// Every span sits on the base canvas (the chat page is one flat surface).
	section := theme.SidebarSection.MarginTop(0).Background(theme.BaseBg)
	content := theme.SidebarContent.Background(theme.BaseBg)
	value := theme.SidebarValue.Background(theme.BaseBg)

	var b strings.Builder

	// All rows share one grid: section headers flush-left, rows indented two
	// columns, second column aligned. Labels are padded by display width.
	const indent = "  "
	row := func(label, val string, labelWidth int) {
		pad := strings.Repeat(" ", max(labelWidth-lipgloss.Width(label), 1))
		b.WriteString(content.Render(indent+label+pad) + value.Render(val))
		b.WriteString("\n")
	}

	sessionID := "new"
	msgs := 0
	if executorService != nil {
		if id := executorService.GetSessionID(); id != 0 {
			sessionID = fmt.Sprintf("#%d", id)
		}
		msgs = executorService.VisibleMessageCount()
	}
	row("id", sessionID, 6)
	row("msgs", fmt.Sprintf("%d", msgs), 6)

	b.WriteString("\n")
	b.WriteString(section.Render("KEYS"))
	b.WriteString("\n")
	keys := []struct {
		binding key.Binding
		desc    string
	}{
		{keyMap.Enter, "send"},
		{keyMap.NewSession, "new chat"},
		{keyMap.SessionSelector, "sessions"},
		{keyMap.ModelSelector, "model"},
		{keyMap.ToggleSidebar, "sidebar"},
	}
	keyRow := func(k, desc string) {
		pad := strings.Repeat(" ", max(7-lipgloss.Width(k), 1))
		b.WriteString(value.Render(indent+k+pad) + content.Render(desc))
		b.WriteString("\n")
	}
	keyRow("↑/↓", "scroll")
	for _, k := range keys {
		if len(k.binding.Keys()) == 0 {
			continue
		}
		keyRow(shortKey(k.binding.Keys()[0]), k.desc)
	}

	// Re-establish the surface after every SGR reset and pad each line to the
	// content width, then fix the body to the panel's content box so the frame
	// aligns with the chat panel (height) and fills uniformly on the base
	// canvas. FillBackground is the can't-miss-a-spot guard against any cell
	// falling through to the terminal background.
	filled := utils.FillBackground(strings.TrimRight(b.String(), "\n"), theme.BaseBg, contentWidth)
	body := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Background(theme.BaseBg).
		Render(filled)

	title := lipgloss.NewStyle().
		Foreground(theme.SidebarSection.GetForeground()).
		Background(theme.BaseBg).
		Bold(true).
		Render("Session")

	return renderTitledPanel(theme, body, title, "", contentWidth, theme.BaseBg)
}

// shortKey abbreviates a key name the way terminals write it: "ctrl+n" -> "^n".
func shortKey(k string) string {
	if after, ok := strings.CutPrefix(k, "ctrl+"); ok {
		return "^" + after
	}
	return k
}
