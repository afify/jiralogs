package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"LogS/app"
	"LogS/config"
	"LogS/jira"
	"LogS/shared"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

func newKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	return km
}

func main() {
	// Clear terminal
	fmt.Print("\033[2J\033[H")

	// Load configuration
	cfg := config.New()
	if !cfg.IsConfigured() {
		fmt.Println("Error: JIRA configuration not found.")
		fmt.Println("Please set the following environment variables:")
		fmt.Println("  JIRA_BASE_URL")
		fmt.Println("  JIRA_EMAIL")
		fmt.Println("  JIRA_API_TOKEN")
		os.Exit(1)
	}

	// Create app context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n\n👋 Goodbye!")
		cancel()
		os.Exit(0)
	}()

	// Initialize app
	app, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.Shutdown()

	// Get worklog service
	ws := app.WorklogService()
	if ws == nil {
		fmt.Println("Error: Worklog service not available")
		os.Exit(1)
	}

	// Banner
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#7D56F4"}).
		Padding(1, 2).
		Margin(1, 0).
		Width(60).
		Align(lipgloss.Center)

	userInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Italic(true).
		Margin(0, 0, 1, 0)

	fmt.Println(headerStyle.Render("LogS"))
	fmt.Println(userInfoStyle.Render(fmt.Sprintf("%s | %s", cfg.JiraEmail, cfg.JiraBaseURL)))

	// Fetch worklogs once for both dashboard and missing days
	now := time.Now()
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	period := shared.Period{Start: startOfYear, End: now}

	shared.StatusBar("📊 Fetching worklog data from JIRA API")
	ticketLogs, dailyHours, _, fetchErr := ws.FetchWorklogs(period)
	if fetchErr != nil {
		shared.StatusFailed()
		fmt.Printf("Error loading worklog data: %v\n", fetchErr)
		os.Exit(1)
	}
	shared.StatusComplete()

	showDashboard(ws, dailyHours)

	missingDays := findMissingWorklogDays(ws, dailyHours)
	hasMissing := len(missingDays) > 0

	if hasMissing {
		missingStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F77F00")).
			Padding(1, 2).
			Margin(1, 0).
			Width(60)

		missingTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F77F00")).
			Margin(0, 0, 1, 0)

		var daysList []string
		for _, day := range missingDays {
			daysList = append(daysList, fmt.Sprintf("  %s (%s)", day.Format("2006-01-02"), day.Format("Monday")))
		}

		missingContent := fmt.Sprintf("Found %d days without logged time:\n\n%s", len(missingDays), strings.Join(daysList, "\n"))

		fmt.Println(missingTitle.Render("Missing Worklog Days"))
		fmt.Println(missingStyle.Render(missingContent))
	} else {
		fmt.Println("\n✓ All working days have logged time!")
	}

	for {
		options := []huh.Option[string]{
			huh.NewOption("Show this month report", "month"),
			huh.NewOption("Show tickets logged (YTD)", "tickets"),
		}
		if hasMissing {
			options = append(options,
				huh.NewOption("Log missing days to custom tickets", "custom"),
				huh.NewOption("Log missing days to single ticket", "single"),
			)
		}
		options = append(options, huh.NewOption("Exit", "exit"))

		var option string
		err = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do?").
				Options(options...).
				Value(&option),
		)).WithTheme(huh.ThemeCharm()).WithKeyMap(newKeyMap()).Run()

		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Println("Exiting...")
			return
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		switch option {
		case "month":
			showMonthlyReport(ws, dailyHours, ticketLogs)
		case "tickets":
			showTicketsLogged(ticketLogs)
		case "single":
			logToSingleTicket(ws, missingDays)
		case "custom":
			logToCustomTickets(ws, app.JiraClient(), missingDays)
		case "exit":
			return
		}
	}
}

