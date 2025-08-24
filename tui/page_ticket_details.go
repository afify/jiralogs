package tui

import (
	"fmt"

	"LogS/shared"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// RenderTicketDetailsPage renders the ticket details table
func (m *AppModel) RenderTicketDetailsPage() string {
	// Wait for terminal size
	if m.width == 0 || m.height == 0 {
		return "\n  Loading ticket details..."
	}

	var content []string

	// Summary section
	if m.summary != nil {
		summaryBox := m.renderTicketSummary()
		content = append(content, summaryBox)
		content = append(content, "")
	}

	// Ticket table
	ticketTable := m.renderTicketTable()
	content = append(content, ticketTable)

	// Join all content
	fullContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	// Apply padding
	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(fullContent)
}

// renderTicketSummary renders a summary box for ticket details
func (m *AppModel) renderTicketSummary() string {
	if m.summary == nil {
		return InfoStyle.Render("Loading ticket data...")
	}

	// Create summary box
	summaryStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(shared.PrimaryColor).
		Padding(1, 2).
		Width(100)

	var lines []string
	lines = append(lines, lipgloss.NewStyle().
		Bold(true).
		Foreground(shared.PrimaryColor).
		Render("📊 TICKET SUMMARY"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Period: %s to %s", m.summary.StartDate, m.summary.EndDate))
	lines = append(lines, fmt.Sprintf("Total Tickets: %d", m.summary.TotalTickets))
	lines = append(lines, fmt.Sprintf("Total Hours Logged: %.1f hours", m.summary.LoggedHours))

	summaryContent := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return summaryStyle.Render(summaryContent)
}

// renderTicketTable renders the scrollable ticket table
func (m *AppModel) renderTicketTable() string {
	// Calculate table dimensions
	availableWidth := m.width - 8
	if availableWidth < 100 {
		availableWidth = 100
	}
	summaryWidth := availableWidth - 60

	// Create table with modern styling
	columns := []table.Column{
		{Title: "TICKET", Width: 12},
		{Title: "SUMMARY", Width: summaryWidth},
		{Title: "HOURS", Width: 10},
		{Title: "DAYS", Width: 8},
		{Title: "STATUS", Width: 15},
		{Title: "LAST LOG", Width: 12},
	}

	var rows []table.Row
	if len(m.ticketLogs) > 0 {
		for key, ticket := range m.ticketLogs {
			summary := ticket.Summary
			if len(summary) > summaryWidth-3 {
				summary = summary[:summaryWidth-3] + "..."
			}

			hours := ticket.Total
			days := int(hours / 8)

			// Status with emoji and color
			status := "✓ Complete"
			if hours == 0 {
				status = "⚠ No logs"
			} else if hours < 40 {
				status = "◐ Partial"
			}

			rows = append(rows, table.Row{
				key,
				summary,
				fmt.Sprintf("%.1f", hours),
				fmt.Sprintf("%d", days),
				status,
				ticket.LastLog,
			})
		}
	} else {
		// Sample data when no tickets loaded
		rows = []table.Row{
			{"PROJ-123", "Fix authentication bug in login system", "32.0", "4", "✓ Complete", "2024-01-15"},
			{"PROJ-124", "Implement new user dashboard features", "24.0", "3", "◐ Partial", "2024-01-14"},
			{"PROJ-125", "Update API documentation for v2.0", "16.0", "2", "◐ Partial", "2024-01-13"},
			{"PROJ-126", "Code review and refactoring tasks", "8.0", "1", "◐ Partial", "2024-01-12"},
			{"PROJ-127", "Database migration script development", "0.0", "0", "⚠ No logs", "N/A"},
		}
	}

	// Calculate table height - table should have scrolling
	// Reserve space for summary box, padding, and help line
	reservedHeight := 15
	tableHeight := m.height - reservedHeight
	if tableHeight < 10 {
		tableHeight = 10
	}
	if tableHeight > 30 {
		tableHeight = 30
	}

	// Create the table with scrolling support
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Apply modern table styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(shared.GradientStart).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Background(lipgloss.Color("#1a1a2e"))

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFF")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true)

	s.Cell = s.Cell.
		Foreground(lipgloss.Color("#FAFAFA"))

	t.SetStyles(s)

	// Store the table for keyboard navigation
	m.ticketTable = t

	// Create a bordered box for the table
	tableBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#383838")).
		Padding(1).
		Render(t.View())

	return lipgloss.JoinVertical(lipgloss.Left,
		sectionTitleStyle.Render("📋 Ticket Details (↑/↓ to scroll)"),
		tableBox,
	)
}

// HandleTicketDetailsInput handles keyboard input for the ticket details page
func (m *AppModel) HandleTicketDetailsInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch msg.String() {
	case "q", "Q":
		// Quit the application
		return tea.Quit

	case "h", "H":
		// Return to home (stats page)
		m.currentView = StatsView
		return nil

	case "m", "M":
		// Go to main menu
		m.currentView = MainMenuView
		return nil

	case "p", "P":
		// Change period
		m.currentView = PeriodSelectionView
		return nil

	case "r", "R":
		// Refresh data
		m.progressPopup.ShowWithSteps("🔄 Refreshing Tickets", []string{
			"Fetching latest data from JIRA",
			"Processing tickets",
			"Updating table",
		})
		return LoadWorklogsWithProgress(m.service, m.period)

	case "esc":
		// Go back to stats page
		m.currentView = StatsView
		return nil

	default:
		// Let the table handle navigation keys
		m.ticketTable, cmd = m.ticketTable.Update(msg)
	}

	return cmd
}
