package messages

import (
	"errors"

	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	recipes "github.com/GarroshIcecream/yummy/yummy/recipe"
	tea "github.com/charmbracelet/bubbletea"
)

type SessionMessage struct {
	SessionID    uint
	Message      string
	Role         string
	ModelName    string
	Content      string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

type GenerateResponseMsg struct{}

type LoadSessionsMsg struct{}

type RecipeSelectedMsg struct {
	RecipeID uint
}

type SessionStateMsg struct {
	SessionState consts.SessionState
}

type EditRecipeMsg struct {
	RecipeID uint
}

type SaveMsg struct {
	Recipe *recipes.RecipeRaw
	Err    error
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

type SessionSelectedMsg struct {
	SessionID uint
}

type LoadSessionMsg struct {
	SessionID uint
	Messages  []SessionMessage
	Err       error
}

func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

func SendCloseDialogMsg() tea.Cmd {
	return CmdHandler(CloseDialogMsg{})
}

func SendRecipeSelectedMsg(recipe_id uint) tea.Cmd {
	return CmdHandler(RecipeSelectedMsg{RecipeID: recipe_id})
}

func SendSessionStateMsg(session_state consts.SessionState) tea.Cmd {
	return CmdHandler(SessionStateMsg{SessionState: session_state})
}

func SendEditRecipeMsg(recipe_id uint) tea.Cmd {
	return CmdHandler(EditRecipeMsg{RecipeID: recipe_id})
}

func SendLoadRecipeMsg(msg LoadRecipeMsg) tea.Cmd {
	return CmdHandler(msg)
}

func SendLoadSessionsMsg() tea.Cmd {
	return CmdHandler(LoadSessionsMsg{})
}

func SendStatusInfoMsg(msg string, msgType int, ttl int) tea.Cmd {
	return CmdHandler(StatusInfoMsg{
		Msg:  msg,
		Type: msgType,
		TTL:  ttl,
	})
}

func SendEmptyResponseMsg() tea.Cmd {
	err := errors.New("empty response")
	return CmdHandler(ResponseMsg{
		Response:         consts.EmptyResponse,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		Error:            err,
	})
}

func SendResponseMsg(response ResponseMsg) tea.Cmd {
	return CmdHandler(response)
}

func SendGenerateResponseMsg() tea.Cmd {
	return CmdHandler(GenerateResponseMsg{})
}

func SendSetFavouriteMsg(recipe_id uint) tea.Cmd {
	return CmdHandler(SetFavouriteMsg{RecipeID: recipe_id})
}

func SendSessionSelectedMsg(sessionID uint) tea.Cmd {
	return CmdHandler(SessionSelectedMsg{SessionID: sessionID})
}

func SendLoadSessionMsg(sessionID uint, messages []SessionMessage, err error) tea.Cmd {
	return CmdHandler(LoadSessionMsg{
		SessionID: sessionID,
		Messages:  messages,
		Err:       err,
	})
}
