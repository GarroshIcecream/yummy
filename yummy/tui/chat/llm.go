package chat

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strings"
	"time"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	ui "github.com/GarroshIcecream/yummy/yummy/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// LLMService handles all language model interactions
type LLMService struct {
	model     llms.Model
	modelName string
	// agent        *agents.ConversationalAgent
	// executor     *agents.Executor
	ollamaStatus *OllamaServiceStatus
	ctx          context.Context
	toolManager  *ToolManager
}

type OllamaServiceStatus struct {
	Installed       bool
	Running         bool
	Functional      bool
	InstalledModels []string
	ModelAvailable  bool
	Errors          []string
}

type LLMResponse struct {
	Response         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Error            error
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

	// All checks passed
	log.Printf("Ollama check passed: model %s is available", ui.DefaultModel)
	return nil
}

func GetOllamaInstalledModels() ([]string, error) {
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	modelList := make([]string, 0)
	lines := strings.Split(string(output), "\n")
	for idx, line := range lines {
		if idx != 0 {
			fields := strings.Split(line, " ")
			if len(fields) > 1 {
				clean_model := strings.TrimSpace(fields[0])
				modelList = append(modelList, clean_model)
			}
		}
	}
	return modelList, nil
}

// GetOllamaServiceStatus returns a detailed status of the Ollama service
func GetOllamaServiceStatus() *OllamaServiceStatus {
	status := &OllamaServiceStatus{
		Installed:       false,
		Running:         false,
		Functional:      false,
		ModelAvailable:  false,
		InstalledModels: []string{},
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
	status.InstalledModels, err = GetOllamaInstalledModels()
	if err != nil {
		status.Errors = append(status.Errors, err.Error())
		return status
	}

	if slices.Contains(status.InstalledModels, ui.DefaultModel) {
		status.ModelAvailable = true
	} else {
		status.Errors = append(status.Errors, "required model "+ui.DefaultModel+" not found")
	}

	return status
}

// NewLLMService creates a new LLM service instance
func NewLLMService(db *db.CookBook) (*LLMService, error) {
	ctx := context.Background()
	// Check if Ollama is available before attempting to create the model
	if err := CheckOllamaAvailable(); err != nil {
		return nil, fmt.Errorf("ollama check failed: %w", err)
	}

	toolManager := NewToolManager()
	modelName := ui.DefaultModel
	model, err := ollama.New(
		ollama.WithModel(modelName),
		ollama.WithSystemPrompt(ui.SystemPrompt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to serve LLM model: %w", err)
	}

	ollamaStatus := GetOllamaServiceStatus()

	// agent := agents.NewConversationalAgent(
	// 	model,
	// 	tools,
	// 	agents.WithMaxIterations(3),
	// )

	// executor := agents.NewExecutor(agent)

	llmService := &LLMService{
		// agent:        agent,
		// executor:     executor,
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
	model, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		return err
	}

	l.setModel(model)
	return nil
}

func (l *LLMService) setModel(model llms.Model) {
	l.model = model
	// l.agent = agents.NewConversationalAgent(
	// 	l.model,
	// 	l.toolManager.GetTools(),
	// 	agents.WithMaxIterations(3),
	// )
	// l.executor = agents.NewExecutor(l.agent)
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
func ParseChoices(completion *llms.ContentResponse) ([]*llms.ContentChoice, error) {
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
func (l *LLMService) GenerateResponse(conversation []llms.MessageContent) ui.ResponseMsg {
	log.Printf("Generating response with model: %s", l.modelName)
	answer, err := l.model.GenerateContent(
		context.Background(),
		conversation,
		llms.WithModel(l.modelName),
		llms.WithTemperature(ui.Temperature),
		llms.WithCandidateCount(1),
		llms.WithTools(l.toolManager.GetTools()),
	)
	if err != nil {
		log.Printf("model.GenerateContent error: %v", err)
		return ui.ResponseMsg{
			Response:         "",
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
			Error:            err,
		}
	}

	if len(answer.Choices) > 0 {
		output := answer.Choices[0].Content
		log.Printf("Generated response: %s", output)
		return ui.ResponseMsg{
			Response:         output,
			PromptTokens:     answer.Choices[0].GenerationInfo["PromptTokens"].(int),
			CompletionTokens: answer.Choices[0].GenerationInfo["CompletionTokens"].(int),
			TotalTokens:      answer.Choices[0].GenerationInfo["TotalTokens"].(int),
			Error:            nil,
		}
	}

	log.Printf("No response from model - no choices available")
	return ui.ResponseMsg{
		Response:         "",
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		Error:            fmt.Errorf("no response from model"),
	}
}

// GenerateResponseAsync generates a response asynchronously and returns a tea.Cmd
func (l *LLMService) GenerateResponseAsync(conversation []llms.MessageContent) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Starting async response generation with model: %s", l.modelName)
		response := l.GenerateResponse(conversation)
		log.Printf("Async response generation completed: %v", response)
		return response
	}
}
