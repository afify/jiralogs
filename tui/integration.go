package tui

import (
	"fmt"
	"time"

	"LogS/shared"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadWorklogsWithProgress(service WorklogServiceInterface, period shared.Period) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(200 * time.Millisecond)

		ticketLogs, dailyHours, dailyTickets, err := service.FetchWorklogs(period)
		if err != nil {
			return errMsg{
				err:     err,
				details: fmt.Sprintf("Failed to fetch worklogs: %v", err),
			}
		}

		reports := service.GenerateDailyReports(period, dailyHours, dailyTickets)
		summary := service.GenerateSummary(period, reports, ticketLogs)

		ticketLogsData := make(map[string]*shared.TicketLogData)
		for k, v := range ticketLogs {
			lastLog := "Never"
			if len(v.Logs) > 0 {
				lastLog = v.Logs[len(v.Logs)-1].Date
			}

			ticketLogsData[k] = &shared.TicketLogData{
				Summary: v.Summary,
				Total:   v.Total,
				LastLog: lastLog,
				Logs:    v.Logs,
			}
		}

		summaryData := &shared.SummaryData{
			StartDate:     summary.StartDate,
			EndDate:       summary.EndDate,
			Workdays:      summary.Workdays,
			LoggedDays:    summary.LoggedDays,
			RequiredHours: summary.RequiredHours,
			LoggedHours:   summary.LoggedHours,
			CompliantDays: summary.CompliantDays,
			PartialDays:   summary.PartialDays,
			MissingDays:   summary.MissingDays,
			TotalTickets:  summary.TotalTickets,
		}

		return ticketsLoadedMsg{
			ticketLogs: ticketLogsData,
			summary:    summaryData,
		}
	}
}

func CreateTicketWithProgress(client JiraClientInterface, projectKey, summary, description, issueType string) tea.Cmd {
	return func() tea.Msg {
		ticketKey, err := client.CreateIssue(projectKey, summary, description, issueType)
		if err != nil {
			return errMsg{
				err:     err,
				details: fmt.Sprintf("Failed to create ticket in project %s", projectKey),
			}
		}

		return ticketCreatedMsg{
			ticketKey: ticketKey,
		}
	}
}

func LogTimeWithProgress(client JiraClientInterface, ticketKey string, date string, hours float64, comment string) tea.Cmd {
	return func() tea.Msg {
		err := client.AddWorklog(ticketKey, date, hours, comment)
		if err != nil {
			return errMsg{
				err:     err,
				details: fmt.Sprintf("Failed to log %.1f hours to %s", hours, ticketKey),
			}
		}

		return timeLoggedMsg{
			ticketKey: ticketKey,
			hours:     hours,
			date:      date,
		}
	}
}
