package main

import (
	"fmt"

	models "github.com/GarroshIcecream/yummy/models"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Printf("Error initializing app: %v\n", err)
		return
	}
	defer f.Close()

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
