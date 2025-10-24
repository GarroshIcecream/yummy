package chat

import (
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strings"
	"time"

	consts "github.com/GarroshIcecream/yummy/yummy/consts"
)

type OllamaServiceStatus struct {
	Installed       bool
	Running         bool
	Functional      bool
	InstalledModels []string
	ModelAvailable  bool
	Error           error
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
	time.Sleep(5 * time.Second)

	// Check if the service is now running
	err = CheckOllamaServiceRunning()
	if err != nil {
		return fmt.Errorf("ollama service failed to start properly: %w", err)
	}

	return nil
}

// CheckOllamaAvailable checks if Ollama is installed and the required model is available
func CheckOllamaAvailable() error {
	_, err := exec.LookPath("ollama")
	if err != nil {
		return fmt.Errorf(consts.OllamaNotInstalledHelp)
	}

	err = CheckOllamaServiceRunning()
	if err != nil {
		log.Printf("Ollama service not running, attempting to start it...")
		startErr := StartOllamaService()
		if startErr != nil {
			return fmt.Errorf("%s\n\nService check error: %v\nStart attempt error: %v", consts.OllamaServiceNotRunningHelp, err, startErr)
		}
		log.Printf("Successfully started Ollama service")
	}

	log.Printf("Ollama check passed: model %s is available", consts.DefaultModel)
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
		Error:           nil,
	}

	// Check if service is running
	err := CheckOllamaAvailable()
	if err != nil {
		status.Error = err
		return status
	}
	status.Installed = true
	status.Running = true
	status.Functional = true

	status.InstalledModels, err = GetOllamaInstalledModels()
	if err != nil {
		status.Error = err
		return status
	}

	if slices.Contains(status.InstalledModels, consts.DefaultModel) {
		status.ModelAvailable = true
	} else {
		status.Error = fmt.Errorf("required model %s not found", consts.DefaultModel)
	}

	return status
}
