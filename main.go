package main

import (
	"log"
	"os"

	"LogS/jira"
	"LogS/shared"
	"LogS/tui"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func main() {
	// Start TUI immediately with initialization function
	app := tui.NewApp(initializeApp)

	p := tea.NewProgram(app,
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program:", err)
		os.Exit(shared.ExitCodeError)
	}
}

func initializeApp() (tui.JiraClientInterface, tui.WorklogServiceInterface, error) {
	// Load configuration
	if err := LoadConfig(); err != nil {
		return nil, nil, err
	}

	// Create Jira client
	client := jira.NewJiraClient(BaseURL, Email, APIToken)

	// Create worklog service
	service, err := jira.NewWorklogService(client)
	if err != nil {
		return nil, nil, err
	}

	// Wrap them to implement the interfaces
	return &jiraClientWrapper{client}, &worklogServiceWrapper{service}, nil
}
