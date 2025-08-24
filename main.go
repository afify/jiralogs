package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"LogS/internal/app"
	"LogS/internal/config"
	"LogS/internal/tui"
	"LogS/shared"
	tea "github.com/charmbracelet/bubbletea/v2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shared.LogErrorf("MAIN", "Starting jiralogs application")

	// Load configuration
	cfg := config.New()
	shared.LogErrorf("CONFIG", "Configuration loaded: IsConfigured=%v", cfg.IsConfigured())

	// Create app
	app, err := app.New(ctx, cfg)
	if err != nil {
		shared.LogError("APP", err)
		log.Fatalf("Failed to create app: %v", err)
	}
	defer app.Shutdown()
	shared.LogErrorf("APP", "Application instance created successfully")

	// Create TUI model
	model := tui.New(app)
	shared.LogErrorf("TUI", "TUI model created successfully")

	// Setup program options
	opts := []tea.ProgramOption{
		tea.WithMouseAllMotion(),
		tea.WithFilter(tui.MouseEventFilter),
	}

	// Create and run the program
	program := tea.NewProgram(model, opts...)
	shared.LogErrorf("PROGRAM", "Bubble Tea program created successfully")

	// Start the app subscription in background
	go app.Subscribe(program)
	shared.LogErrorf("SUBSCRIPTION", "App subscription started in background")

	// Handle shutdown signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		shared.LogErrorf("SIGNAL", "Shutdown signal received")
		program.Quit()
		cancel()
	}()

	// Run the program
	shared.LogErrorf("PROGRAM", "Starting Bubble Tea program")
	if _, err := program.Run(); err != nil {
		shared.LogError("PROGRAM", err)
		log.Fatalf("Program error: %v", err)
	}
	shared.LogErrorf("MAIN", "Application shutdown complete")
}
