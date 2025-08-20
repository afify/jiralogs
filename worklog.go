package main

import (
	"fmt"
	"sort"
	"time"
)

type WorklogService struct {
	client *JiraClient
	user   *User
}

func NewWorklogService(client *JiraClient) (*WorklogService, error) {
	user, err := client.GetCurrentUser()
	if err != nil {
		return nil, fmt.Errorf("initializing worklog service: %w", err)
	}

	return &WorklogService{
		client: client,
		user:   user,
	}, nil
}

func (w *WorklogService) FetchWorklogs(period Period) (map[string]*TicketWorklog, map[string]float64, map[string][]string, error) {
	startStr := period.Start.Format("2006-01-02")
	endStr := period.End.Format("2006-01-02")

	// Search for issues assigned to me OR issues I've logged work to
	jql := fmt.Sprintf("(assignee = currentUser() AND status != Closed) OR (worklogAuthor = '%s' AND worklogDate >= '%s' AND worklogDate <= '%s')",
		Email, startStr, endStr)

	issues, err := w.client.SearchIssues(jql)
	if err != nil {
		return nil, nil, nil, err
	}

	ticketLogs := make(map[string]*TicketWorklog)
	dailyHours := make(map[string]float64)
	dailyTickets := make(map[string][]string)

	// Process each issue
	for _, issue := range issues {
		ticketLog, err := w.processIssueWorklogs(issue, period)
		if err != nil {
			continue // Skip issues with errors
		}

		if ticketLog != nil {
			// Include ticket even if it has no worklogs (for assigned tickets)
			ticketLogs[issue.Key] = ticketLog

			// Update daily aggregates only if there are logs
			if len(ticketLog.Logs) > 0 {
				for _, log := range ticketLog.Logs {
					dailyHours[log.Date] += log.Hours
					if !contains(dailyTickets[log.Date], issue.Key) {
						dailyTickets[log.Date] = append(dailyTickets[log.Date], issue.Key)
					}
				}
			}
		}
	}

	return ticketLogs, dailyHours, dailyTickets, nil
}

func (w *WorklogService) processIssueWorklogs(issue Issue, period Period) (*TicketWorklog, error) {
	worklogs, err := w.client.GetIssueWorklogs(issue.Key)
	if err != nil {
		return nil, err
	}

	ticketLog := &TicketWorklog{
		Key:     issue.Key,
		Summary: issue.Fields.Summary,
		Logs:    []DayLog{},
		Total:   0,
	}

	startStr := period.Start.Format("2006-01-02")
	endStr := period.End.Format("2006-01-02")

	for _, wl := range worklogs {
		if wl.Author.AccountID != w.user.AccountID {
			continue
		}

		// Parse Jira date format: "2025-08-14T11:13:21.924+0300"
		t, err := time.Parse("2006-01-02T15:04:05.000-0700", wl.Started)
		if err != nil {
			continue
		}

		date := t.Format("2006-01-02")
		if date < startStr || date > endStr {
			continue
		}

		hours := float64(wl.TimeSpentSeconds) / 3600
		ticketLog.Logs = append(ticketLog.Logs, DayLog{
			Date:    date,
			Hours:   hours,
			Comment: wl.Comment,
		})
		ticketLog.Total += hours
	}

	// Sort logs by date
	sort.Slice(ticketLog.Logs, func(i, j int) bool {
		return ticketLog.Logs[i].Date < ticketLog.Logs[j].Date
	})

	return ticketLog, nil
}

func (w *WorklogService) GenerateDailyReports(period Period, dailyHours map[string]float64, dailyTickets map[string][]string) []DailyReport {
	var reports []DailyReport

	for d := period.Start; !d.After(period.End); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		day := d.Weekday().String()[:3] // Use full weekday name, take first 3 chars

		report := DailyReport{
			Date:    date,
			Day:     day,
			Hours:   dailyHours[date],
			Tickets: dailyTickets[date],
		}

		if IsWeekend(d) {
			report.Status = DayWeekend
		} else if report.Hours >= RequiredHoursPerDay {
			report.Status = DayCompliant
		} else if report.Hours > 0 {
			report.Status = DayPartial
		} else {
			report.Status = DayMissing
		}

		reports = append(reports, report)
	}

	return reports
}

func (w *WorklogService) GenerateSummary(period Period, reports []DailyReport, ticketLogs map[string]*TicketWorklog) Summary {
	summary := Summary{
		StartDate:    period.Start.Format("2006-01-02"),
		EndDate:      period.End.Format("2006-01-02"),
		TotalTickets: len(ticketLogs),
	}

	for _, report := range reports {
		if report.Status == DayWeekend {
			continue
		}

		summary.Workdays++
		summary.LoggedHours += report.Hours

		switch report.Status {
		case DayCompliant:
			summary.CompliantDays = append(summary.CompliantDays, report.Date)
			summary.LoggedDays++
		case DayPartial:
			summary.PartialDays = append(summary.PartialDays, report.Date)
			summary.LoggedDays++
		case DayMissing:
			summary.MissingDays = append(summary.MissingDays, report.Date)
		}
	}

	summary.RequiredHours = float64(summary.Workdays) * RequiredHoursPerDay

	return summary
}

func (w *WorklogService) LogMissingDays(ticketKey string, dates []string, comment string) error {
	for _, date := range dates {
		err := w.client.AddWorklog(ticketKey, date, RequiredHoursPerDay, comment)
		if err != nil {
			return fmt.Errorf("logging %s: %w", date, err)
		}
		fmt.Printf("✓ Logged %s to %s\n", date, ticketKey)
	}
	return nil
}

func (w *WorklogService) LogDistribution(distribution map[string][]string, comment string) error {
	for ticket, dates := range distribution {
		fmt.Printf("\nLogging %d days to %s...\n", len(dates), ticket)
		err := w.LogMissingDays(ticket, dates, comment)
		if err != nil {
			return err
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
