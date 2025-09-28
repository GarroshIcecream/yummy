package chat

import (
	"fmt"
	"strings"

	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
)

func RenderSidebar(messageCount int, tokenCount int, ollamaStatus OllamaServiceStatus, llmService *LLMService, sidebarWidth int, sidebarHeight int) string {
	var sidebar strings.Builder

	// Model Information
	if llmService != nil {
		sidebar.WriteString(styles.SidebarSectionStyle.Render(fmt.Sprintf("🧠 Model: %s", llmService.modelName)))
		sidebar.WriteString("\n\n")
	}

	// Ollama Health Status
	status := ollamaStatus
	if status.Functional && status.ModelAvailable {
		sidebar.WriteString(styles.SidebarSectionStyle.Render("🔧 Ollama Status: ✅"))
	} else {
		sidebar.WriteString(styles.SidebarSectionStyle.Render("🔧 Ollama Status: ❌"))
		sidebar.WriteString("\n")
		if len(status.Errors) > 0 {
			for _, err := range status.Errors {
				sidebar.WriteString(styles.SidebarErrorStyle.Render(fmt.Sprintf("   • %s", err)))
				sidebar.WriteString("\n")
			}
		}
	}
	sidebar.WriteString("\n")

	// Available Tools
	sidebar.WriteString(styles.SidebarSectionStyle.Render("🛠️  Available Tools"))
	sidebar.WriteString("\n")
	if llmService != nil && llmService.toolManager != nil {
		tools := llmService.toolManager.GetTools()
		if len(tools) > 0 {
			for _, tool := range tools {
				sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("   • %s", tool.Function.Name)))
				sidebar.WriteString("\n")
			}
		} else {
			sidebar.WriteString(styles.SidebarContentStyle.Render("   No tools loaded"))
			sidebar.WriteString("\n")
		}
	} else {
		sidebar.WriteString(styles.SidebarContentStyle.Render("   No tool manager"))
		sidebar.WriteString("\n")
	}
	sidebar.WriteString("\n")

	// Session Stats
	sidebar.WriteString(styles.SidebarSectionStyle.Render("📊 Session Stats"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("   Messages: %d", messageCount)))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("   Tokens: %s", formatTokenCount(tokenCount))))
	sidebar.WriteString("\n\n")

	// Controls
	sidebar.WriteString(styles.SidebarSectionStyle.Render("⌨️  Controls"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("   Enter      Send message"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("   ↑/↓        Scroll"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("   Ctrl+N     Load Session"))

	// Create a dynamic sidebar style based on the current width
	sidebarStyle := styles.SidebarStyle.Width(sidebarWidth - 4).Height(sidebarHeight)

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
