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
					m.pushView(StatsView)
				} else {
					m.progressPopup.ShowWithSteps("Loading Worklogs", []string{
						"Fetching worklogs from JIRA",
						"Processing tickets",
						"Calculating statistics",
					})
					return m.loadWorklogsWithProgress("Loading worklogs...")
				}
			case "⏰ Log Time":
				m.progressPopup.Show("Loading Time Logging", "Preparing time logging interface...")
				return m.loadTimeLoggingWithProgress()
			case "🎫 Create Ticket":
				m.progressPopup.Show("Loading Ticket Creation", "Preparing ticket creation form...")
				return m.loadTicketCreationWithProgress()
			case "📅 Filter Period":
				m.progressPopup.Show("Loading Period Selection", "Loading period options...")
				return m.loadPeriodSelectionWithProgress()
			case "🔄 Refresh":
				m.progressPopup.ShowWithSteps("Refreshing Data", []string{
					"Fetching latest from JIRA",
					"Processing worklogs",
					"Updating statistics",
				})
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
			m.currentView = StatsView // Go to stats view
			m.progressPopup.ShowWithSteps("Loading Period Data", []string{
				"Fetching worklogs for selected period",
				"Processing data",
				"Calculating statistics",
			})
			return m.loadWorklogsWithProgress("Loading data for new period...")
		}
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
