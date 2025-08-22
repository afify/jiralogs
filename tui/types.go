package tui

type ticketCreatedMsg struct {
	ticketKey string
}

type timeLoggedMsg struct {
	ticketKey string
	hours     float64
	date      string
}
