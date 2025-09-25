package chat

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	ui "github.com/GarroshIcecream/yummy/yummy/ui"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// LLMService handles all language model interactions
type LLMService struct {
	model       llms.Model
	modelName   string
	agent       *agents.ConversationalAgent
	executor    *agents.Executor
	ctx         context.Context
	toolManager *ToolManager
}

type OllamaServiceStatus struct {
	Installed       bool
	Running         bool
	Functional      bool
	ModelAvailable  bool
	Errors          []string
}

type LLMResponse struct {
	Response string
	Error    error
}

// CheckOllamaServiceRunning checks if the Ollama service is running and responsive
func CheckOllamaServiceRunning() error {
	// Check if ollama command exists first
	_, err := exec.LookPath("ollama")
	if err != nil {
		return fmt.Errorf("ollama command not found in PATH")
	}

	// Try to ping the service
	cmd := exec.Command("ollama", "ps")
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("ollama service is not running or not responding")
	}

	// Try to list models to ensure service is fully functional
	cmd = exec.Command("ollama", "list")
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("ollama service is running but not fully functional")
	}

	return nil
}

// StartOllamaService attempts to start the Ollama service
func StartOllamaService() error {
	// Check if ollama command exists first
	_, err := exec.LookPath("ollama")
	if err != nil {
		return fmt.Errorf("ollama command not found in PATH")
	}

	// Try to start the service in the background
	cmd := exec.Command("ollama", "serve")
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ollama service: %w", err)
	}

	// Give the service a moment to start up
	time.Sleep(2 * time.Second)

	// Check if the service is now running
	err = CheckOllamaServiceRunning()
	if err != nil {
		return fmt.Errorf("ollama service failed to start properly: %w", err)
	}

	return nil
}

// CheckOllamaAvailable checks if Ollama is installed and the required model is available
func CheckOllamaAvailable() error {
	// Step 1: Check if ollama command exists
	_, err := exec.LookPath("ollama")
	if err != nil {
		return fmt.Errorf(`ollama is not installed or not found in PATH.

To fix this:
1. Install Ollama from https://ollama.ai
2. Make sure Ollama is added to your system PATH
3. Restart your terminal/command prompt
4. Try running this application again

For more help, visit: https://ollama.ai/install`)
	}

	// Step 2: Check if Ollama service is running
	err = CheckOllamaServiceRunning()
	if err != nil {
		// Try to start the service automatically
		log.Printf("Ollama service not running, attempting to start it...")
		startErr := StartOllamaService()
		if startErr != nil {
			return fmt.Errorf(`ollama service is not running and could not be started automatically.

To fix this:
1. Start the Ollama service manually by running: ollama serve
2. Or restart your computer if Ollama is set to start automatically
3. Make sure no firewall is blocking Ollama
4. Check if there are any error messages in the Ollama logs
5. Try running this application again

Service check error: %v
Start attempt error: %v`, err, startErr)
		}
		log.Printf("Successfully started Ollama service")
	}

	// Step 2b: Final verification that service is working
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(`ollama service is running but not fully functional.

To fix this:
1. Restart the Ollama service: ollama serve
2. Check if there are any error messages in the Ollama logs
3. Make sure you have sufficient disk space
4. Try running this application again

Error details: %w`, err)
	}

	// Step 3: Check if the required model is available
	modelList := string(output)
	if !strings.Contains(modelList, ui.LlamaModel) {
		return fmt.Errorf(`required model "%s" is not installed.

To fix this:
1. Run the following command in your terminal:
   ollama pull %s
2. Wait for the download to complete (this may take several minutes)
3. Try running this application again

Note: The model download requires internet connection and sufficient disk space`, ui.LlamaModel, ui.LlamaModel)
	}

	// All checks passed
	log.Printf("Ollama check passed: model %s is available", ui.LlamaModel)
	return nil
}

// GetOllamaServiceStatus returns a detailed status of the Ollama service
func GetOllamaServiceStatus() OllamaServiceStatus {
	status := OllamaServiceStatus{
		Installed:       false,
		Running:         false,
		Functional:      false,
		ModelAvailable: false,
		Errors:          []string{},
	}

	// Check if ollama command exists
	_, err := exec.LookPath("ollama")
	if err != nil {
		status.Errors = append(status.Errors, "ollama command not found in PATH")
		return status
	}
	status.Installed = true

	// Check if service is running
	err = CheckOllamaServiceRunning()
	if err != nil {
		status.Errors = append(status.Errors, err.Error())
		return status
	}
	status.Running = true
	status.Functional = true

	// Check if required model is available
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		status.Errors = append(status.Errors, "failed to list models: "+err.Error())
		return status
	}

	modelList := string(output)
	if strings.Contains(modelList, ui.LlamaModel) {
		status.ModelAvailable = true
	} else {
		status.Errors = append(status.Errors, "required model "+ui.LlamaModel+" not found")
	}

	return status
}

// NewLLMService creates a new LLM service instance
func NewLLMService() (*LLMService, error) {
	// Check if Ollama is available before attempting to create the model
	if err := CheckOllamaAvailable(); err != nil {
		return nil, fmt.Errorf("ollama check failed: %w", err)
	}

	modelName := ui.LlamaModel
	model, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		return nil, fmt.Errorf("failed to serve LLM model: %w", err)
	}

	toolManager := NewToolManager()
	tools := toolManager.GetTools()

	agent := agents.NewConversationalAgent(
		model,
		tools,
		agents.WithMaxIterations(3),
	)

	executor := agents.NewExecutor(agent)

	llmService := &LLMService{
		agent:       agent,
		executor:    executor,
		model:       model,
		modelName:   modelName,
		ctx:         context.Background(),
		toolManager: toolManager,
	}

	return llmService, nil
}

// SetModel sets the model for the LLM service
func (l *LLMService) SetModelByName(modelName string) error {
	l.modelName = modelName
	model, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		return err
	}
	
	l.setModel(model)
	return nil
}

func (l *LLMService) setModel(model llms.Model) {
	l.model = model
	l.agent = agents.NewConversationalAgent(
		l.model,
		l.toolManager.GetTools(),
		agents.WithMaxIterations(3),
	)
	l.executor = agents.NewExecutor(l.agent)
}

func (l *LLMService) GetCurrentModel() llms.Model {
	return l.model
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
func (l *LLMService) GenerateResponse(conversation []llms.MessageContent) *LLMResponse {
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

	// Use the agent executor to generate a response with tools
	answer, err := l.executor.Call(context.Background(), map[string]any{
		"input": input,
	})

	if err != nil {
		log.Printf("executor.Call error: %v", err)
		return &LLMResponse{
			Response: "",
			Error:    err,
		}
	}
	
	if output, ok := answer["output"].(string); ok {
		return &LLMResponse{
			Response: output,
			Error:    nil,
		}
	}
	
	return &LLMResponse{
		Response: "",
		Error:    fmt.Errorf("no response from model"),
	}
}