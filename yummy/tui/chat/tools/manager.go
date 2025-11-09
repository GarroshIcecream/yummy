package tools

import (
	"log"

	"github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/tmc/langchaingo/tools"
)

// ToolManager manages available tools and their execution
type ToolManager struct {
	tools    []tools.Tool
	cookbook *db.CookBook
}

// NewToolManagerWithCookbook creates a new tool manager with cookbook access
func NewToolManager(cookbook *db.CookBook) *ToolManager {
	tm := &ToolManager{
		tools:    make([]tools.Tool, 0),
		cookbook: cookbook,
	}

	// ddg, err := duckduckgo.New(10, "github.com/GarroshIcecream/yummy/yummy/tui/chat/tools")
	// if err != nil {
	// 	slog.Error("Failed to create DuckDuckGo tool", "error", err)
	// }

	tm.RegisterTool(NewGetRecipeNameTool(cookbook))
	tm.RegisterTool(NewGetRecipeIdTool(cookbook))
	//tm.RegisterTool(ddg)
	return tm
}

// RegisterTool registers a new tool with the manager
func (tm *ToolManager) RegisterTool(tool tools.Tool) {
	log.Printf("Registered tool: %s (%s)", tool.Name(), tool.Description())
	tm.tools = append(tm.tools, tool)
}

// GetTools returns all available tools converted to tools.Tool for langchaingo agents
func (tm *ToolManager) GetTools() []tools.Tool {
	return tm.tools
}
