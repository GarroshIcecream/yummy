package chat

import (
	"log"

	tools "github.com/tmc/langchaingo/tools"
	scraper "github.com/tmc/langchaingo/tools/scraper"
)

// ToolManager manages available tools and their execution
type ToolManager struct {
	tools   []tools.Tool
}

// NewToolManager creates a new tool manager with available tools
func NewToolManager() *ToolManager {
	tm := &ToolManager{
		tools:   make([]tools.Tool, 0),
	}

	scrapeTool, err := scraper.New()
	if err != nil {
		log.Fatalf("Failed to create scraper tool: %v", err)
	}

	tm.RegisterTool(scrapeTool)
	
	return tm
}

// RegisterTool registers a new tool with the manager
func (tm *ToolManager) RegisterTool(tool tools.Tool) {
	log.Printf("Registered tool: %+v", tool.Name())
	tm.tools = append(tm.tools, tool)
}

// GetTools returns all available tools
func (tm *ToolManager) GetTools() []tools.Tool {
	return tm.tools
}
