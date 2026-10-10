package tools

import (
	"log/slog"

	"github.com/GarroshIcecream/yummy/internal/db"
	"github.com/tmc/langchaingo/tools"
)

// ToolManager manages available tools and their execution
type ToolManager struct {
	tools    []tools.Tool
	cookbook *db.CookBook
}

// NewToolManager creates a new tool manager with cookbook access
func NewToolManager(cookbook *db.CookBook) *ToolManager {
	tm := &ToolManager{
		tools:    make([]tools.Tool, 0),
		cookbook: cookbook,
	}

	tm.RegisterTool(NewGetRecipeNameTool(cookbook))
	tm.RegisterTool(NewGetRecipeIdTool(cookbook))
	tm.RegisterTool(NewListRecipesTool(cookbook))
	tm.RegisterTool(NewFindByIngredientsTool(cookbook))
	tm.RegisterTool(NewCookbookOverviewTool(cookbook))
	return tm
}

// RegisterTool registers a new tool with the manager
func (tm *ToolManager) RegisterTool(tool tools.Tool) {
	slog.Debug("Registered tool", "name", tool.Name())
	tm.tools = append(tm.tools, tool)
}

// GetTools returns all available tools converted to tools.Tool for langchaingo agents
func (tm *ToolManager) GetTools() []tools.Tool {
	return tm.tools
}