func findMissingWorklogDays(ws *jira.WorklogService, dailyHours map[string]float64) []time.Time {
	var missingDays []time.Time

	endDate := time.Now()
	startDate := time.Date(endDate.Year(), 1, 1, 0, 0, 0, 0, endDate.Location())

	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		if d.After(time.Now()) {
			continue
		}

		if ws.ShouldSkipDay(d) {
			continue
		}

		dateStr := d.Format("2006-01-02")
		if dailyHours[dateStr] == 0 {
			missingDays = append(missingDays, d)
		}
	}

	return missingDays
}

func logToSingleTicket(ws *jira.WorklogService, missingDays []time.Time) {
	var ticket string
	var hours string
	var description string

	// Create form for single ticket logging
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter ticket key").
				Placeholder("PROJ-123").
				Value(&ticket).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("ticket key is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Hours per day").
				Placeholder("8").
				Value(&hours),

			huh.NewText().
				Title("Description (optional)").
				Placeholder("Working on tasks...").
				Value(&description),
		),
	).WithTheme(huh.ThemeCharm()).WithKeyMap(newKeyMap())

	err := form.Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Parse hours (default to 8)
	hoursInt := 8
	if hours != "" {
		h, err := strconv.Atoi(hours)
		if err == nil && h > 0 {
			hoursInt = h
		}
	}

	// Confirm before logging
	var confirm bool
	err = huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(fmt.Sprintf("Log %d hours per day to %s for %d days?", hoursInt, ticket, len(missingDays))).
			Value(&confirm),
	)).WithTheme(huh.ThemeCharm()).WithKeyMap(newKeyMap()).Run()

	if errors.Is(err, huh.ErrUserAborted) || !confirm {
		return
	}

	// Log work for each missing day
	fmt.Printf("\nLogging %d hours per day to %s...\n", hoursInt, ticket)

	successCount := 0
	for _, day := range missingDays {
		fmt.Printf("  Logging %s... ", day.Format("2006-01-02"))
		err := ws.LogWork(ticket, float64(hoursInt), description, day)
		if err != nil {
			fmt.Printf("✗ Error: %v\n", err)
		} else {
			fmt.Println("✓")
			successCount++
		}
	}

	fmt.Printf("\nCompleted: %d/%d days logged successfully\n", successCount, len(missingDays))
}

type DayTicketAssignment struct {
	Date        time.Time
	TicketKey   string
	TicketTitle string
	Hours       int
	Description string
}

