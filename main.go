package main

import (
	"fmt"
	"os"
	"recipe_me/db"
	"recipe_me/models"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	_, err := db.NewCookBook()
	if err != nil {
		fmt.Println("Error creating CookBook:", err)
	}

	p := tea.NewProgram(models.NewModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
