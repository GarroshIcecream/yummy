package chat

import (
	"fmt"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"time"
)

type OllamaServiceStatus struct {
	Installed       bool
	Running         bool
	Functional      bool
	InstalledModels []string
	ModelAvailable  bool
}

// CheckOllamaServiceRunning checks if the Ollama service is running and responsive
func CheckOllamaServiceRunning() error {
	// Check if ollama command exists first
	_, err := exec.LookPath("ollama")
	if err != nil {
		slog.Error("Ollama command not found in PATH", "error", err)
		return err
	}

	// Try to ping the service
	cmd := exec.Command("ollama", "ps")
	_, err = cmd.Output()
	if err != nil {
		slog.Error("Ollama service is not running or not responding", "error", err)
		return err
	}

	return nil
}

// StartOllamaService attempts to start the Ollama service
func StartOllamaService() error {
	// Check if ollama command exists first
	_, err := exec.LookPath("ollama")
	if err != nil {
		slog.Error("Ollama command not found in PATH", "error", err)
		return err
	}

	// Try to start the service in the background
	cmd := exec.Command("ollama", "serve")
	err = cmd.Start()
	if err != nil {
		slog.Error("Failed to start ollama service", "error", err)
		return err
	}

	// Give the service a moment to start up
	time.Sleep(5 * time.Second)

	// Check if the service is now running
	err = CheckOllamaServiceRunning()
	if err != nil {
		slog.Error("Ollama service failed to start properly", "error", err)
		return err
	}

	return nil
}

// CheckOllamaAvailable checks if Ollama is installed and the required model is available
func CheckOllamaAvailable() error {
	_, err := exec.LookPath("ollama")
	if err != nil {
		slog.Error("Ollama not installed", "error", err)
		return err
	}

	err = CheckOllamaServiceRunning()
	if err != nil {
		slog.Error("Ollama service not running, attempting to start it...", "error", err)
		startErr := StartOllamaService()
		if startErr != nil {
			slog.Error("Failed to start ollama service", "error", startErr)
			return startErr
		}
		slog.Info("Successfully started Ollama service")
	}

	return nil
}

func GetOllamaInstalledModels() ([]string, error) {
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to get ollama installed models", "error", err)
		return nil, err
	}

	modelList := make([]string, 0)
	lines := strings.Split(string(output), "\n")
	for idx, line := range lines {
		if idx != 0 {
			fields := strings.Split(line, " ")
			if len(fields) > 1 {
				cleanModel := strings.TrimSpace(fields[0])
				modelList = append(modelList, cleanModel)
			}
		}
	}
	slog.Debug("Ollama installed models", "models", modelList)
	return modelList, nil
}

// GetOllamaServiceStatus returns a detailed status of the Ollama service
func GetOllamaServiceStatus(modelName string) (*OllamaServiceStatus, error) {
	slog.Debug("Getting ollama service status", "model", modelName)
	status := &OllamaServiceStatus{
		Installed:       false,
		Running:         false,
		Functional:      false,
		ModelAvailable:  false,
		InstalledModels: []string{},
	}

	// Check if service is running
	err := CheckOllamaAvailable()
	if err != nil {
		slog.Error("Failed to check ollama available", "error", err)
		return nil, err
	}
	status.Installed = true
	status.Running = true
	status.Functional = true

	status.InstalledModels, err = GetOllamaInstalledModels()
	if err != nil {
		slog.Error("Failed to get ollama installed models", "error", err)
		return nil, err
	}

	if slices.Contains(status.InstalledModels, modelName) {
		status.ModelAvailable = true
	} else {
		slog.Error("Required model not found", "model", modelName)
		return nil, fmt.Errorf("required model %s not found", modelName)
	}

	return status, nil
}
