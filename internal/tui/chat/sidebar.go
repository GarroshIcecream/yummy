package chat

import (
	"fmt"
	"strings"

	db "github.com/GarroshIcecream/yummy/internal/db"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
)

func RenderSidebar(sessionStats db.SessionStats, ollamaStatus OllamaServiceStatus, executorService *ExecutorService, theme *themes.Theme, sidebarWidth int, sidebarHeight int) string {
	var sidebar strings.Builder

	// Status indicator
	if ollamaStatus.Functional && ollamaStatus.ModelAvailable {
		sidebar.WriteString(theme.SidebarSuccess.Render("● ") + theme.SidebarValue.Render("connected"))
	} else {
		sidebar.WriteString(theme.SidebarError.Render("● ") + theme.SidebarContent.Render("offline"))
	}

	// Tools section
	sidebar.WriteString("\n")
	sidebar.WriteString(theme.SidebarSection.Render("Tools"))
	sidebar.WriteString("\n")
	if executorService != nil {
		tools := executorService.toolManager.GetTools()
		if len(tools) > 0 {
			for _, tool := range tools {
				sidebar.WriteString(theme.SidebarContent.Render("  · " + tool.Name()))
				sidebar.WriteString("\n")
			}
		} else {
			sidebar.WriteString(theme.SidebarContent.Render("  · none"))
			sidebar.WriteString("\n")
		}
	}

	// Stats section
	sidebar.WriteString(theme.SidebarSection.Render("Stats"))
	sidebar.WriteString("\n")
	sidebar.WriteString(theme.SidebarContent.Render("  msgs   ") + theme.SidebarValue.Render(fmt.Sprintf("%d", sessionStats.MessageCount)))
	sidebar.WriteString("\n")
	sidebar.WriteString(theme.SidebarContent.Render("  tokens ") + theme.SidebarValue.Render(formatTokenCount(sessionStats.TotalTokens)))

	// Keys section
	sidebar.WriteString("\n")
	sidebar.WriteString(theme.SidebarSection.Render("Keys"))
	sidebar.WriteString("\n")
	keys := []struct{ key, desc string }{
		{"enter  ", "send"},
		{"↑/↓    ", "scroll"},
		{"ctrl+n ", "sessions"},
		{"ctrl+a ", "new"},
	}
	for _, k := range keys {
		sidebar.WriteString("  " + theme.SidebarValue.Render(k.key) + theme.SidebarContent.Render(k.desc))
		sidebar.WriteString("\n")
	}

	sidebarStyle := theme.Sidebar.Width(sidebarWidth - 4).Height(sidebarHeight)
	return sidebarStyle.Render(sidebar.String())
}

// formatTokenCount formats large token counts with K/M suffixes for better readability
func formatTokenCount(count int) string {
	if count >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(count)/1000000.0)
	} else if count >= 1000 {
		return fmt.Sprintf("%.1fK", float64(count)/1000.0)
	}
	return fmt.Sprintf("%d", count)
}
