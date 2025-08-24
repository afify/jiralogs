package jira

import (
	"fmt"
	"sort"
	"time"

	"LogS/shared"
)

func (w *WorklogService) FetchWorklogs(period shared.Period) (map[string]*shared.TicketWorklog, map[string]float64, map[string][]string, error) {
	startStr := period.Start.Format(shared.DateFormat)
	endStr := period.End.Format(shared.DateFormat)

	jql := fmt.Sprintf("(assignee = currentUser() AND status != Closed) OR (worklogAuthor = '%s' AND worklogDate >= '%s' AND worklogDate <= '%s')",
		w.client.email, startStr, endStr)

	issues, err := w.client.SearchIssues(jql)
	if err != nil {
		shared.LogError("WorklogService.FetchWorklogs", err)
		return nil, nil, nil, err
	}

	ticketLogs := make(map[string]*shared.TicketWorklog)
	dailyHours := make(map[string]float64)
	dailyTickets := make(map[string][]string)

	for _, issue := range issues {
		ticketLog, err := w.processIssueWorklogs(issue, period)
		if err != nil {
			shared.LogError("WorklogService.FetchWorklogs", err)
			continue
		}

		if ticketLog != nil {
			ticketLogs[issue.Key] = ticketLog

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

func (w *WorklogService) processIssueWorklogs(issue shared.Issue, period shared.Period) (*shared.TicketWorklog, error) {
	worklogs, err := w.client.GetIssueWorklogs(issue.Key)
	if err != nil {
		return nil, err
	}

	ticketLog := &shared.TicketWorklog{
		Key:     issue.Key,
		Summary: issue.Fields.Summary,
		Logs:    []shared.DayLog{},
		Total:   shared.ZeroHours,
	}

	startStr := period.Start.Format(shared.DateFormat)
	endStr := period.End.Format(shared.DateFormat)

	for _, wl := range worklogs {
		if wl.Author.AccountID != w.user.AccountID {
			continue
		}

		t, err := time.Parse(shared.WorklogDateFormat, wl.Started)
		if err != nil {
			continue
		}

		date := t.Format(shared.DateFormat)
		if date < startStr || date > endStr {
			continue
		}

		hours := float64(wl.TimeSpentSeconds) / shared.SecondsPerHour
		ticketLog.Logs = append(ticketLog.Logs, shared.DayLog{
			Date:    date,
			Hours:   hours,
			Comment: wl.Comment,
		})
		ticketLog.Total += hours
	}

	sort.Slice(ticketLog.Logs, func(i, j int) bool {
		return ticketLog.Logs[i].Date < ticketLog.Logs[j].Date
	})

	return ticketLog, nil
}