func logToCustomTickets(ws *jira.WorklogService, jiraClient *jira.JiraClient, missingDays []time.Time) {
	fmt.Println("\nAssign tickets to missing worklog days")
	fmt.Println("======================================")

	// Fetch available tickets with spinner
	var tickets []shared.Issue
	var err error

	ticketAction := func() {
		tickets, err = ws.GetAllTickets()
	}

	spinnerErr := spinner.New().
		Title("🎫 Loading your tickets...").
		Action(ticketAction).
		Run()

	if spinnerErr != nil || err != nil {
		_ = huh.NewNote().
			Title("❌ Error").
			Description(fmt.Sprintf("Failed to load tickets: %v", err)).
			Run()
		return
	}



	// Create assignments for each day
	assignments := make([]DayTicketAssignment, len(missingDays))
	for i, day := range missingDays {
		assignments[i] = DayTicketAssignment{
			Date:  day,
			Hours: 8, // default
		}
	}

	// Let user assign tickets to each day
	for i := range assignments {
		day := &assignments[i]

		fmt.Printf("\n📅 %s (%s)\n", day.Date.Format("2006-01-02"), day.Date.Format("Monday"))

		// Ask if they want to skip this day
		var skipDay bool
		err := huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("Skip this day?").
				Value(&skipDay),
		)).WithTheme(huh.ThemeCharm()).WithKeyMap(newKeyMap()).Run()

		if errors.Is(err, huh.ErrUserAborted) {
			return
		}
		if err != nil || skipDay {
			day.TicketKey = "SKIP"
			continue
		}

		// Create ticket options
		ticketOptions := []huh.Option[string]{
			huh.NewOption("➕ Create new ticket", "create"),
			huh.NewOption("📝 Enter custom ticket", "custom"),
		}

		for _, ticket := range tickets {
			summary := ticket.Fields.Summary
			if len(summary) > 50 {
				summary = summary[:47] + "..."
			}
			ticketOptions = append(ticketOptions,
				huh.NewOption(fmt.Sprintf("🎫 %s - %s", ticket.Key, summary), ticket.Key))
		}

		// Ticket selection
		var selectedTicket string
		err = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select ticket for this day").
				Description("Type to search through tickets").
				Options(ticketOptions...).
				Filtering(true).
				Height(10).
				Value(&selectedTicket),
		)).WithTheme(huh.ThemeCharm()).WithKeyMap(newKeyMap()).Run()

		if errors.Is(err, huh.ErrUserAborted) {
			return
		}
		if err != nil {
			day.TicketKey = "SKIP"
			continue
		}

		// Handle create new ticket
		if selectedTicket == "create" {
			createdTicket, createdTitle, err := createNewTicket(jiraClient)
			if errors.Is(err, huh.ErrUserAborted) {
				return
			}
			if err != nil {
				fmt.Printf("❌ Error creating ticket: %v\n", err)
				day.TicketKey = "SKIP"
				continue
			}
			selectedTicket = createdTicket
			day.TicketTitle = createdTitle
		}

		// Handle custom ticket
		if selectedTicket == "custom" {
			err = huh.NewForm(huh.NewGroup(
				huh.NewInput().
					Title("Enter ticket key").
					Placeholder("PROJ-123").
					Value(&selectedTicket).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("ticket key is required")
						}
						return nil
					}),
			)).WithTheme(huh.ThemeCharm()).WithKeyMap(newKeyMap()).Run()

			if errors.Is(err, huh.ErrUserAborted) {
				return
			}
			if err != nil {
				day.TicketKey = "SKIP"
				continue
			}
		}

		// Get hours and description
		var hoursStr = "8"
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Hours").
					Value(&hoursStr).
					Validate(func(s string) error {
						h, err := strconv.Atoi(s)
						if err != nil || h <= 0 || h > 12 {
							return fmt.Errorf("hours must be between 1-12")
						}
						return nil
					}),
				huh.NewText().
					Title("Description (optional)").
					Placeholder("Working on tasks...").
					Value(&day.Description),
			),
		).WithKeyMap(newKeyMap())

		err = form.Run()
		if errors.Is(err, huh.ErrUserAborted) {
			return
		}
		if err != nil {
			day.TicketKey = "SKIP"
			continue
		}

		day.TicketKey = selectedTicket
		day.Hours, _ = strconv.Atoi(hoursStr)

		// Find ticket title for display
		for _, ticket := range tickets {
			if ticket.Key == selectedTicket {
				day.TicketTitle = ticket.Fields.Summary
				break
			}
		}
		if day.TicketTitle == "" && selectedTicket != "custom" {
			day.TicketTitle = "Custom ticket"
		}
	}

	// Show summary and confirm
	fmt.Println("\n📋 WORKLOG SUMMARY")
	fmt.Println("==================")

	validAssignments := []DayTicketAssignment{}
	for _, assignment := range assignments {
		if assignment.TicketKey != "SKIP" && assignment.TicketKey != "" {
			validAssignments = append(validAssignments, assignment)

			title := assignment.TicketTitle
			if len(title) > 40 {
				title = title[:37] + "..."
			}

			fmt.Printf("📅 %s  🎫 %s  ⏰ %dh\n",
				assignment.Date.Format("2006-01-02"),
				assignment.TicketKey,
				assignment.Hours)

			if title != "" {
				fmt.Printf("   📝 %s\n", title)
			}
			if assignment.Description != "" {
				fmt.Printf("   💬 %s\n", assignment.Description)
			}
			fmt.Println()
		}
	}

	if len(validAssignments) == 0 {
		fmt.Println("No assignments to log.")
		return
	}

	// Final confirmation
	var confirmLog bool
	err = huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(fmt.Sprintf("Log work for %d days?", len(validAssignments))).
			Value(&confirmLog),
	)).WithTheme(huh.ThemeCharm()).WithKeyMap(newKeyMap()).Run()

	if errors.Is(err, huh.ErrUserAborted) || !confirmLog {
		return
	}

	// Execute logging
	fmt.Println("\n🚀 LOGGING WORK...")
	fmt.Println("==================")

	successCount := 0
	for _, assignment := range validAssignments {
		fmt.Printf("📅 %s ➤ %s (%dh)... ",
			assignment.Date.Format("2006-01-02"),
			assignment.TicketKey,
			assignment.Hours)

		err := ws.LogWork(assignment.TicketKey, float64(assignment.Hours), assignment.Description, assignment.Date)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Println("✅ Success")
			successCount++
		}
	}

	fmt.Printf("\n🎉 Completed: %d/%d days logged successfully\n", successCount, len(validAssignments))
}

