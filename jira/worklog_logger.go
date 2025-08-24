package jira

import (
	"LogS/shared"
	"fmt"
)

func (w *WorklogService) LogMissingDays(ticketKey string, dates []string, comment string) error {
	for _, date := range dates {
		err := w.client.AddWorklog(ticketKey, date, shared.RequiredHoursPerDay, comment)
		if err != nil {
			return fmt.Errorf("logging %s: %w", date, err)
		}
		fmt.Printf("✓ Logged %s to %s\n", date, ticketKey)
	}
	return nil
}

func (w *WorklogService) LogDistribution(distribution map[string][]string, comment string) error {
	for ticket, dates := range distribution {
		fmt.Println()
		fmt.Printf("ℹ Logging %d days to %s...\n", len(dates), ticket)
		err := w.LogMissingDays(ticket, dates, comment)
		if err != nil {
			return err
		}
	}
	return nil
}
