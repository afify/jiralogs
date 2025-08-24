package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func SimulateProgressSteps() tea.Cmd {
	return tea.Sequence(
		tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
			return progressMsg{percent: 0.2, step: "Authenticating user..."}
		}),
		tea.Tick(600*time.Millisecond, func(t time.Time) tea.Msg {
			return progressMsg{percent: 0.4, step: "Fetching tickets..."}
		}),
		tea.Tick(900*time.Millisecond, func(t time.Time) tea.Msg {
			return progressMsg{percent: 0.6, step: "Loading worklogs..."}
		}),
		tea.Tick(1200*time.Millisecond, func(t time.Time) tea.Msg {
			return progressMsg{percent: 0.8, step: "Calculating summary..."}
		}),
		tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
			return progressMsg{percent: 0.95, step: "Preparing display..."}
		}),
	)
}

func SimulateActionProgress(completionMsg tea.Msg) tea.Cmd {
	return tea.Sequence(
		tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
			return progressMsg{percent: 0.25, step: "Processing..."}
		}),
		tea.Tick(400*time.Millisecond, func(t time.Time) tea.Msg {
			return progressMsg{percent: 0.5, step: "Loading..."}
		}),
		tea.Tick(600*time.Millisecond, func(t time.Time) tea.Msg {
			return progressMsg{percent: 0.75, step: "Almost ready..."}
		}),
		tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
			return progressMsg{percent: 1.0, step: "Complete!"}
		}),
		tea.Tick(1000*time.Millisecond, func(t time.Time) tea.Msg {
			return completionMsg
		}),
	)
}
