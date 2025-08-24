package tui

import (
	"fmt"

	"LogS/shared"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *AppModel) loadWorklogsWithProgress(message string) tea.Cmd {
	m.loading = true
	m.loadingMsg = message
	m.progressPercent = 0
	m.currentStep = 0
	m.loadingSteps = []string{
		"Connecting to JIRA...",
		"Authenticating user...",
		"Fetching tickets...",
		"Loading worklogs...",
		"Calculating summary...",
		"Preparing display...",
	}
	// Show progress popup for loading
	m.progressPopup.ShowWithSteps("Loading Worklogs", m.loadingSteps)

	return tea.Batch(
		SimulateProgressSteps(),
		LoadWorklogsWithProgress(m.service, m.period),
	)
}

func (m *AppModel) loadCurrentUser() tea.Cmd {
	return func() tea.Msg {
		user, err := m.client.GetCurrentUser()
		if err != nil {
			return userLoadedMsg{user: nil}
		}

		jiraUser := &shared.JiraUser{
			AccountID:    user.AccountID,
			EmailAddress: m.client.GetConfiguredEmail(),
		}

		return userLoadedMsg{user: jiraUser}
	}
}

func (m *AppModel) loadTimeLoggingWithProgress() tea.Cmd {
	m.loading = true
	m.loadingMsg = "Preparing time logging form..."
	m.progressPercent = 0
	m.currentStep = 0
	m.loadingSteps = []string{
		"Loading user tickets...",
		"Preparing form fields...",
		"Validating permissions...",
		"Ready!",
	}

	// Show progress popup
	m.progressPopup.ShowWithSteps("Preparing Time Logging", m.loadingSteps)

	return tea.Batch(
		SimulateActionProgress(timeLoggingReadyMsg{}),
	)
}

func (m *AppModel) loadTicketCreationWithProgress() tea.Cmd {
	m.loading = true
	m.loadingMsg = "Preparing ticket creation form..."
	m.progressPercent = 0
	m.currentStep = 0
	m.loadingSteps = []string{
		"Loading projects...",
		"Fetching issue types...",
		"Preparing form fields...",
		"Ready!",
	}

	// Show progress popup
	m.progressPopup.ShowWithSteps("Preparing Ticket Creation", m.loadingSteps)

	return tea.Batch(
		SimulateActionProgress(ticketCreationReadyMsg{}),
	)
}

func (m *AppModel) loadPeriodSelectionWithProgress() tea.Cmd {
	m.loading = true
	m.loadingMsg = "Loading period options..."
	m.progressPercent = 0
	m.currentStep = 0
	m.loadingSteps = []string{
		"Calculating date ranges...",
		"Preparing options...",
		"Ready!",
	}

	// Show progress popup
	m.progressPopup.ShowWithSteps("Loading Period Selection", m.loadingSteps)

	return tea.Batch(
		SimulateActionProgress(periodSelectionReadyMsg{}),
	)
}

func (m *AppModel) getEmailFromConfig() string {
	return m.client.GetConfiguredEmail()
}

func (m *AppModel) startInitialization() tea.Cmd {
	return func() tea.Msg {
		if m.initFunc == nil {
			return initErrorMsg{
				err:     fmt.Errorf("initialization function not provided"),
				details: "No initialization function was provided",
			}
		}

		client, service, err := m.initFunc()
		if err != nil {
			return initErrorMsg{
				err:     err,
				details: fmt.Sprintf("Initialization failed: %v", err),
			}
		}

		return initCompleteMsg{
			client:  client,
			service: service,
		}
	}
}
