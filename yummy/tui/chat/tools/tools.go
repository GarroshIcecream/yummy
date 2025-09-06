package tools

import (
	"log"

	tools "github.com/tmc/langchaingo/tools"
)

// ToolManager manages available tools and their execution
type ToolManager struct {
	tools   []tools.Tool
	scraper *ScraperTool
}

// NewToolManager creates a new tool manager with available tools
func NewToolManager() *ToolManager {
	tm := &ToolManager{
		tools:   make([]tools.Tool, 0),
		scraper: NewScraperTool(),
	}

	// Register the scraping tool
	scrapeTool := NewScraperTool()

	tm.RegisterTool(scrapeTool)

	// Debug: Log the registered tool
	log.Printf("Registered tool: %+v", scrapeTool)

	return tm
}

// RegisterTool registers a new tool with the manager
func (tm *ToolManager) RegisterTool(tool tools.Tool) {
	tm.tools = append(tm.tools, tool)
}

// GetTools returns all available tools
func (tm *ToolManager) GetTools() []tools.Tool {
	return tm.tools
}
