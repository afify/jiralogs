package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Reporter struct {
	service *WorklogService
}

func NewReporter(format string, service *WorklogService) *Reporter {
	return &Reporter{service: service}
}

func (r *Reporter) GenerateReport(ticketLogs map[string]*TicketWorklog, reports []DailyReport, summary Summary) {
	r.printCompactReport(ticketLogs, summary)

	if len(summary.MissingDays) > 0 && len(ticketLogs) > 0 {
		r.interactiveLogging(ticketLogs, summary)
	} else if len(summary.MissingDays) == 0 {
		fmt.Println("\n✓ All days are logged!")
	}
}

func (r *Reporter) printCompactReport(ticketLogs map[string]*TicketWorklog, summary Summary) {
	fmt.Println("==============================================")
	fmt.Println("JIRA WORKLOG SUMMARY")
	fmt.Println("==============================================")

	fmt.Printf("Period: %s to %s\n", summary.StartDate, summary.EndDate)
	fmt.Printf("Logged: %d/%d days (%.1f/%d hours)\n",
		summary.LoggedDays, summary.Workdays, summary.LoggedHours, int(summary.RequiredHours))

	// Show available tickets
	fmt.Printf("\nAvailable Tickets (%d):\n", len(ticketLogs))
	keys := make([]string, 0, len(ticketLogs))
	for k := range ticketLogs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		ticket := ticketLogs[key]
		days := int(ticket.Total / 8)
		if ticket.Total > 0 {
			fmt.Printf("  %d. %s: %s (%.1fh / %d days logged)\n", i+1, ticket.Key, ticket.Summary, ticket.Total, days)
		} else {
			fmt.Printf("  %d. %s: %s (no time logged)\n", i+1, ticket.Key, ticket.Summary)
		}
	}

	// Show missing days as array
	if len(summary.MissingDays) > 0 {
		fmt.Printf("\nMissing Days (%d): [", len(summary.MissingDays))
		for i, date := range summary.MissingDays {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("'%s'", date)
		}
		fmt.Println("]")
	}
}

