package tui

import (
	"LogS/shared"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RenderInitializationPage renders the initialization view with banner in the middle
func (m *AppModel) RenderInitializationPage() string {
	// Wait for terminal size to be set
	if m.width == 0 || m.height == 0 {
		return "\n  Initializing..."
	}

	var content []string

	// Add the banner at the top
	banner := RenderBanner()
	content = append(content, banner)
	content = append(content, "")
	content = append(content, "")

	// Title
	title := InfoStyle.Render("🚀 Initializing JIRA Worklog Manager")
	content = append(content, title)
	content = append(content, "")

	// Progress bar with proper width
	progressWidth := 50
	if m.width > 70 {
		progressWidth = m.width - 40
		if progressWidth > 80 {
			progressWidth = 80
		}
	}

	// Create a local progress bar for initialization display
	prog := progress.New(
		progress.WithGradient(string(shared.GradientStart), string(shared.GradientEnd)),
		progress.WithWidth(progressWidth),
	)
	prog.SetPercent(m.progressPercent)
	bar := prog.View()
	content = append(content, bar)
	content = append(content, "")

	// Steps with real-time updates
	for i, step := range m.initSteps {
		prefix := "  ○ "
		style := DescStyle
		if i < m.initCurrentStep {
			// Completed step
			prefix = "  ✓ "
			style = SuccessStyle
		} else if i == m.initCurrentStep {
			// Current step in progress
			prefix = "  " + m.spinner.View() + " "
			style = InfoStyle
		}
		content = append(content, style.Render(prefix+step))
	}

	// Show error if initialization failed
	if m.initError != nil {
		content = append(content, "")
		content = append(content, ErrorStyle.Render("❌ Initialization failed: "+m.initError.Error()))
		content = append(content, "")
		content = append(content, HelpStyle.Render("Press R to retry or ESC to quit"))
	}

	// Center the content in the terminal
	contentStr := lipgloss.JoinVertical(lipgloss.Center, content...)

	// Use Place to center in the full terminal window
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		contentStr,
	)
}

// LoadInitializationData starts the initialization process
func (m *AppModel) LoadInitializationData() tea.Cmd {
	return m.startInitialization()
}

// HandleInitializationInput handles input for initialization view
func (m *AppModel) HandleInitializationInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "r":
		// Retry initialization if it failed
		if m.initError != nil {
			m.initError = nil
			m.initCurrentStep = 0
			m.progressPercent = 0.0
			return m.startInitialization()
		}
	case "esc":
		// Allow quitting if initialization failed
		if m.initError != nil {
			return tea.Quit
		}
	}
	return nil
}
