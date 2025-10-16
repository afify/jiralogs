package app

import (
	"context"
	"sync"

	"LogS/config"
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

	// Initialize JIRA services (config is guaranteed to exist by main.go)
	shared.LogErrorf("APP_INIT", "Initializing JIRA client")
	jiraClient := jira.NewJiraClient(cfg.JiraBaseURL, cfg.JiraEmail, cfg.JiraToken)

	// Create service - only show spinner if we need to fetch user data
	var worklogService *jira.WorklogService
	var err error

	// Check if user is cached first
	if jira.IsCachedUserAvailable() {
		// User is cached, load silently
		worklogService, err = jira.NewWorklogService(jiraClient)
	} else {
		// User not cached, show status and make API call
		shared.StatusBar("📡 Fetching user info from JIRA API")
		worklogService, err = jira.NewWorklogService(jiraClient)
		if err == nil {
			shared.StatusComplete()
		} else {
			shared.StatusFailed()
		}
	}

	if err != nil {
		shared.LogError("APP_INIT", err)
		return nil, err
	}

	app.jiraClient = jiraClient
	app.worklogService = worklogService
	shared.LogErrorf("APP_INIT", "JIRA services initialized successfully")

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
