package main

import (
	"fmt"
	"log"
	"os"

	models "github.com/GarroshIcecream/yummy/models"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	m, err := models.NewManager()
	if err != nil {
		fmt.Printf("Error initializing app: %v\n", err)
		return
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithMouseAllMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