func createNewTicket(jiraClient *jira.JiraClient) (string, string, error) {
	projects, err := jiraClient.GetProjects()
	if err != nil {
		return "", "", fmt.Errorf("fetching projects: %w", err)
	}

	if len(projects) == 0 {
		return "", "", fmt.Errorf("no projects available")
	}

	projectOptions := []huh.Option[string]{}
	for _, project := range projects {
		projectOptions = append(projectOptions,
			huh.NewOption(fmt.Sprintf("%s - %s", project.Key, project.Name), project.Key))
	}

	var selectedProject string
	err = huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Select project").
			Options(projectOptions...).
			Filtering(true).
			Height(10).
			Value(&selectedProject),
	)).WithTheme(huh.ThemeCharm()).WithKeyMap(newKeyMap()).Run()

	if err != nil {
		return "", "", err
	}

	var summary string
	var description string
	issueType := "Task"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Ticket summary").
				Placeholder("Brief description of the work").
				Value(&summary).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("summary is required")
					}
					return nil
				}),

			huh.NewText().
				Title("Description (optional)").
				Placeholder("Detailed description...").
				Value(&description),

			huh.NewInput().
				Title("Issue type").
				Placeholder("Task").
				Value(&issueType).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("issue type is required")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm()).WithKeyMap(newKeyMap())

	err = form.Run()
	if err != nil {
		return "", "", err
	}

	fmt.Printf("\n🎫 Creating ticket in %s...\n", selectedProject)
	ticketKey, err := jiraClient.CreateIssue(selectedProject, summary, description, issueType)
	if err != nil {
		return "", "", fmt.Errorf("creating ticket: %w", err)
	}

	fmt.Printf("✅ Created ticket: %s\n", ticketKey)
	return ticketKey, summary, nil
}

func showDashboard(ws *jira.WorklogService, dailyHours map[string]float64) {
	now := time.Now()
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	totalLoggedHours := 0.0
	daysLogged := 0
	workingDays := 0
	leaveDays := 0
	weekendDays := 0

	for d := startOfYear; d.Before(now) || d.Equal(now.Truncate(24*time.Hour)); d = d.AddDate(0, 0, 1) {
		if d.After(now) {
			continue
		}

		if ws.IsWeekend(d) {
			weekendDays++
		} else if ws.IsLeaveDay(d) {
			leaveDays++
		} else {
			workingDays++
			dateStr := d.Format("2006-01-02")
			if dailyHours[dateStr] > 0 {
				daysLogged++
				totalLoggedHours += dailyHours[dateStr]
			}
		}
	}

	requiredHours := float64(workingDays) * 8.0
	percentage := (totalLoggedHours / requiredHours) * 100
	if requiredHours == 0 {
		percentage = 0
	}

	// Beautiful stats display using lipgloss
	var statsColor lipgloss.Color
	var statsTitle string
	if percentage >= 90 {
		statsColor = lipgloss.Color("#43BF6D") // Green
		statsTitle = "Excellent Progress"
	} else if percentage >= 70 {
		statsColor = lipgloss.Color("#FDFF90") // Yellow
		statsTitle = "Good Progress"
	} else if percentage >= 50 {
		statsColor = lipgloss.Color("#F77F00") // Orange
		statsTitle = "Moderate Progress"
	} else {
		statsColor = lipgloss.Color("#FF5F87") // Red
		statsTitle = "Needs Attention"
	}

	statsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(statsColor).
		Padding(1, 2).
		Margin(1, 0).
		Width(60)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(statsColor).
		Margin(0, 0, 1, 0)

	statsContent := fmt.Sprintf(`Days Logged: %d / %d working days
Hours Logged: %.1f hours
Required Hours: %.1f hours
Leave Days: %d days
Weekend Days: %d days
Progress: %.1f%%

%s`, daysLogged, workingDays, totalLoggedHours, requiredHours, leaveDays, weekendDays, percentage, createProgressBar(percentage))

	fmt.Println(titleStyle.Render(fmt.Sprintf("%s - Year to Date Stats", statsTitle)))
	fmt.Println(statsStyle.Render(statsContent))

	fmt.Println() // Add some space
}

