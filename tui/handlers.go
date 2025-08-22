package tui

import (
	"time"

	"LogS/shared"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *AppModel) handleMainMenuKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		i, ok := m.mainMenu.SelectedItem().(menuItem)
		if ok {
			switch i.title {
			case "📊 View Worklogs":
				if len(m.ticketLogs) > 0 {
					m.pushView(WorklogDisplayView)
				} else {
					m.currentView = LoadingView
					return m.loadWorklogsWithProgress("Loading worklogs...")
				}
			case "⏰ Log Time":
				m.currentView = LoadingView
				return m.loadTimeLoggingWithProgress()
			case "🎫 Create Ticket":
				m.currentView = LoadingView
				return m.loadTicketCreationWithProgress()
			case "📅 Filter Period":
				m.currentView = LoadingView
				return m.loadPeriodSelectionWithProgress()
			case "🔄 Refresh":
				m.currentView = LoadingView
				return m.loadWorklogsWithProgress("Refreshing data...")
			case "❌ Quit":
				return tea.Quit
			}
		}
	}
	return nil
}

func (m *AppModel) handlePeriodSelectionKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		i, ok := m.periodList.SelectedItem().(periodItem)
		if ok {
			now := time.Now()
			var start time.Time

			switch i.id {
			case "3":
				start = now.AddDate(0, -3, 0)
			case "6":
				start = now.AddDate(0, -6, 0)
			case "9":
				start = now.AddDate(0, -9, 0)
			case "w":
				weekday := int(now.Weekday())
				if weekday == 0 {
					weekday = 7
				}
				daysBack := (weekday - 1) % 7
				start = now.AddDate(0, 0, -daysBack)
				start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
			case "y":
				start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
			default:
				start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			}

			m.period = shared.Period{Start: start, End: now}
			m.currentView = LoadingView
			return m.loadWorklogsWithProgress("Loading data for new period...")
		}
	}
	return nil
}

func (m *AppModel) handleWorklogDisplayKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "l":
		m.currentView = LoadingView
		return m.loadTimeLoggingWithProgress()
	case "c":
		m.currentView = LoadingView
		return m.loadTicketCreationWithProgress()
	case "r":
		m.currentView = LoadingView
		return m.loadWorklogsWithProgress("Refreshing worklogs...")
	}
	return nil
}

func (m *AppModel) handleTimeLoggingKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "tab", "down":
		m.activeInput = (m.activeInput + 1) % len(m.textInputs)
		return m.updateInputFocus()
	case "shift+tab", "up":
		m.activeInput--
		if m.activeInput < 0 {
			m.activeInput = len(m.textInputs) - 1
		}
		return m.updateInputFocus()
	}
	return nil
}

func (m *AppModel) updateInputFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))
	for i := range m.textInputs {
		if i == m.activeInput {
			cmds[i] = m.textInputs[i].Focus()
		} else {
			m.textInputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}
