package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	now := time.Now()

	fmt.Println("==============================================")
	fmt.Println("SELECT PERIOD")
	fmt.Println("==============================================")
	fmt.Println("1. Year to date (from Jan 1)")
	fmt.Println("2. Last 9 months")
	fmt.Println("3. Last 6 months")
	fmt.Println("4. Last 3 months")
	fmt.Println("5. Month to date (from 1st of current month)")
	fmt.Println("6. Week to date (from Monday)")
	fmt.Print("\nChoice (1-6, default 1): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var start time.Time
	switch choice {
	case "2":
		start = now.AddDate(0, -9, 0)
	case "3":
		start = now.AddDate(0, -6, 0)
	case "4":
		start = now.AddDate(0, -3, 0)
	case "5":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "6":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday
		}
		// Go back to Monday (day 1)
		daysBack := (weekday - 1) % 7
		start = now.AddDate(0, 0, -daysBack)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	default:
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	}

	end := now
	p := Period{Start: start, End: end}

	client := NewJiraClient()

	service, err := NewWorklogService(client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ticketLogs, dailyHours, dailyTickets, err := service.FetchWorklogs(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching worklogs: %v\n", err)
		os.Exit(1)
	}

	reports := service.GenerateDailyReports(p, dailyHours, dailyTickets)

	summary := service.GenerateSummary(p, reports, ticketLogs)

	reporter := NewReporter("detailed", service)
	reporter.GenerateReport(ticketLogs, reports, summary)
}
