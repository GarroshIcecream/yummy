package chat

import (
	"fmt"
	"strings"

	styles "github.com/GarroshIcecream/yummy/yummy/tui/styles"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
)

func RenderSidebar(messageCount int, tokenCount int, ollamaStatus OllamaServiceStatus, llmService *LLMService, sidebarWidth int, sidebarHeight int) string {
	var sidebar strings.Builder

	// Model Information
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Model"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("• %s", ui.LlamaModel)))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("• Thinking On"))
	sidebar.WriteString("\n\n")

	// Usage Statistics
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Usage"))
	sidebar.WriteString("\n")
	// Calculate rough percentage based on message count
	usagePercent := (messageCount * 5) % 100 // Simple calculation for demo
	sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("%d%% (%dK) $%.2f", usagePercent, tokenCount/1000, float64(tokenCount)*0.0001)))
	sidebar.WriteString("\n\n")

	// Ollama Health Status
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Ollama Status"))
	sidebar.WriteString("\n")

	status := ollamaStatus
	if status.Functional && status.ModelAvailable {
		sidebar.WriteString(styles.SidebarSuccessStyle.Render("✅ Service Healthy"))
	} else {
		sidebar.WriteString(styles.SidebarErrorStyle.Render("❌ Service Issues"))
		if len(status.Errors) > 0 {
			for _, err := range status.Errors {
				sidebar.WriteString("\n")
				sidebar.WriteString(styles.SidebarErrorStyle.Render(fmt.Sprintf("  • %s", err)))
			}
		}
	}
	sidebar.WriteString("\n\n")

	// Available Tools
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Available Tools"))
	sidebar.WriteString("\n")
	if llmService != nil && llmService.toolManager != nil {
		tools := llmService.toolManager.GetTools()
		for _, tool := range tools {
			sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("• %s", tool.Name())))
			sidebar.WriteString("\n")
		}
	} else {
		sidebar.WriteString(styles.SidebarContentStyle.Render("• No tools available"))
	}
	sidebar.WriteString("\n\n")

	// Session Stats
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Session Stats"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("• Messages: %d", messageCount)))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render(fmt.Sprintf("• Tokens: %d", tokenCount)))
	sidebar.WriteString("\n\n")

	// Controls
	sidebar.WriteString(styles.SidebarSectionStyle.Render("Controls"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("• Enter: Send message"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("• Ctrl+C: Exit"))
	sidebar.WriteString("\n")
	sidebar.WriteString(styles.SidebarContentStyle.Render("• ↑/↓: Scroll messages"))

	// Create a dynamic sidebar style based on the current width
	sidebarStyle := styles.SidebarStyle.Width(sidebarWidth - 4).Height(sidebarHeight)

	return sidebarStyle.Render(sidebar.String())
}
