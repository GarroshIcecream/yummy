package chat

import (
	"fmt"
	"strings"

	themes "github.com/GarroshIcecream/yummy/yummy/themes"
)

func RenderSidebar(messageCount int, tokenCount int, ollamaStatus OllamaServiceStatus, llmService *LLMService, theme *themes.Theme, sidebarWidth int, sidebarHeight int) string {
	var sidebar strings.Builder

	// Model Information
	if llmService != nil {
		sidebar.WriteString(theme.SidebarSection.Render(fmt.Sprintf("🧠 Model: %s", llmService.modelName)))
		sidebar.WriteString("\n\n")
	}

	// Ollama Health Status
	status := ollamaStatus
	if status.Functional && status.ModelAvailable {
		sidebar.WriteString(theme.SidebarSection.Render("🔧 Ollama Status: ✅"))
	} else {
		sidebar.WriteString(theme.SidebarSection.Render("🔧 Ollama Status: ❌"))
		sidebar.WriteString("\n")
		if status.Error != nil {
			sidebar.WriteString(theme.SidebarError.Render(fmt.Sprintf("   • %s", status.Error)))
			sidebar.WriteString("\n")
		}
	}
	sidebar.WriteString("\n")

	// Available Tools
	sidebar.WriteString(theme.SidebarSection.Render("🛠️  Available Tools"))
	sidebar.WriteString("\n")
	if llmService != nil && llmService.toolManager != nil {
		tools := llmService.toolManager.GetTools()
		if len(tools) > 0 {
			for _, tool := range tools {
				sidebar.WriteString(theme.SidebarContent.Render(fmt.Sprintf("   • %s", tool.Function.Name)))
				sidebar.WriteString("\n")
			}
		} else {
			sidebar.WriteString(theme.SidebarContent.Render("   No tools loaded"))
			sidebar.WriteString("\n")
		}
	} else {
		sidebar.WriteString(theme.SidebarContent.Render("   No tool manager"))
		sidebar.WriteString("\n")
	}
	sidebar.WriteString("\n")

	// Session Stats
	sidebar.WriteString(theme.SidebarSection.Render("📊 Session Stats"))
	sidebar.WriteString("\n")
	sidebar.WriteString(theme.SidebarContent.Render(fmt.Sprintf("   Messages: %d", messageCount)))
	sidebar.WriteString("\n")
	sidebar.WriteString(theme.SidebarContent.Render(fmt.Sprintf("   Tokens: %s", formatTokenCount(tokenCount))))
	sidebar.WriteString("\n\n")

	// Controls
	sidebar.WriteString(theme.SidebarSection.Render("⌨️  Controls"))
	sidebar.WriteString("\n")
	sidebar.WriteString(theme.SidebarContent.Render("   Enter      Send message"))
	sidebar.WriteString("\n")
	sidebar.WriteString(theme.SidebarContent.Render("   ↑/↓        Scroll"))
	sidebar.WriteString("\n")
	sidebar.WriteString(theme.SidebarContent.Render("   Ctrl+N     Load Session"))

	// Create a dynamic sidebar style based on the current width
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
