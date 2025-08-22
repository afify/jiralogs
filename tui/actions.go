package tui

import (
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
	m.currentView = LoadingView

	return tea.Batch(
		m.progressBar.SetPercent(0),
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

	return tea.Batch(
		m.progressBar.SetPercent(0),
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

	return tea.Batch(
		m.progressBar.SetPercent(0),
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

	return tea.Batch(
		m.progressBar.SetPercent(0),
		SimulateActionProgress(periodSelectionReadyMsg{}),
	)
}

func (m *AppModel) getEmailFromConfig() string {
	return m.client.GetConfiguredEmail()
}
