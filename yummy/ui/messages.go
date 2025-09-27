package ui

import (
	"errors"

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

type CloseDialogMsg struct{}

type StatusInfoMsg struct {
	Msg  string
	Type int
	TTL  int
}

type ResponseMsg struct {
	Response         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Error            error
}

type SetFavouriteMsg struct {
	RecipeID uint
}

func SendCloseDialogMsg() tea.Cmd {
	return tui.CmdHandler(CloseDialogMsg{})
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

func SendEmptyResponseMsg() tea.Cmd {
	err := errors.New("empty response")
	return tui.CmdHandler(ResponseMsg{
		Response:         EmptyResponse,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		Error:            err,
	})
}

func SendResponseMsg(response string) tea.Cmd {
	return tui.CmdHandler(ResponseMsg{
		Response:         response,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		Error:            nil,
	})
}

func SendSetFavouriteMsg(recipe_id uint) tea.Cmd {
	return tui.CmdHandler(SetFavouriteMsg{RecipeID: recipe_id})
}
