package messages

import (
	"github.com/GarroshIcecream/yummy/yummy/config"
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
	Recipe *utils.RecipeRaw
}

type SaveMsg struct {
	RecipeID uint
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

type OpenModalViewMsg struct {
	ModalModel tea.Model
	ModalType  common.ModalType
}

type CloseModalViewMsg struct{}

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

type ModelSelectedMsg struct {
	ModelName string
}

type LoadSessionMsg struct {
	SessionID uint
}

type RenderConversationAsMarkdownMsg struct{}

type FavouriteSetMsg struct {
	Content string
}

func CmdHandler(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

func SendCloseDialogMsg() tea.Cmd {
	return CmdHandler(CloseDialogMsg{})
}

func SendFavouriteSetMsg(isFavourite bool) tea.Cmd {
	var content string
	if isFavourite {
		content = config.GetListConfig().ViewStatusMessageFavouriteSet
	} else {
		content = config.GetListConfig().ViewStatusMessageFavouriteRemoved
	}
	return CmdHandler(FavouriteSetMsg{Content: content})
}

func SendRecipeSelectedMsg(recipeID uint) tea.Cmd {
	return CmdHandler(RecipeSelectedMsg{RecipeID: recipeID})
}

func SendOpenModalViewMsg(modalModel tea.Model, modalType common.ModalType) tea.Cmd {
	return CmdHandler(OpenModalViewMsg{ModalModel: modalModel, ModalType: modalType})
}

func SendCloseModalViewMsg() tea.Cmd {
	return CmdHandler(CloseModalViewMsg{})
}

func SendSessionStateMsg(sessionState common.SessionState) tea.Cmd {
	return CmdHandler(SessionStateMsg{SessionState: sessionState})
}

func SendEditRecipeMsg(recipe *utils.RecipeRaw) tea.Cmd {
	return CmdHandler(EditRecipeMsg{Recipe: recipe})
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

func SendModelSelectedMsg(modelName string) tea.Cmd {
	return CmdHandler(ModelSelectedMsg{ModelName: modelName})
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
