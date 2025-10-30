package messages

import (
	common "github.com/GarroshIcecream/yummy/yummy/models/common"
	utils "github.com/GarroshIcecream/yummy/yummy/utils"
	tea "github.com/charmbracelet/bubbletea"
)

type GenerateResponseMsg struct {
	UserInput string
}

type LoadSessionsMsg struct{}

type RecipeSelectedMsg struct {
	RecipeID uint
}

type SessionStateMsg struct {
	SessionState common.SessionState
}

type EditRecipeMsg struct {
	RecipeID uint
}

type SaveMsg struct {
	Recipe *utils.RecipeRaw
	Err    error
}

type LoadRecipeMsg struct {
	Recipe   *utils.RecipeRaw
	Markdown string
	Content  string
}

type CloseDialogMsg struct{}

type StatusInfoMsg struct {
	Msg  string
	Type int
	TTL  int
}

type ResponseMsg struct {
	Response string
}

type StreamingChunkMsg struct {
	Chunk string
}

type SetFavouriteMsg struct {
	RecipeID uint
}

type SessionSelectedMsg struct {
	SessionID uint
}

type LoadSessionMsg struct {
	SessionID uint
}

type RenderConversationAsMarkdownMsg struct{}

func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

func SendCloseDialogMsg() tea.Cmd {
	return CmdHandler(CloseDialogMsg{})
}

func SendRecipeSelectedMsg(recipeID uint) tea.Cmd {
	return CmdHandler(RecipeSelectedMsg{RecipeID: recipeID})
}

func SendSessionStateMsg(sessionState common.SessionState) tea.Cmd {
	return CmdHandler(SessionStateMsg{SessionState: sessionState})
}

func SendEditRecipeMsg(recipeID uint) tea.Cmd {
	return CmdHandler(EditRecipeMsg{RecipeID: recipeID})
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

func SendResponseMsg(response string) tea.Cmd {
	return CmdHandler(ResponseMsg{
		Response: response,
	})
}

func SendGenerateResponseMsg(userInput string) tea.Cmd {
	return CmdHandler(GenerateResponseMsg{UserInput: userInput})
}

func SendSetFavouriteMsg(recipeID uint) tea.Cmd {
	return CmdHandler(SetFavouriteMsg{RecipeID: recipeID})
}

func SendSessionSelectedMsg(sessionID uint) tea.Cmd {
	return CmdHandler(SessionSelectedMsg{SessionID: sessionID})
}

func SendLoadSessionMsg(sessionID uint) tea.Cmd {
	return CmdHandler(LoadSessionMsg{
		SessionID: sessionID,
	})
}

func SendStreamingChunkMsg(chunk string) tea.Cmd {
	return CmdHandler(StreamingChunkMsg{Chunk: chunk})
}

func SendRenderConversationAsMarkdownMsg() tea.Cmd {
	return CmdHandler(RenderConversationAsMarkdownMsg{})
}
