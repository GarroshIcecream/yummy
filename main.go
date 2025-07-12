package main

import (
	"fmt"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	manager "github.com/GarroshIcecream/yummy/yummy/models/manager"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Printf("Error initializing app: %v\n", err)
		return
	}
	defer f.Close()

	cookbook, err := db.NewCookBook("my_cookbook.db")
	if err != nil {
		fmt.Printf("Error initializing app: %v\n", err)
		return
	}

	m, err := manager.New(cookbook)
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
