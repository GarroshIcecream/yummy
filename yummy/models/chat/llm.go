package chat

import (
	"context"
	"fmt"
	"log"

	tools "github.com/GarroshIcecream/yummy/yummy/models/chat/tools"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// LLMService handles all language model interactions
type LLMService struct {
	model       llms.Model
	agent       agents.ConversationalAgent
	executor    agents.Executor
	ctx         context.Context
	toolManager *tools.ToolManager
}

// Message types for the tea program
type ResponseMsg struct {
	Content string
	Error   error
}

// CheckOllamaAvailable checks if Ollama is installed and available
// func CheckOllamaAvailable() error {
// 	// Check if ollama command exists
// 	_, err := exec.LookPath("ollama")
// 	if err != nil {
// 		return fmt.Errorf("ollama is not installed or not in PATH. Please install Ollama from https://ollama.ai")
// 	}

// 	// check that user has the model downloaded
// 	// execute this command in the background so that user wont see the output
// 	// if the model is not found, pull it
// 	// if the model is found, do nothing

// 	// check if the model is in the output
// 	if !strings.Contains(string(output), ui.LlamaModel) {
// 		log.Printf("Model %s not found, attempting to pull...", ui.LlamaModel)
// 		// pull the model if not available
// 		cmd = exec.Command("ollama", "pull", ui.LlamaModel)
// 		if err := cmd.Run(); err != nil {
// 			return fmt.Errorf("failed to pull model %s: %w. Try running 'ollama pull %s' manually", ui.LlamaModel, err, ui.LlamaModel)
// 		}
// 		log.Printf("Successfully pulled model %s", ui.LlamaModel)
// 	}

// 	go func() {
// 		cmd := exec.Command("ollama", "list")
// 		output, err := cmd.Output()
// 		if err != nil {
// 			log.Printf("failed to list models: %v", err)
// 		}
// 		if !strings.Contains(string(output), ui.LlamaModel) {
// 			log.Printf("Model %s not found, attempting to pull...", ui.LlamaModel)
// 			cmd = exec.Command("ollama", "pull", ui.LlamaModel)
// 			if err := cmd.Run(); err != nil {
// 				log.Printf("failed to pull model %s: %v. Try running 'ollama pull %s' manually", ui.LlamaModel, err, ui.LlamaModel)
// 			}
// 			log.Printf("Successfully pulled model %s", ui.LlamaModel)
// 		}
// 	}()

// 	return nil
// }

// NewLLMService creates a new LLM service instance
func NewLLMService() (*LLMService, error) {
	// Check if Ollama is available before attempting to create the model
	// if err := CheckOllamaAvailable(); err != nil {
	// 	return nil, fmt.Errorf("ollama check failed: %w", err)
	// }

	model, err := ollama.New(ollama.WithModel(ui.LlamaModel))
	if err != nil {
		return nil, fmt.Errorf("failed to serve LLM model: %w", err)
	}

	agent := agents.NewConversationalAgent(
		model,
		tools.NewToolManager().GetTools(),
		agents.WithMaxIterations(3),
	)

	executor := agents.NewExecutor(agent)

	llmService := &LLMService{
		agent:       *agent,
		executor:    *executor,
		model:       model,
		ctx:         context.Background(),
		toolManager: tools.NewToolManager(),
	}

	return llmService, nil
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
func GetChoices(completion *llms.ContentResponse) ([]*llms.ContentChoice, error) {
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no response from model")
	}
	return completion.Choices, nil
}

// GetSystemPrompt returns the system prompt for the cooking assistant
func (l *LLMService) GetSystemPrompt() string {
	basePrompt := ui.SystemPrompt
	return basePrompt
}

// GenerateResponse generates a response for the given conversation
func (l *LLMService) GenerateResponse(conversation []llms.MessageContent) ResponseMsg {
	// Debug: Log the conversation being sent
	log.Printf("Conversation length: %d", len(conversation))
	for i, msg := range conversation {
		log.Printf("Message %d - Role: %s, Content: %s", i, msg.Role, msg.Parts)
	}

	// Convert conversation to string for the agent
	var input string
	for _, msg := range conversation {
		if msg.Role == llms.ChatMessageTypeHuman {
			for _, part := range msg.Parts {
				if textPart, ok := part.(llms.TextContent); ok {
					input += textPart.Text + " "
				}
			}
		}
	}

	// Use the agent executor to generate a response
	answer, err := l.executor.Call(context.Background(), map[string]any{
		"input": input,
	})

	chains.Run(context.Background(), l.agent.Chain, conversation)

	// if there is an error, return it
	if err != nil {
		return ResponseMsg{
			Content: "",
			Error:   fmt.Errorf("failed to generate content: %w", err),
		}
	}

	// Get the answer from the array and convert to string
	answerStr := fmt.Sprintf("%v", answer["output"])

	return ResponseMsg{
		Content: answerStr,
		Error:   nil,
	}
}
