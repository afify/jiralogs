package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func (m *AppModel) renderLoadingView() string {
	var content []string

	content = append(content, m.spinner.View()+" "+InfoStyle.Render(m.loadingMsg))
	content = append(content, "")

	progressWidth := 40
	if m.width > 60 {
		progressWidth = m.width - 20
	}
	bar := m.progressBar.ViewAs(m.progressPercent)
	content = append(content, bar)
	content = append(content, "")

	percentText := fmt.Sprintf("%.0f%%", m.progressPercent*100)
	content = append(content, ProgressTextStyle.Render(percentText))

	if len(m.loadingSteps) > 0 {
		content = append(content, "")
		content = append(content, HelpStyle.Render("Steps:"))
		for i, step := range m.loadingSteps {
			prefix := "  ○ "
			style := DescStyle
			if i < m.currentStep {
				prefix = "  ✓ "
				style = SuccessStyle
			} else if i == m.currentStep {
				prefix = "  ▶ "
				style = InfoStyle
			}
			content = append(content, style.Render(prefix+step))
		}
	}

	box := DialogStyle.Width(progressWidth + 10).Render(
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)

	return ContentStyle.Render(box)
}

func (m *AppModel) renderErrorView() string {
	var content []string

	content = append(content, ErrorStyle.Render("⚠ Error Occurred"))
	content = append(content, "")

	errMsg := fmt.Sprintf("%v", m.error)
	content = append(content, lipgloss.NewStyle().Width(50).Render(errMsg))

	if m.errorDetails != "" {
		content = append(content, "")
		content = append(content, HelpStyle.Render("Details:"))
		content = append(content, DescStyle.Width(50).Render(m.errorDetails))
	}

	content = append(content, "")
	content = append(content, InfoStyle.Render("Suggestions:"))
	content = append(content, "  • Check your JIRA credentials")
	content = append(content, "  • Verify network connection")
	content = append(content, "  • Ensure JIRA_BASE_URL is correct")

	content = append(content, "")
	content = append(content, HelpStyle.Render("Press ESC to go back or R to retry"))

	errorBox := DialogStyle.
		BorderForeground(Red).
		Width(60).
		Render(lipgloss.JoinVertical(lipgloss.Left, content...))

	return ContentStyle.Render(errorBox)
}

func (m *AppModel) renderMainMenuView() string {
	return ContentStyle.Render(m.mainMenu.View())
}

func (m *AppModel) renderPeriodSelectionView() string {
	return ContentStyle.Render(m.periodList.View())
}

func (m *AppModel) renderWorklogDisplayView() string {
	frameH, _ := ContentStyle.GetFrameSize()
	availableWidth := m.width - frameH
	if availableWidth < 80 {
		availableWidth = 80
	}

	summaryWidth := availableWidth - 54
	if summaryWidth < 20 {
		summaryWidth = 20
	}

	columns := []table.Column{
		{Title: "Ticket", Width: 12},
		{Title: "Summary", Width: summaryWidth},
		{Title: "Hours", Width: 10},
		{Title: "Days", Width: 8},
		{Title: "Status", Width: 12},
		{Title: "Last Log", Width: 12},
	}

	var rows []table.Row

	if len(m.ticketLogs) > 0 {
		for key, ticket := range m.ticketLogs {
			ticketKey := key
			summary := "N/A"
			hours := 0.0
			lastLog := "N/A"

			if ticket.Summary != "" {
				maxLen := summaryWidth - 3
				if len(ticket.Summary) > maxLen {
					summary = ticket.Summary[:maxLen] + "..."
				} else {
					summary = ticket.Summary
				}
			}
			hours = ticket.Total
			lastLog = ticket.LastLog

			days := int(hours / 8)
			status := "✓ Complete"
			if hours == 0 {
				status = "⚠ No logs"
			} else if hours < 40 {
				status = "◐ Partial"
			}

			rows = append(rows, table.Row{
				ticketKey,
				summary,
				fmt.Sprintf("%.1f", hours),
				fmt.Sprintf("%d", days),
				status,
				lastLog,
			})
		}
	} else {
		rows = []table.Row{
			{"PROJ-123", "Fix authentication bug in login system", "32.0", "4", "✓ Complete", "2024-01-15"},
			{"PROJ-124", "Implement new user dashboard features", "24.0", "3", "◐ Partial", "2024-01-14"},
			{"PROJ-125", "Update API documentation for v2.0", "16.0", "2", "◐ Partial", "2024-01-13"},
			{"PROJ-126", "Code review and refactoring tasks", "8.0", "1", "◐ Partial", "2024-01-12"},
			{"PROJ-127", "Database migration script development", "0.0", "0", "⚠ Missing", "N/A"},
		}
	}

	_, frameV := ContentStyle.GetFrameSize()
	tableHeight := m.height - frameV - 15
	if tableHeight < 5 {
		tableHeight = 5
	}
	if tableHeight > 20 {
		tableHeight = 20
	}

	m.ticketTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Cyan).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(Orange).
		Background(BgLight).
		Bold(true)
	m.ticketTable.SetStyles(s)

	var summaryLines []string
	if m.summary != nil {
		summaryLines = append(summaryLines, HeaderStyle.Render("WORKLOG SUMMARY"))
		summaryLines = append(summaryLines, "")
		summaryLines = append(summaryLines, fmt.Sprintf("Period: %s to %s", m.summary.StartDate, m.summary.EndDate))
		summaryLines = append(summaryLines, fmt.Sprintf("Logged: %d/%d days (%.1f/%.0f hours)",
			m.summary.LoggedDays, m.summary.Workdays, m.summary.LoggedHours, m.summary.RequiredHours))

		progress := m.summary.LoggedHours / m.summary.RequiredHours
		if progress > 1.0 {
			progress = 1.0
		}
		progressBar := m.progressBar.ViewAs(progress)
		summaryLines = append(summaryLines, progressBar)

		if m.summary.LoggedDays == m.summary.Workdays {
			summaryLines = append(summaryLines, "")
			summaryLines = append(summaryLines, SuccessStyle.Render("✓ All days are logged!"))
		} else {
			summaryLines = append(summaryLines, "")
			missingDays := m.summary.Workdays - m.summary.LoggedDays
			summaryLines = append(summaryLines, WarningStyle.Render(fmt.Sprintf("⚠ %d days missing", missingDays)))
		}
	} else {
		summaryLines = append(summaryLines, HeaderStyle.Render("WORKLOG SUMMARY"))
		summaryLines = append(summaryLines, "")
		summaryLines = append(summaryLines, InfoStyle.Render("Loading worklog data..."))
	}

	summary := lipgloss.JoinVertical(lipgloss.Left, summaryLines...)

	tableTitle := ListTitleStyle.Render("Ticket Details:")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		summary,
		"",
		tableTitle,
		m.ticketTable.View(),
	)

	return ContentStyle.Render(content)
}

