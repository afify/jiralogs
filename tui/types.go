package tui

type ticketCreatedMsg struct {
	ticketKey string
}

type timeLoggedMsg struct {
	ticketKey string
	hours     float64
	date      string
}

// Initialization messages
type initStepMsg struct {
	Step    string
	Current int
	Total   int
}

type initCompleteMsg struct {
	client  JiraClientInterface
	service WorklogServiceInterface
}

type initErrorMsg struct {
	err     error
	details string
}
