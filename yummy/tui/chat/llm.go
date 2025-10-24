package chat

import (
	"context"
	"fmt"
	"log"
	"slices"

	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	db "github.com/GarroshIcecream/yummy/yummy/db"
	messages "github.com/GarroshIcecream/yummy/yummy/models/msg"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type LLMService struct {
	model        *ollama.LLM
	modelName    string
	ollamaStatus *OllamaServiceStatus
	ctx          context.Context
	toolManager  *ToolManager
}

type LLMResponse struct {
	Response         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Error            error
}

// NewLLMService creates a new LLM service instance
func NewLLMService(db *db.CookBook) (*LLMService, error) {
	ctx := context.Background()
	toolManager := NewToolManager()
	modelName := consts.DefaultModel
	model, err := ollama.New(
		ollama.WithModel(modelName),
		ollama.WithSystemPrompt(consts.SystemPrompt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to serve LLM model: %w", err)
	}

	ollamaStatus := GetOllamaServiceStatus()
	llmService := &LLMService{
		model:        model,
		modelName:    modelName,
		ollamaStatus: ollamaStatus,
		ctx:          ctx,
		toolManager:  toolManager,
	}

	return llmService, nil
}

// SetModel sets the model for the LLM service
func (l *LLMService) SetModelByName(modelName string) error {
	if !slices.Contains(l.ollamaStatus.InstalledModels, modelName) {
		return fmt.Errorf("model %s is not installed", modelName)
	}

	l.modelName = modelName
	model, err := ollama.New(
		ollama.WithModel(modelName),
		ollama.WithSystemPrompt(consts.SystemPrompt),
	)
	if err != nil {
		return err
	}

	l.model = model

	return nil
}

func (l *LLMService) GetCurrentModel() *ollama.LLM {
	return l.model
}

func (l *LLMService) GetCurrentModelName() string {
	return l.modelName
}

func (l *LLMService) GetCurrentModelStatus() *OllamaServiceStatus {
	return l.ollamaStatus
}

func (l *LLMService) GetCurrentModelTools() []llms.Tool {
	return l.toolManager.GetTools()
}

func (l *LLMService) GetSystemPrompt() string {
	return consts.SystemPrompt
}

// AppendMessage adds a message to the conversation
func AppendMessage(conversation []llms.MessageContent, role llms.ChatMessageType, content string) []llms.MessageContent {
	conversation = append(conversation, llms.MessageContent{
		Role:  role,
		Parts: []llms.ContentPart{llms.TextPart(content)},
	})
	return conversation
}

// GetChoices extracts the first choice from a completion response
func ParseChoices(completion *llms.ContentResponse) ([]*llms.ContentChoice, error) {
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices available")
	}
	return completion.Choices, nil
}

// GenerateResponse generates a response for the given conversation
func (l *LLMService) GenerateResponse(conversation []llms.MessageContent) messages.ResponseMsg {
	log.Printf("Generating response with model: %s", l.modelName)
	response := messages.ResponseMsg{
		Response:         "",
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		Error:            nil,
	}

	answer, err := l.model.GenerateContent(
		context.Background(),
		conversation,
		llms.WithModel(l.modelName),
		llms.WithTemperature(consts.Temperature),
		llms.WithCandidateCount(1),
		llms.WithTools(l.toolManager.GetTools()),
	)
	if err != nil {
		log.Printf("model.GenerateContent error: %v", err)
		response.Error = err
		return response
	}

	choices, err := ParseChoices(answer)
	if err != nil {
		log.Printf("Error parsing choices: %v", err)
		response.Error = err
		return response
	}

	response.Response = choices[0].Content
	response.PromptTokens = choices[0].GenerationInfo["PromptTokens"].(int)
	response.CompletionTokens = choices[0].GenerationInfo["CompletionTokens"].(int)
	response.TotalTokens = choices[0].GenerationInfo["TotalTokens"].(int)

	return response
}
