package app

import (
	"context"
	"sync"

	"LogS/internal/config"
	"LogS/jira"
	"LogS/shared"
)

type App struct {
	config *config.Config

	// JIRA services
	jiraClient     *jira.JiraClient
	worklogService *jira.WorklogService

	// Application context
	globalCtx context.Context

	// For cleanup
	cleanupFuncs []func()
	mutex        sync.RWMutex
}

// New initializes a new application instance.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	app := &App{
		globalCtx:    ctx,
		config:       cfg,
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
}