func showMonthlyReport(ws *jira.WorklogService, dailyHours map[string]float64, ticketLogs map[string]*shared.TicketWorklog) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	totalLoggedHours := 0.0
	daysLogged := 0
	workingDays := 0
	leaveDays := 0
	weekendDays := 0
	var offDays []time.Time

	for d := startOfMonth; d.Before(now) || d.Equal(now.Truncate(24*time.Hour)); d = d.AddDate(0, 0, 1) {
		if d.After(now) {
			continue
		}

		if ws.IsWeekend(d) {
			weekendDays++
		} else if ws.IsLeaveDay(d) {
			leaveDays++
			offDays = append(offDays, d)
		} else {
			workingDays++
			dateStr := d.Format("2006-01-02")
			if dailyHours[dateStr] > 0 {
				daysLogged++
				totalLoggedHours += dailyHours[dateStr]
			}
		}
	}

	requiredHours := float64(workingDays) * 8.0
	percentage := (totalLoggedHours / requiredHours) * 100
	if requiredHours == 0 {
		percentage = 0
	}

	var statsColor lipgloss.Color
	var statsTitle string
	if percentage >= 90 {
		statsColor = lipgloss.Color("#43BF6D")
		statsTitle = "Excellent Progress"
	} else if percentage >= 70 {
		statsColor = lipgloss.Color("#FDFF90")
		statsTitle = "Good Progress"
	} else if percentage >= 50 {
		statsColor = lipgloss.Color("#F77F00")
		statsTitle = "Moderate Progress"
	} else {
		statsColor = lipgloss.Color("#FF5F87")
		statsTitle = "Needs Attention"
	}

	statsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(statsColor).
		Padding(1, 2).
		Margin(1, 0).
		Width(60)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(statsColor).
		Margin(0, 0, 1, 0)

	statsContent := fmt.Sprintf(`Days Logged: %d / %d working days
Hours Logged: %.1f hours
Required Hours: %.1f hours
Leave Days: %d days
Weekend Days: %d days
Progress: %.1f%%`, daysLogged, workingDays, totalLoggedHours, requiredHours, leaveDays, weekendDays, percentage)

	if len(offDays) > 0 {
		var offLines []string
		for _, d := range offDays {
			offLines = append(offLines, fmt.Sprintf("  %s (%s)", d.Format("2006-01-02"), d.Format("Monday")))
		}
		statsContent += "\n\nOff Days:\n" + strings.Join(offLines, "\n")
	}

	statsContent += "\n\n" + createProgressBar(percentage)

	fmt.Println(titleStyle.Render(fmt.Sprintf("%s - %s Report", statsTitle, now.Format("January 2006"))))
	fmt.Println(statsStyle.Render(statsContent))

	startStr := startOfMonth.Format(shared.DateFormat)
	endStr := now.Format(shared.DateFormat)

	var rows []ticketSummary
	for _, t := range ticketLogs {
		if t == nil {
			continue
		}
		monthHours := 0.0
		dateSet := make(map[string]time.Time)
		for _, wl := range t.Logs {
			started, err := time.Parse(shared.WorklogDateFormat, wl.Started)
			if err != nil {
				continue
			}
			date := started.Format(shared.DateFormat)
			if date >= startStr && date <= endStr {
				monthHours += float64(wl.TimeSpentSeconds) / shared.SecondsPerHour
				dateSet[date] = started
			}
		}
		if monthHours > 0 {
			sortedKeys := make([]string, 0, len(dateSet))
			for k := range dateSet {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)
			dates := make([]string, 0, len(sortedKeys))
			for _, k := range sortedKeys {
				dates = append(dates, dateSet[k].Format("2 Jan"))
			}
			rows = append(rows, ticketSummary{Key: t.Key, Summary: t.Summary, Hours: monthHours, Dates: dates})
		}
	}

	renderTicketsPanel(fmt.Sprintf("Tickets Logged - %s", now.Format("January 2006")), statsColor, rows)
}

