package messages

import (
	"github.com/GarroshIcecream/yummy/internal/config"
	common "github.com/GarroshIcecream/yummy/internal/models/common"
	utils "github.com/GarroshIcecream/yummy/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
)

type GenerateResponseMsg struct {
	UserInput    string // prompt sent to the LLM (may include resolved recipe context)
	DisplayInput string // compact text saved to memory/DB (e.g. with @[Recipe] intact)
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
	Response     string
	GenerationID uint64
}

type StreamingChunkMsg struct {
	Chunk        string
	GenerationID uint64
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

type ThemeSelectedMsg struct {
	ThemeName string
}

// CommandPaletteActionMsg is sent when the user selects a command from the palette.
type CommandPaletteActionMsg struct {
	Action string
}

type LoadSessionMsg struct {
	SessionID uint
}

type RenderConversationAsMarkdownMsg struct{}

type FavouriteSetMsg struct {
	Content string
}

// RecipeAddedFromURLMsg is sent when a recipe was successfully added from URL;
// the list view uses it to refresh and show a status message.
type RecipeAddedFromURLMsg struct {
	RecipeID      uint
	StatusMessage string
}

// ScrapeResultMsg is deprecated — retained for backward compatibility.
// Use scrapeAndSaveResultMsg (internal to dialog package) instead.
type ScrapeResultMsg struct {
	URL string
	Err error
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

func SendRecipeAddedFromURLMsg(recipeID uint, statusMessage string) tea.Cmd {
	return CmdHandler(RecipeAddedFromURLMsg{RecipeID: recipeID, StatusMessage: statusMessage})
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

func SendResponseMsg(response string, genID uint64) tea.Cmd {
	return CmdHandler(ResponseMsg{
		Response:     response,
		GenerationID: genID,
	})
}

func SendGenerateResponseMsg(userInput string, displayInput string) tea.Cmd {
	return CmdHandler(GenerateResponseMsg{UserInput: userInput, DisplayInput: displayInput})
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

func SendThemeSelectedMsg(themeName string) tea.Cmd {
	return CmdHandler(ThemeSelectedMsg{ThemeName: themeName})
}

func SendLoadSessionMsg(sessionID uint) tea.Cmd {
	return CmdHandler(LoadSessionMsg{
		SessionID: sessionID,
	})
}

func SendStreamingChunkMsg(chunk string, genID uint64) tea.Cmd {
	return CmdHandler(StreamingChunkMsg{Chunk: chunk, GenerationID: genID})
}

func SendRenderConversationAsMarkdownMsg() tea.Cmd {
	return CmdHandler(RenderConversationAsMarkdownMsg{})
}

func SendCommandPaletteActionMsg(action string) tea.Cmd {
	return CmdHandler(CommandPaletteActionMsg{Action: action})
}

type EnterCookingModeMsg struct {
	Recipe *utils.RecipeRaw
}

func SendEnterCookingModeMsg(recipe *utils.RecipeRaw) tea.Cmd {
	return CmdHandler(EnterCookingModeMsg{Recipe: recipe})
}

// RatingSelectedMsg is sent when the user confirms a rating in the rating dialog.
type RatingSelectedMsg struct {
	RecipeID uint
	Rating   int8
}

func SendRatingSelectedMsg(recipeID uint, rating int8) tea.Cmd {
	return CmdHandler(RatingSelectedMsg{RecipeID: recipeID, Rating: rating})
}
