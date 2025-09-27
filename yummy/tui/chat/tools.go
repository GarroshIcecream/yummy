package chat

import (
	"log"

	"github.com/tmc/langchaingo/llms"
)

// ToolManager manages available tools and their execution
type ToolManager struct {
	tools []llms.Tool
}

// NewToolManager creates a new tool manager with available tools
func NewToolManager() *ToolManager {
	tm := &ToolManager{
		tools: make([]llms.Tool, 0),
	}

	return tm
}

// RegisterTool registers a new tool with the manager
func (tm *ToolManager) RegisterTool(tool llms.Tool) {
	log.Printf("Registered tool: %+v (%+v)", tool.Function.Name, tool.Function.Description)
	tm.tools = append(tm.tools, tool)
}

// GetTools returns all available tools
func (tm *ToolManager) GetTools() []llms.Tool {
	return tm.tools
}