func (r *Reporter) interactiveLogging(ticketLogs map[string]*TicketWorklog, summary Summary) {
	reader := bufio.NewReader(os.Stdin)

	// Sort tickets
	keys := make([]string, 0, len(ticketLogs))
	for k := range ticketLogs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for {
		fmt.Println("\n==============================================")
		fmt.Println("WORKLOG GENERATOR")
		fmt.Println("==============================================")
		fmt.Printf("You have %d missing days to log across %d tickets.\n\n", len(summary.MissingDays), len(ticketLogs))

		fmt.Println("Choose an option:")
		fmt.Println("1. Log missing days - even distribution across all tickets")
		fmt.Println("2. Log missing days - all to single ticket")
		fmt.Println("3. Log missing days - custom distribution")
		fmt.Println("4. Log single day (8h) to specific ticket")
		fmt.Println("5. Show ticket list with logged hours")
		fmt.Println("6. Show distribution suggestions")
		fmt.Println("q. Quit")
		fmt.Print("\nChoice (1-6 or q): ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		if choice == "q" || choice == "quit" {
			fmt.Println("\nExiting...")
			break
		}

		switch choice {
		case "1":
			// Execute even distribution
			r.executeEvenDistribution(keys, summary.MissingDays)
		case "2":
			// Execute single ticket
			fmt.Println("\nSelect ticket:")
			for i, key := range keys {
				ticket := ticketLogs[key]
				fmt.Printf("  %d. %s: %s\n", i+1, ticket.Key, ticket.Summary)
			}
			fmt.Print("\nTicket number: ")
			ticketNum, _ := reader.ReadString('\n')
			ticketNum = strings.TrimSpace(ticketNum)
			idx, _ := strconv.Atoi(ticketNum)
			if idx > 0 && idx <= len(keys) {
				r.executeSingleTicket(keys[idx-1], summary.MissingDays)
			}
		case "3":
			// Custom distribution
			r.executeCustomDistribution(keys, ticketLogs, summary.MissingDays, reader)
		case "4":
			// Log single day to specific ticket
			r.logSingleDay(keys, ticketLogs, reader)
		case "5":
			// Show ticket list with logged hours
			r.showTicketList(keys, ticketLogs)
		case "6":
			r.showSuggestions(keys, summary.MissingDays)
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func (r *Reporter) executeCustomDistribution(tickets []string, ticketLogs map[string]*TicketWorklog, missingDays []string, reader *bufio.Reader) {
	fmt.Println("\n==============================================")
	fmt.Println("CUSTOM DISTRIBUTION")
	fmt.Println("==============================================")
	fmt.Println("For each ticket, enter the number of days to log (0 to skip):")

	distribution := make(map[string][]string)
	totalAssigned := 0
	dayIndex := 0

	for i, ticket := range tickets {
		summaryPreview := ticketLogs[ticket].Summary
		if len(summaryPreview) > 50 {
			summaryPreview = summaryPreview[:50] + "..."
		}
		fmt.Printf("%d. %s (%s): ", i+1, ticket, summaryPreview)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		days, _ := strconv.Atoi(input)

		if days > 0 {
			var ticketDays []string
			for j := 0; j < days && dayIndex < len(missingDays); j++ {
				ticketDays = append(ticketDays, missingDays[dayIndex])
				dayIndex++
				totalAssigned++
			}
			if len(ticketDays) > 0 {
				distribution[ticket] = ticketDays
			}
		}
	}

	if totalAssigned == 0 {
		fmt.Println("No days assigned. Cancelled.")
		return
	}

	// Confirm before executing
	fmt.Printf("\nAbout to log %d days across %d tickets. Continue? (y/n): ", totalAssigned, len(distribution))
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Cancelled")
		return
	}

	// Execute the logging
	err := r.service.LogDistribution(distribution, "")
	if err != nil {
		fmt.Printf("\n❌ Error: %v\n", err)
	} else {
		fmt.Printf("\n✅ Successfully logged %d days!\n", totalAssigned)
	}
}

func (r *Reporter) executeEvenDistribution(tickets []string, missingDays []string) {
	fmt.Println("\n==============================================")
	fmt.Println("EXECUTING EVEN DISTRIBUTION")
	fmt.Println("==============================================")

	daysPerTicket := len(missingDays) / len(tickets)
	remainder := len(missingDays) % len(tickets)
	dayIndex := 0

	distribution := make(map[string][]string)

	for i, ticket := range tickets {
		days := daysPerTicket
		if i < remainder {
			days++
		}

		if days == 0 {
			continue
		}

		var ticketDays []string
		for j := 0; j < days && dayIndex < len(missingDays); j++ {
			ticketDays = append(ticketDays, missingDays[dayIndex])
			dayIndex++
		}
		distribution[ticket] = ticketDays
	}

	// Execute the logging
	err := r.service.LogDistribution(distribution, "")
	if err != nil {
		fmt.Printf("\n❌ Error: %v\n", err)
	} else {
		fmt.Printf("\n✅ Successfully logged %d days!\n", len(missingDays))
	}
}

func (r *Reporter) executeSingleTicket(ticket string, missingDays []string) {
	fmt.Printf("\n==============================================\n")
	fmt.Printf("LOGGING %d DAYS TO %s\n", len(missingDays), ticket)
	fmt.Printf("==============================================\n")

	err := r.service.LogMissingDays(ticket, missingDays, "")
	if err != nil {
		fmt.Printf("\n❌ Error: %v\n", err)
	} else {
		fmt.Printf("\n✅ Successfully logged %d days to %s!\n", len(missingDays), ticket)
	}
}

func (r *Reporter) logSingleDay(tickets []string, ticketLogs map[string]*TicketWorklog, reader *bufio.Reader) {
	fmt.Println("\n==============================================")
	fmt.Println("LOG SINGLE DAY (8 HOURS)")
	fmt.Println("==============================================")

	// Select ticket
	fmt.Println("\nSelect ticket:")
	for i, key := range tickets {
		ticket := ticketLogs[key]
		fmt.Printf("  %d. %s: %s\n", i+1, ticket.Key, ticket.Summary)
	}
	fmt.Print("\nTicket number: ")
	ticketNum, _ := reader.ReadString('\n')
	ticketNum = strings.TrimSpace(ticketNum)
	idx, _ := strconv.Atoi(ticketNum)

	if idx < 1 || idx > len(tickets) {
		fmt.Println("Invalid ticket number")
		return
	}

	selectedTicket := tickets[idx-1]

	// Get date
	fmt.Print("\nEnter date (YYYY-MM-DD) or press Enter for today: ")
	dateInput, _ := reader.ReadString('\n')
	dateInput = strings.TrimSpace(dateInput)

	var date string
	if dateInput == "" {
		// Use today's date
		date = time.Now().Format("2006-01-02")
	} else {
		// Validate date format
		_, err := time.Parse("2006-01-02", dateInput)
		if err != nil {
			fmt.Printf("Invalid date format: %s\n", dateInput)
			return
		}
		date = dateInput
	}

	// Get optional comment
	fmt.Print("\nEnter comment (optional): ")
	comment, _ := reader.ReadString('\n')
	comment = strings.TrimSpace(comment)

	// Confirm
	fmt.Printf("\nAbout to log:\n")
	fmt.Printf("  Ticket: %s\n", selectedTicket)
	fmt.Printf("  Date: %s\n", date)
	fmt.Printf("  Hours: 8 (1 day)\n")
	if comment != "" {
		fmt.Printf("  Comment: %s\n", comment)
	}
	fmt.Print("\nConfirm? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Cancelled")
		return
	}

	// Execute logging
	fmt.Printf("\nLogging to %s...\n", selectedTicket)
	err := r.service.client.AddWorklog(selectedTicket, date, 8.0, comment)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully logged 8 hours to %s for %s!\n", selectedTicket, date)
	}
}

func (r *Reporter) showTicketList(tickets []string, ticketLogs map[string]*TicketWorklog) {
	fmt.Println("\n==============================================")
	fmt.Println("TICKET LIST WITH LOGGED TIME")
	fmt.Println("==============================================")

	totalHours := 0.0
	totalDays := 0

	for i, key := range tickets {
		ticket := ticketLogs[key]
		days := int(ticket.Total / 8)
		hours := ticket.Total
		totalHours += hours
		totalDays += days

		fmt.Printf("%2d. %s\n", i+1, key)
		fmt.Printf("    %s\n", ticket.Summary)
		if hours > 0 {
			fmt.Printf("    Logged: %.1f hours (%d days)\n", hours, days)
		} else {
			fmt.Printf("    Logged: No time logged yet\n")
		}
		fmt.Println()
	}

	fmt.Println("----------------------------------------------")
	fmt.Printf("Total: %d tickets, %.1f hours (%d days) logged\n", len(tickets), totalHours, totalDays)
}

func (r *Reporter) showSuggestions(tickets []string, missingDays []string) {
	fmt.Println("\n==============================================")
	fmt.Println("SUGGESTED DISTRIBUTION")
	fmt.Println("==============================================")

	daysPerTicket := len(missingDays) / len(tickets)
	remainder := len(missingDays) % len(tickets)

	for i, ticket := range tickets {
		days := daysPerTicket
		if i < remainder {
			days++
		}
		if days > 0 {
			fmt.Printf("  %s: %d days\n", ticket, days)
		}
	}
}