func (m *AppModel) renderTimeLoggingView() string {
	return ContentStyle.Render("Time logging form will be shown here")
}

func (m *AppModel) renderTicketCreationView() string {
	return ContentStyle.Render("Ticket creation form will be shown here")
}

func (m *AppModel) renderHelpLine() string {
	var help string
	switch m.currentView {
	case MainMenuView:
		help = "↑/↓: Navigate • Enter: Select • Ctrl+C: Quit"
	case PeriodSelectionView:
		help = "↑/↓: Navigate • Enter: Select • ESC: Back • Ctrl+C: Quit"
	case WorklogDisplayView:
		help = "L: Log Time • C: Create Ticket • R: Refresh • ESC: Back • Ctrl+C: Quit"
	case TimeLoggingView, TicketCreationView:
		help = "Tab: Next Field • Shift+Tab: Previous • Enter: Submit • ESC: Cancel"
	default:
		help = "ESC: Back • Ctrl+C: Quit • ?: Help"
	}

	return HelpStyle.Width(m.width).Align(lipgloss.Center).Render(help)
}

func (m *AppModel) renderHeaderInfo() string {
	now := time.Now()

	zone, offset := now.Zone()
	offsetHours := offset / 3600
	offsetMins := (offset % 3600) / 60

	dateTime := now.Format("2006-01-02 15:04:05")
	timezone := fmt.Sprintf("%s (UTC%+d:%02d)", zone, offsetHours, offsetMins)

	var userEmail, userID string
	if m.currentUser != nil {
		userEmail = m.currentUser.EmailAddress
		userID = m.currentUser.AccountID

		if userEmail == "" {
			userEmail = m.getEmailFromConfig()
		}
	}

	if userEmail == "" {
		userEmail = "Unknown"
	}
	if userID == "" {
		userID = "Unknown"
	}

	var headerLines []string
	headerLines = append(headerLines, InfoStyle.Render("📧 "+userEmail))
	headerLines = append(headerLines, InfoStyle.Render("🆔 "+userID))
	headerLines = append(headerLines, InfoStyle.Render("📅 "+dateTime))
	headerLines = append(headerLines, InfoStyle.Render("🌍 "+timezone))
	headerLines = append(headerLines, "")

	periodStart := m.period.Start.Format("2006-01-02")
	periodEnd := m.period.End.Format("2006-01-02")
	headerLines = append(headerLines, WarningStyle.Render("📊 Period: "+periodStart+" to "+periodEnd))

	headerContent := lipgloss.JoinVertical(lipgloss.Right, headerLines...)

	return lipgloss.NewStyle().
		Align(lipgloss.Right).
		Render(headerContent)
}
