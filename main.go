package main

import (
	"log"
	"os"

	"LogS/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	client := NewJiraClient()

	service, err := NewWorklogService(client)
	if err != nil {
		log.Fatal("Error initializing service:", err)
	}

	app := tui.NewApp(client, service)

	p := tea.NewProgram(app,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program:", err)
		os.Exit(ExitCodeError)
	}
}