type ticketSummary struct {
	Key     string
	Summary string
	Hours   float64
	Dates   []string
}

func wrapJoined(items []string, sep, indent string, width int) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(indent)
	lineLen := len(indent)
	for i, item := range items {
		piece := item
		if i < len(items)-1 {
			piece += sep
		}
		if i > 0 && lineLen+len(piece) > width {
			b.WriteString("\n" + indent)
			lineLen = len(indent)
		}
		b.WriteString(piece)
		lineLen += len(piece)
	}
	return b.String()
}

func renderTicketsPanel(title string, accent lipgloss.Color, rows []ticketSummary) {
	if len(rows) == 0 {
		return
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Hours > rows[j].Hours
	})

	keyWidth := 0
	for _, r := range rows {
		if len(r.Key) > keyWidth {
			keyWidth = len(r.Key)
		}
	}
	const panelInnerWidth = 56
	summaryWidth := panelInnerWidth - keyWidth - 4 - 6 - 2

	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3FC"))

	var lines []string
	for i, r := range rows {
		summary := r.Summary
		if len(summary) > summaryWidth {
			summary = summary[:summaryWidth-3] + "..."
		}
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, fmt.Sprintf("%-*s    %5.1fh  %s", keyWidth, r.Key, r.Hours, summary))
		if len(r.Dates) > 0 {
			lines = append(lines, dateStyle.Render(wrapJoined(r.Dates, ", ", "    ", panelInnerWidth)))
		}
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent).
		Margin(0, 0, 1, 0)

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 2).
		Margin(1, 0).
		Width(60)

	fmt.Println(titleStyle.Render(title))
	fmt.Println(panelStyle.Render(strings.Join(lines, "\n")))
}

func showTicketsLogged(ticketLogs map[string]*shared.TicketWorklog) {
	rows := make([]ticketSummary, 0, len(ticketLogs))
	for _, t := range ticketLogs {
		if t != nil && t.Total > 0 {
			rows = append(rows, ticketSummary{Key: t.Key, Summary: t.Summary, Hours: t.Total})
		}
	}
	renderTicketsPanel("Tickets Logged - Year to Date", lipgloss.Color("#7D56F4"), rows)
}

func createProgressBar(percentage float64) string {
	barWidth := 40
	filledWidth := int((percentage / 100) * float64(barWidth))
	if filledWidth > barWidth {
		filledWidth = barWidth
	}

	// Choose color based on percentage
	var fillColor lipgloss.Color
	if percentage >= 90 {
		fillColor = lipgloss.Color("#43BF6D") // Green
	} else if percentage >= 70 {
		fillColor = lipgloss.Color("#FDFF90") // Yellow
	} else if percentage >= 50 {
		fillColor = lipgloss.Color("#F77F00") // Orange
	} else {
		fillColor = lipgloss.Color("#FF5F87") // Red
	}

	// Create filled and empty parts using lipgloss
	filledStyle := lipgloss.NewStyle().Background(fillColor)
	emptyStyle := lipgloss.NewStyle().Background(lipgloss.Color("#333"))

	// Build the bar
	filled := filledStyle.Render(strings.Repeat(" ", filledWidth))
	empty := emptyStyle.Render(strings.Repeat(" ", barWidth-filledWidth))

	// Wrap in border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#666"))

	return borderStyle.Render(filled + empty)
}
