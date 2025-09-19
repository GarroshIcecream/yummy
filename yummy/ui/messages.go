package ui

import (
	recipes "github.com/GarroshIcecream/yummy/yummy/recipe"
	tui "github.com/GarroshIcecream/yummy/yummy/tui/utils"
	tea "github.com/charmbracelet/bubbletea"
)

type RecipeSelectedMsg struct {
	RecipeID uint
}

type SessionStateMsg struct {
	SessionState SessionState
}

type SaveMsg struct {
	Recipe *recipes.RecipeRaw
	Err    error
}

type EditRecipeMsg struct {
	RecipeID uint
}

type LoadRecipeMsg struct {
	Recipe   *recipes.RecipeRaw
	Markdown string
	Content  string
	Err      error
}

type StatusInfoMsg struct {
	Msg  string
	Type int
	TTL  int
}

func SendRecipeSelectedMsg(recipe_id uint) tea.Cmd {
	return tui.CmdHandler(RecipeSelectedMsg{RecipeID: recipe_id})
}

func SendSessionStateMsg(session_state SessionState) tea.Cmd {
	return tui.CmdHandler(SessionStateMsg{SessionState: session_state})
}

func SendEditRecipeMsg(recipe_id uint) tea.Cmd {
	return tui.CmdHandler(EditRecipeMsg{RecipeID: recipe_id})
}

func SendLoadRecipeMsg(msg LoadRecipeMsg) tea.Cmd {
	return tui.CmdHandler(msg)
}

func SendStatusInfoMsg(msg string, msgType int, ttl int) tea.Cmd {
	return tui.CmdHandler(StatusInfoMsg{
		Msg:  msg,
		Type: msgType,
		TTL:  ttl,
	})
}
