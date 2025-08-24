package app

import (
	"context"
	"sync"

	"LogS/internal/config"
	"LogS/jira"
	"LogS/shared"
	tea "github.com/charmbracelet/bubbletea/v2"
)

type App struct {
	config *config.Config

	// JIRA services
	jiraClient     *jira.JiraClient
	worklogService *jira.WorklogService

	// Application context
	globalCtx context.Context
	events    chan tea.Msg

	// For cleanup
	cleanupFuncs []func()
	mutex        sync.RWMutex
}

// New initializes a new application instance.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	app := &App{
		globalCtx:    ctx,
		config:       cfg,
		events:       make(chan tea.Msg, 100),
		cleanupFuncs: []func(){},
	}

	// Initialize JIRA services if configured
	if cfg.IsConfigured() {
		shared.LogErrorf("APP_INIT", "Initializing JIRA client")
		jiraClient := jira.NewJiraClient(cfg.JiraBaseURL, cfg.JiraEmail, cfg.JiraToken)
		worklogService, err := jira.NewWorklogService(jiraClient)
		if err != nil {
			shared.LogError("APP_INIT", err)
			return nil, err
		}

		app.jiraClient = jiraClient
		app.worklogService = worklogService
		shared.LogErrorf("APP_INIT", "JIRA services initialized successfully")
	} else {
		shared.LogErrorf("APP_INIT", "JIRA not configured, skipping service initialization")
	}

	return app, nil
}

// Config returns the application configuration.
func (app *App) Config() *config.Config {
	return app.config
}

// WorklogService returns the JIRA worklog service.
func (app *App) WorklogService() *jira.WorklogService {
	return app.worklogService
}

// JiraClient returns the JIRA client.
func (app *App) JiraClient() *jira.JiraClient {
	return app.jiraClient
}

// Subscribe sends events to the TUI as tea.Msgs.
func (app *App) Subscribe(program *tea.Program) {
	// Simple event subscription for jiralogs
	// This is a simplified version compared to Crush
	go func() {
		for {
			select {
			case <-app.globalCtx.Done():
				return
			case msg, ok := <-app.events:
				if !ok {
					return
				}
				program.Send(msg)
			}
		}
	}()
}

// SendEvent sends an event to the TUI
func (app *App) SendEvent(msg tea.Msg) {
	select {
	case app.events <- msg:
	default:
		// Drop message if channel is full
	}
}

// Shutdown performs a graceful shutdown of the application.
func (app *App) Shutdown() {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// Call cleanup functions
	for _, cleanup := range app.cleanupFuncs {
		if cleanup != nil {
			cleanup()
		}
	}

	// Close events channel
	close(app.events)
}
