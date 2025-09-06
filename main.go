package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	db "github.com/GarroshIcecream/yummy/yummy/db"
	"github.com/GarroshIcecream/yummy/yummy/log"
	manager "github.com/GarroshIcecream/yummy/yummy/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Printf("Error initializing app: %v\n", err)
		return
	}
	defer f.Close()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	dbDir := filepath.Join(homeDir, ".yummy")
	dbPath := filepath.Join(dbDir, "cookbook.db")

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		fmt.Printf("Error creating database directory: %v\n", err)
		return
	}

	cookbook, err := db.NewCookBook(dbPath)
	if err != nil {
		fmt.Printf("Error initializing app: %v\n", err)
		return
	}

	m, err := manager.New(cookbook)
	if err != nil {
		fmt.Printf("Error initializing app: %v\n", err)
		return
	}

	defer log.RecoverPanic("main", func() {
		slog.Error("Application terminated due to unhandled panic")
	})

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
