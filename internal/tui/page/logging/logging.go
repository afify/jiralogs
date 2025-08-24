package logging

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"LogS/internal/app"
	"LogS/internal/tui/components/core"
	"LogS/internal/tui/components/core/layout"
	"LogS/internal/tui/components/logo"
	"LogS/internal/tui/page"
	"LogS/internal/tui/page/welcome"
	"LogS/internal/tui/styles"
	"LogS/internal/tui/util"
	"LogS/shared"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

var LoggingPageID page.PageID = "logging"

type LoggingPage interface {
	util.Model
	layout.Help
}

type LoggingMode int

const (
	ModeLogToday           LoggingMode = iota // Log today (8hrs) to a specific ticket
	ModeEvenDistribution                      // Log missing days with even distribution
	ModeSingleTicket                          // Log missing days all to single ticket
	ModeCustomDistribution                    // Log missing days with custom distribution
	ModeCreateTicket                          // Create a new JIRA ticket
)

type FormField int

const (
	FieldMode FormField = iota
	FieldTicketKey
	FieldHours
	FieldDate
	FieldComment
	FieldProject     // For ticket creation
	FieldSummary     // For ticket creation
	FieldDescription // For ticket creation
	FieldIssueType   // For ticket creation
	FieldSubmit
)

type SubmissionStatus int

const (
	StatusIdle SubmissionStatus = iota
	StatusSubmitting
	StatusSuccess
	StatusError
)

type loggingPage struct {
	width, height int
	app           *app.App
	keyMap        KeyMap

	// Form fields
	selectedMode   LoggingMode
	ticketKeyInput textinput.Model
	hoursInput     textinput.Model
	dateInput      textinput.Model
	commentInput   textinput.Model

	// Ticket creation fields
	projectInput     textinput.Model
	summaryInput     textinput.Model
	descriptionInput textinput.Model
	issueTypeInput   textinput.Model

	// Form state
	focusedField   FormField
	submission     SubmissionStatus
	errorMessage   string
	successMessage string

	// Available tickets for suggestion
	availableTickets    []string
	ticketSuggestions   []string
	selectedTicketIndex int // For ticket list selection

	// Available projects for ticket creation
	availableProjects    []string
	projectSuggestions   []string
	selectedProjectIndex int // For project list selection

	// Missing days data
	missingDays []string           // List of missing workdays
	dailyHours  map[string]float64 // Existing daily hours

	// Ticket data with details
	ticketLogs map[string]*shared.TicketWorklog // Full ticket data

	// Custom distribution data
	customDistribution map[string]map[string]float64 // ticket -> date -> hours
	selectedTicketIdx  int                           // For custom distribution ticket selection
	selectedDayIdx     int                           // For custom distribution day selection
}

func New(app *app.App) LoggingPage {
	shared.LogErrorf("LOGGING_NEW", "Creating new logging page")

	// Initialize form inputs
	ticketKeyInput := textinput.New()
	ticketKeyInput.Placeholder = "Enter ticket key (e.g., PT-1234)"
	ticketKeyInput.CharLimit = 20
	ticketKeyInput.SetWidth(30)

	hoursInput := textinput.New()
	hoursInput.Placeholder = "Hours (e.g., 8.0)"
	hoursInput.CharLimit = 10
	hoursInput.SetWidth(20)

	dateInput := textinput.New()
	dateInput.Placeholder = "YYYY-MM-DD"
	dateInput.CharLimit = 10
	dateInput.SetWidth(20)
	dateInput.SetValue(time.Now().Format("2006-01-02")) // Default to today

	commentInput := textinput.New()
	commentInput.Placeholder = "Work description (optional)"
	commentInput.CharLimit = 200
	commentInput.SetWidth(60)

	// Ticket creation fields
	projectInput := textinput.New()
	projectInput.Placeholder = "Project key (e.g., PT)"
	projectInput.CharLimit = 20
	projectInput.SetWidth(30)

	summaryInput := textinput.New()
	summaryInput.Placeholder = "Ticket summary/title"
	summaryInput.CharLimit = 100
	summaryInput.SetWidth(60)

	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "Detailed description (optional)"
	descriptionInput.CharLimit = 500
	descriptionInput.SetWidth(80)

	issueTypeInput := textinput.New()
	issueTypeInput.Placeholder = "Issue type (e.g., Task, Story, Bug)"
	issueTypeInput.CharLimit = 50
	issueTypeInput.SetWidth(40)
	issueTypeInput.SetValue("Task") // Default value

	return &loggingPage{
		app:                app,
		keyMap:             DefaultKeyMap(),
		selectedMode:       ModeLogToday, // Default to log today mode
		ticketKeyInput:     ticketKeyInput,
		hoursInput:         hoursInput,
		dateInput:          dateInput,
		commentInput:       commentInput,
		projectInput:       projectInput,
		summaryInput:       summaryInput,
		descriptionInput:   descriptionInput,
		issueTypeInput:     issueTypeInput,
		focusedField:       FieldMode,
		submission:         StatusIdle,
		availableTickets:   make([]string, 0),
		ticketSuggestions:  make([]string, 0),
		availableProjects:  make([]string, 0),
		projectSuggestions: make([]string, 0),
		missingDays:        make([]string, 0),
		dailyHours:         make(map[string]float64),
		customDistribution: make(map[string]map[string]float64),
	}
}

func (p *loggingPage) Init() tea.Cmd {
	shared.LogErrorf("LOGGING_INIT", "Initializing logging page")
	// Focus the first input field
	p.ticketKeyInput.Focus()
	return textinput.Blink
}

func (p *loggingPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	shared.LogErrorf("LOGGING_UPDATE", "Received message type: %T", msg)

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case welcome.WorklogDataMsg:
		shared.LogErrorf("LOGGING_UPDATE", "Received WorklogDataMsg")
		p.setAvailableTickets(msg.TicketLogs)
		p.setWorklogData(msg.DailyHours)
		return p, nil

	case WorklogSubmittedMsg:
		shared.LogErrorf("LOGGING_UPDATE", "Received WorklogSubmittedMsg")
		if msg.Success {
			p.submission = StatusSuccess
			p.successMessage = msg.Message
			p.errorMessage = ""
		} else {
			p.submission = StatusError
			p.errorMessage = msg.Message
			p.successMessage = ""
		}
		return p, nil

	case tea.WindowSizeMsg:
		shared.LogErrorf("LOGGING_UPDATE", "Window resize: %dx%d", msg.Width, msg.Height)
		return p, p.SetSize(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		shared.LogErrorf("LOGGING_UPDATE", "Key press: %s", msg.String())

		switch {
		case key.Matches(msg, p.keyMap.Cancel):
			shared.LogErrorf("LOGGING_UPDATE", "Cancel key pressed")
			return p, func() tea.Msg {
				return page.PageChangeMsg{ID: "stats"}
			}

		case key.Matches(msg, p.keyMap.NextField):
			p.nextField()
			return p, nil

		case key.Matches(msg, p.keyMap.PrevField):
			p.prevField()
			return p, nil

		case key.Matches(msg, p.keyMap.Up):
			if p.focusedField == FieldMode {
				p.cycleModeUp()
				return p, nil
			} else if p.focusedField == FieldTicketKey && p.selectedMode == ModeLogToday {
				p.prevTicket()
				return p, nil
			} else if p.focusedField == FieldTicketKey && p.selectedMode == ModeCustomDistribution {
				p.prevCustomTicket()
				return p, nil
			} else if p.focusedField == FieldProject && p.selectedMode == ModeCreateTicket {
				p.prevProject()
				return p, nil
			}

		case key.Matches(msg, p.keyMap.Down):
			if p.focusedField == FieldMode {
				p.cycleModeDown()
				return p, nil
			} else if p.focusedField == FieldTicketKey && p.selectedMode == ModeLogToday {
				p.nextTicket()
				return p, nil
			} else if p.focusedField == FieldTicketKey && p.selectedMode == ModeCustomDistribution {
				p.nextCustomTicket()
				return p, nil
			} else if p.focusedField == FieldProject && p.selectedMode == ModeCreateTicket {
				p.nextProject()
				return p, nil
			}

		case key.Matches(msg, p.keyMap.Submit):
			if p.focusedField == FieldSubmit {
				return p, p.submitWorklog()
			} else {
				p.nextField()
				return p, nil
			}

		case key.Matches(msg, p.keyMap.Clear):
			p.clearForm()
			return p, nil

		case key.Matches(msg, p.keyMap.Quit):
			shared.LogErrorf("LOGGING_UPDATE", "Quit key pressed")
			return p, tea.Quit

		default:
			// Handle custom distribution input
			if p.focusedField == FieldTicketKey && p.selectedMode == ModeCustomDistribution {
				return p, p.handleCustomDistributionInput(msg)
			}
		}
	}

	// Update the focused input field
	var cmd tea.Cmd
	switch p.focusedField {
	case FieldTicketKey:
		p.ticketKeyInput, cmd = p.ticketKeyInput.Update(msg)
		cmds = append(cmds, cmd)
		// Update ticket suggestions
		p.updateTicketSuggestions()
	case FieldHours:
		p.hoursInput, cmd = p.hoursInput.Update(msg)
		cmds = append(cmds, cmd)
	case FieldDate:
		p.dateInput, cmd = p.dateInput.Update(msg)
		cmds = append(cmds, cmd)
	case FieldComment:
		p.commentInput, cmd = p.commentInput.Update(msg)
		cmds = append(cmds, cmd)
	case FieldProject:
		p.projectInput, cmd = p.projectInput.Update(msg)
		cmds = append(cmds, cmd)
		// Update project suggestions
		p.updateProjectSuggestions()
	case FieldSummary:
		p.summaryInput, cmd = p.summaryInput.Update(msg)
		cmds = append(cmds, cmd)
	case FieldDescription:
		p.descriptionInput, cmd = p.descriptionInput.Update(msg)
		cmds = append(cmds, cmd)
	case FieldIssueType:
		p.issueTypeInput, cmd = p.issueTypeInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return p, tea.Batch(cmds...)
}

func (p *loggingPage) View() string {
	shared.LogErrorf("LOGGING_VIEW", "Rendering logging view - dimensions: %dx%d", p.width, p.height)

	if p.width == 0 || p.height == 0 {
		shared.LogErrorf("LOGGING_VIEW", "Dimensions not set, returning empty view")
		return ""
	}

	t := styles.CurrentTheme()
	shared.LogErrorf("LOGGING_VIEW", "Retrieved current theme")

	// Create the full-width top logo with Crush's exact colors
	logoOpts := logo.Opts{
		FieldColor:   t.Primary,
		TitleColorA:  t.Secondary,
		TitleColorB:  t.Primary,
		CharmColor:   t.Secondary,
		VersionColor: t.Primary,
		Width:        p.width - 6,
	}

	logoStr := logo.Render(false, logoOpts)
	shared.LogErrorf("LOGGING_VIEW", "Full-width logo created")

	// Create form content
	formContent := p.renderForm(t)
	shared.LogErrorf("LOGGING_VIEW", "Form content rendered")

	// Join logo and form vertically
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		logoStr,
		"", // Empty line for spacing
		formContent,
	)

	// Apply base styling
	baseStyle := t.S().Base.
		Width(p.width).
		Height(p.height).
		Padding(1, 2)

	return baseStyle.Render(content)
}

func (p *loggingPage) renderForm(t *styles.Theme) string {
	// Form container
	formStyle := t.S().Base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(2, 4).
		Width(p.width - 8)

	// Form title
	title := t.S().Base.
		Foreground(t.Primary).
		Bold(true).
		Align(lipgloss.Center).
		Render("Log Time to JIRA Ticket")

	// Form fields
	var formFields []string

	// Mode selection field
	modeLabel := p.renderFieldLabel(t, "Logging Mode:", p.focusedField == FieldMode)
	modeField := p.renderModeSelector(t)
	formFields = append(formFields, modeLabel, modeField, "")

	// Show different fields based on selected mode
	switch p.selectedMode {
	case ModeLogToday:
		// Log Today mode: show ticket list selection and today's info
		ticketLabel := p.renderFieldLabel(t, "Select Ticket for Today:", p.focusedField == FieldTicketKey)
		ticketList := p.renderTicketList(t)
		formFields = append(formFields, ticketLabel, ticketList, "")

		// Show today's date and 8 hours info
		todayInfo := p.renderTodayInfo(t)
		formFields = append(formFields, todayInfo, "")

	case ModeEvenDistribution:
		// Even distribution: show missing days info and distribution plan
		infoText := p.renderMissingDaysInfo(t)
		formFields = append(formFields, infoText, "")

		// Show distribution plan preview
		distributionPlan := p.renderEvenDistributionPlan(t)
		formFields = append(formFields, distributionPlan, "")

	case ModeSingleTicket:
		// Single ticket for all missing days: show ticket and comment
		ticketLabel := p.renderFieldLabel(t, "Ticket Key:", p.focusedField == FieldTicketKey)
		ticketField := p.ticketKeyInput.View()
		if len(p.ticketSuggestions) > 0 && p.focusedField == FieldTicketKey {
			suggestions := p.renderSuggestions(t)
			ticketField = lipgloss.JoinVertical(lipgloss.Left, ticketField, suggestions)
		}
		formFields = append(formFields, ticketLabel, ticketField, "")

		infoText := p.renderMissingDaysInfo(t)
		formFields = append(formFields, infoText, "")

	case ModeCustomDistribution:
		// Custom distribution: show ticket-day matrix interface
		infoText := p.renderMissingDaysInfo(t)
		formFields = append(formFields, infoText, "")

		customLabel := p.renderFieldLabel(t, "Custom Distribution (Hours per Ticket per Day):", p.focusedField == FieldTicketKey)
		customMatrix := p.renderCustomDistributionMatrix(t)
		formFields = append(formFields, customLabel, customMatrix, "")

	case ModeCreateTicket:
		// Create ticket mode: show project selection list, summary, description, issue type fields
		projectLabel := p.renderFieldLabel(t, "Select Project:", p.focusedField == FieldProject)
		projectList := p.renderProjectList(t)
		formFields = append(formFields, projectLabel, projectList, "")

		summaryLabel := p.renderFieldLabel(t, "Summary/Title:", p.focusedField == FieldSummary)
		summaryField := p.summaryInput.View()
		formFields = append(formFields, summaryLabel, summaryField, "")

		issueTypeLabel := p.renderFieldLabel(t, "Issue Type:", p.focusedField == FieldIssueType)
		issueTypeField := p.issueTypeInput.View()
		formFields = append(formFields, issueTypeLabel, issueTypeField, "")

		descriptionLabel := p.renderFieldLabel(t, "Description:", p.focusedField == FieldDescription)
		descriptionField := p.descriptionInput.View()
		formFields = append(formFields, descriptionLabel, descriptionField, "")
	}

	// Comment field (common for most modes, but not custom distribution or create ticket)
	if p.selectedMode != ModeCustomDistribution && p.selectedMode != ModeCreateTicket {
		commentLabel := p.renderFieldLabel(t, "Comment:", p.focusedField == FieldComment)
		commentField := p.commentInput.View()
		formFields = append(formFields, commentLabel, commentField, "")
	}

	// Submit button
	submitButton := p.renderSubmitButton(t)
	formFields = append(formFields, submitButton)

	// Status messages
	if p.submission == StatusSubmitting {
		spinner := t.S().Base.Foreground(t.Citron).Render("Submitting worklog...")
		formFields = append(formFields, "", spinner)
	} else if p.submission == StatusSuccess {
		successMsg := t.S().Base.Foreground(t.Green).Render("✓ " + p.successMessage)
		formFields = append(formFields, "", successMsg)
	} else if p.submission == StatusError {
		errorMsg := t.S().Base.Foreground(t.Blue).Render("✗ " + p.errorMessage)
		formFields = append(formFields, "", errorMsg)
	}

	formContent := lipgloss.JoinVertical(lipgloss.Left, formFields...)

	return formStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		formContent,
	))
}

func (p *loggingPage) renderFieldLabel(t *styles.Theme, label string, focused bool) string {
	style := t.S().Base.Foreground(t.FgMuted)
	if focused {
		style = style.Foreground(t.Primary).Bold(true)
	}
	return style.Render(label)
}

func (p *loggingPage) renderSuggestions(t *styles.Theme) string {
	if len(p.ticketSuggestions) == 0 {
		return ""
	}

	maxSuggestions := 5
	if len(p.ticketSuggestions) > maxSuggestions {
		p.ticketSuggestions = p.ticketSuggestions[:maxSuggestions]
	}

	suggestionStyle := t.S().Base.
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		Background(t.BgSubtle).
		Padding(0, 1).
		Width(p.ticketKeyInput.Width())

	var suggestions []string
	for i, ticket := range p.ticketSuggestions {
		style := t.S().Base.Foreground(t.FgBase)
		if i == 0 { // Highlight first suggestion
			style = style.Foreground(t.Primary).Bold(true)
		}
		suggestions = append(suggestions, style.Render("  "+ticket))
	}

	return suggestionStyle.Render(lipgloss.JoinVertical(lipgloss.Left, suggestions...))
}

func (p *loggingPage) renderSubmitButton(t *styles.Theme) string {
	buttonText := "Submit Worklog"
	if p.focusedField == FieldSubmit {
		return t.S().Base.
			Background(t.Primary).
			Foreground(t.White).
			Bold(true).
			Padding(0, 2).
			Render(buttonText)
	}
	return t.S().Base.
		Background(t.BgSubtle).
		Foreground(t.FgBase).
		Padding(0, 2).
		Render(buttonText)
}

func (p *loggingPage) nextField() {
	p.blurCurrentField()

	// Navigate based on current mode
	switch p.focusedField {
	case FieldMode:
		switch p.selectedMode {
		case ModeEvenDistribution:
			p.focusedField = FieldComment
		case ModeCreateTicket:
			p.focusedField = FieldProject
		default:
			p.focusedField = FieldTicketKey
		}
	case FieldTicketKey:
		p.focusedField = FieldComment
	case FieldHours:
		p.focusedField = FieldDate
	case FieldDate:
		p.focusedField = FieldComment
	case FieldComment:
		p.focusedField = FieldSubmit
	case FieldProject:
		p.focusedField = FieldSummary
	case FieldSummary:
		p.focusedField = FieldIssueType
	case FieldIssueType:
		p.focusedField = FieldDescription
	case FieldDescription:
		p.focusedField = FieldSubmit
	case FieldSubmit:
		p.focusedField = FieldMode
	}

	p.focusNewField()
}

func (p *loggingPage) prevField() {
	p.blurCurrentField()

	// Navigate based on current mode
	switch p.focusedField {
	case FieldMode:
		p.focusedField = FieldSubmit
	case FieldTicketKey:
		p.focusedField = FieldMode
	case FieldHours:
		p.focusedField = FieldTicketKey
	case FieldDate:
		p.focusedField = FieldHours
	case FieldComment:
		if p.selectedMode == ModeEvenDistribution {
			p.focusedField = FieldMode
		} else {
			p.focusedField = FieldTicketKey
		}
	case FieldProject:
		p.focusedField = FieldMode
	case FieldSummary:
		p.focusedField = FieldProject
	case FieldIssueType:
		p.focusedField = FieldSummary
	case FieldDescription:
		p.focusedField = FieldIssueType
	case FieldSubmit:
		switch p.selectedMode {
		case ModeEvenDistribution:
			p.focusedField = FieldComment
		case ModeCreateTicket:
			p.focusedField = FieldDescription
		default:
			p.focusedField = FieldComment
		}
	}

	p.focusNewField()
}

func (p *loggingPage) blurCurrentField() {
	switch p.focusedField {
	case FieldTicketKey:
		p.ticketKeyInput.Blur()
	case FieldHours:
		p.hoursInput.Blur()
	case FieldDate:
		p.dateInput.Blur()
	case FieldComment:
		p.commentInput.Blur()
	case FieldProject:
		p.projectInput.Blur()
	case FieldSummary:
		p.summaryInput.Blur()
	case FieldDescription:
		p.descriptionInput.Blur()
	case FieldIssueType:
		p.issueTypeInput.Blur()
	}
}

func (p *loggingPage) focusNewField() {
	switch p.focusedField {
	case FieldTicketKey:
		p.ticketKeyInput.Focus()
	case FieldHours:
		p.hoursInput.Focus()
	case FieldDate:
		p.dateInput.Focus()
	case FieldComment:
		p.commentInput.Focus()
	case FieldProject:
		p.projectInput.Focus()
	case FieldSummary:
		p.summaryInput.Focus()
	case FieldDescription:
		p.descriptionInput.Focus()
	case FieldIssueType:
		p.issueTypeInput.Focus()
	}
}

func (p *loggingPage) clearForm() {
	p.ticketKeyInput.SetValue("")
	p.hoursInput.SetValue("")
	p.dateInput.SetValue(time.Now().Format("2006-01-02"))
	p.commentInput.SetValue("")
	p.submission = StatusIdle
	p.errorMessage = ""
	p.successMessage = ""
	p.focusedField = FieldMode
	p.blurCurrentField() // Blur all fields since Mode doesn't have a text input
}

func (p *loggingPage) validateForm() error {
	switch p.selectedMode {
	case ModeLogToday:
		// Log Today mode - check if ticket is selected
		if len(p.availableTickets) == 0 {
			return fmt.Errorf("no tickets available to log time to")
		}
		if p.selectedTicketIndex >= len(p.availableTickets) {
			return fmt.Errorf("invalid ticket selection")
		}

	case ModeEvenDistribution:
		// Even distribution - no specific validation needed
		if len(p.missingDays) == 0 {
			return fmt.Errorf("no missing days to log")
		}

	case ModeSingleTicket:
		// Single ticket for all missing days
		if strings.TrimSpace(p.ticketKeyInput.Value()) == "" {
			return fmt.Errorf("ticket key is required")
		}

		if len(p.missingDays) == 0 {
			return fmt.Errorf("no missing days to log")
		}

	case ModeCustomDistribution:
		// Custom distribution - validate that total hours make sense
		if len(p.missingDays) == 0 {
			return fmt.Errorf("no missing days to log")
		}

		totalHours := 0.0
		for _, dayHours := range p.customDistribution {
			for _, hours := range dayHours {
				totalHours += hours
			}
		}

		if totalHours == 0 {
			return fmt.Errorf("no hours specified in custom distribution")
		}

	case ModeCreateTicket:
		// Create ticket mode - validate required fields
		if len(p.availableProjects) == 0 {
			return fmt.Errorf("no projects available")
		}
		if p.selectedProjectIndex >= len(p.availableProjects) {
			return fmt.Errorf("please select a valid project")
		}
		if strings.TrimSpace(p.summaryInput.Value()) == "" {
			return fmt.Errorf("ticket summary is required")
		}
		if strings.TrimSpace(p.issueTypeInput.Value()) == "" {
			return fmt.Errorf("issue type is required")
		}
	}

	return nil
}

func (p *loggingPage) submitWorklog() tea.Cmd {
	if err := p.validateForm(); err != nil {
		p.submission = StatusError
		p.errorMessage = err.Error()
		return nil
	}

	p.submission = StatusSubmitting
	p.errorMessage = ""
	p.successMessage = ""

	// Actual JIRA API submission based on mode
	return func() tea.Msg {
		jiraClient := p.app.JiraClient()
		if jiraClient == nil {
			return WorklogSubmittedMsg{
				Success: false,
				Message: "JIRA client not available",
			}
		}

		comment := strings.TrimSpace(p.commentInput.Value())

		var totalEntries int
		var totalHours float64
		var errors []string

		switch p.selectedMode {
		case ModeLogToday:
			// Log 8 hours today to selected ticket
			selectedTicket := p.availableTickets[p.selectedTicketIndex]
			today := time.Now().Format("2006-01-02")

			err := jiraClient.AddWorklog(selectedTicket, today, 8.0, comment)
			if err != nil {
				return WorklogSubmittedMsg{
					Success: false,
					Message: fmt.Sprintf("Failed to log to %s: %v", selectedTicket, err),
				}
			}

			totalHours = 8.0
			totalEntries = 1
			return WorklogSubmittedMsg{
				Success: true,
				Message: fmt.Sprintf("Logged 8.0 hours to %s on %s (today)", selectedTicket, today),
			}

		case ModeEvenDistribution:
			// Distribute missing days evenly across all available tickets using calculated plan
			if len(p.availableTickets) == 0 || len(p.missingDays) == 0 {
				return WorklogSubmittedMsg{
					Success: false,
					Message: "No tickets or missing days available",
				}
			}

			// Use the same distribution calculation as shown in preview
			distributionMap := p.calculateEvenDistribution()

			for ticket, days := range distributionMap {
				for _, day := range days {
					err := jiraClient.AddWorklog(ticket, day, 8.0, comment)
					if err != nil {
						errors = append(errors, fmt.Sprintf("%s on %s: %v", ticket, day, err))
					} else {
						totalHours += 8.0
						totalEntries++
					}
				}
			}

		case ModeSingleTicket:
			// Log all missing days to single ticket
			ticketKey := strings.TrimSpace(p.ticketKeyInput.Value())

			for _, day := range p.missingDays {
				err := jiraClient.AddWorklog(ticketKey, day, 8.0, comment)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s on %s: %v", ticketKey, day, err))
				} else {
					totalHours += 8.0
					totalEntries++
				}
			}

		case ModeCustomDistribution:
			// Log custom hours per ticket per day
			for ticketKey, dayHours := range p.customDistribution {
				for day, hours := range dayHours {
					if hours > 0 {
						err := jiraClient.AddWorklog(ticketKey, day, hours, comment)
						if err != nil {
							errors = append(errors, fmt.Sprintf("%s on %s (%.1fh): %v", ticketKey, day, hours, err))
						} else {
							totalHours += hours
							totalEntries++
						}
					}
				}
			}

		case ModeCreateTicket:
			// Create a new JIRA ticket using selected project from list
			var projectKey string
			if len(p.availableProjects) > 0 && p.selectedProjectIndex < len(p.availableProjects) {
				projectKey = p.availableProjects[p.selectedProjectIndex]
			}
			summary := strings.TrimSpace(p.summaryInput.Value())
			description := strings.TrimSpace(p.descriptionInput.Value())
			issueType := strings.TrimSpace(p.issueTypeInput.Value())

			ticketKey, err := jiraClient.CreateIssue(projectKey, summary, description, issueType)
			if err != nil {
				return WorklogSubmittedMsg{
					Success: false,
					Message: fmt.Sprintf("Failed to create ticket: %v", err),
				}
			}

			return WorklogSubmittedMsg{
				Success: true,
				Message: fmt.Sprintf("Created ticket %s: %s", ticketKey, summary),
			}
		}

		// Return result based on success/failure
		if len(errors) > 0 {
			errorMsg := fmt.Sprintf("Completed %d entries (%.1fh) with errors: %s",
				totalEntries, totalHours, errors[0])
			if len(errors) > 1 {
				errorMsg += fmt.Sprintf(" (and %d more)", len(errors)-1)
			}
			return WorklogSubmittedMsg{
				Success: totalEntries > 0, // Partial success if some entries worked
				Message: errorMsg,
			}
		}

		// Full success
		var message string
		switch p.selectedMode {
		case ModeEvenDistribution:
			message = fmt.Sprintf("Logged %.1f hours evenly across %d days to %d tickets",
				totalHours, len(p.missingDays), len(p.availableTickets))
		case ModeSingleTicket:
			message = fmt.Sprintf("Logged %.1f hours across %d days to %s",
				totalHours, len(p.missingDays), p.ticketKeyInput.Value())
		case ModeCustomDistribution:
			message = fmt.Sprintf("Logged %.1f hours across %d custom entries",
				totalHours, totalEntries)
		default:
			message = fmt.Sprintf("Logged %.1f hours in %d entries", totalHours, totalEntries)
		}

		return WorklogSubmittedMsg{
			Success: true,
			Message: message,
		}
	}
}

func (p *loggingPage) parseHours() float64 {
	hours, _ := strconv.ParseFloat(strings.TrimSpace(p.hoursInput.Value()), 64)
	return hours
}

func (p *loggingPage) updateTicketSuggestions() {
	input := strings.ToUpper(strings.TrimSpace(p.ticketKeyInput.Value()))
	if input == "" {
		p.ticketSuggestions = []string{}
		return
	}

	var suggestions []string
	for _, ticket := range p.availableTickets {
		if strings.HasPrefix(strings.ToUpper(ticket), input) {
			suggestions = append(suggestions, ticket)
		}
	}

	p.ticketSuggestions = suggestions
}

func (p *loggingPage) updateProjectSuggestions() {
	input := strings.ToUpper(strings.TrimSpace(p.projectInput.Value()))
	if input == "" {
		p.projectSuggestions = []string{}
		return
	}

	var suggestions []string
	for _, project := range p.availableProjects {
		if strings.HasPrefix(strings.ToUpper(project), input) {
			suggestions = append(suggestions, project)
		}
	}

	p.projectSuggestions = suggestions
}

func (p *loggingPage) renderProjectSuggestions(t *styles.Theme) string {
	if len(p.projectSuggestions) == 0 {
		return ""
	}

	maxSuggestions := 5
	if len(p.projectSuggestions) > maxSuggestions {
		p.projectSuggestions = p.projectSuggestions[:maxSuggestions]
	}

	suggestionStyle := t.S().Base.
		Border(lipgloss.NormalBorder()).
		BorderForeground(t.Border).
		Background(t.BgSubtle).
		Padding(0, 1).
		Width(p.projectInput.Width())

	var suggestions []string
	for i, project := range p.projectSuggestions {
		style := t.S().Base.Foreground(t.FgBase)
		if i == 0 { // Highlight first suggestion
			style = style.Foreground(t.Primary).Bold(true)
		}
		suggestions = append(suggestions, style.Render("  "+project))
	}

	return suggestionStyle.Render(lipgloss.JoinVertical(lipgloss.Left, suggestions...))
}

func (p *loggingPage) setAvailableTickets(ticketLogs map[string]*shared.TicketWorklog) {
	p.ticketLogs = ticketLogs
	p.availableTickets = make([]string, 0, len(ticketLogs))
	for ticketKey := range ticketLogs {
		p.availableTickets = append(p.availableTickets, ticketKey)
	}

	// Also load available projects for ticket creation
	p.loadAvailableProjects()
}

func (p *loggingPage) loadAvailableProjects() {
	jiraClient := p.app.JiraClient()
	if jiraClient == nil {
		return
	}

	// Fetch projects in background
	go func() {
		projects, err := jiraClient.GetProjects()
		if err != nil {
			shared.LogErrorf("LOGGING_PROJECTS", "Failed to load projects: %v", err)
			return
		}

		projectKeys := make([]string, 0, len(projects))
		for _, project := range projects {
			projectKeys = append(projectKeys, project.Key)
		}

		p.availableProjects = projectKeys
		shared.LogErrorf("LOGGING_PROJECTS", "Loaded %d projects", len(projectKeys))
	}()
}

func (p *loggingPage) setWorklogData(dailyHours map[string]float64) {
	p.dailyHours = dailyHours
	p.calculateMissingDays()
}

func (p *loggingPage) calculateMissingDays() {
	// Calculate missing workdays in the current period (last 30 days for example)
	now := time.Now()
	start := now.AddDate(0, -1, 0) // Last month

	p.missingDays = make([]string, 0)

	current := start
	for !current.After(now) {
		// Skip weekends (Friday/Saturday for Middle East)
		if current.Weekday() != time.Friday && current.Weekday() != time.Saturday {
			dateStr := current.Format("2006-01-02")
			// Check if this day has logged hours
			if hours, exists := p.dailyHours[dateStr]; !exists || hours < 8.0 {
				p.missingDays = append(p.missingDays, dateStr)
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	shared.LogErrorf("LOGGING_DATA", "Found %d missing days", len(p.missingDays))
}

func (p *loggingPage) cycleModeUp() {
	switch p.selectedMode {
	case ModeLogToday:
		p.selectedMode = ModeCreateTicket
	case ModeEvenDistribution:
		p.selectedMode = ModeLogToday
	case ModeSingleTicket:
		p.selectedMode = ModeEvenDistribution
	case ModeCustomDistribution:
		p.selectedMode = ModeSingleTicket
	case ModeCreateTicket:
		p.selectedMode = ModeCustomDistribution
	}
}

func (p *loggingPage) cycleModeDown() {
	switch p.selectedMode {
	case ModeLogToday:
		p.selectedMode = ModeEvenDistribution
	case ModeEvenDistribution:
		p.selectedMode = ModeSingleTicket
	case ModeSingleTicket:
		p.selectedMode = ModeCustomDistribution
	case ModeCustomDistribution:
		p.selectedMode = ModeCreateTicket
	case ModeCreateTicket:
		p.selectedMode = ModeLogToday
	}
}

func (p *loggingPage) renderModeSelector(t *styles.Theme) string {
	modes := []struct {
		text  string
		mode  LoggingMode
		color color.Color
	}{
		{"1. Log Today → Log 8 hours today to a ticket", ModeLogToday, t.Blue},
		{"2. Even Distribution → Missing days evenly across tickets", ModeEvenDistribution, t.Green},
		{"3. Single Ticket → All missing days to one ticket", ModeSingleTicket, t.Citron},
		{"4. Custom Distribution → Manual hours per ticket per day", ModeCustomDistribution, t.FgMuted},
		{"5. Create Ticket → Create a new JIRA ticket", ModeCreateTicket, t.Error},
	}

	var modeLines []string

	// Add header
	headerStyle := t.S().Base.
		Foreground(t.Blue).
		Bold(true)
	modeLines = append(modeLines, headerStyle.Render("Select Logging Mode (↑/↓ to change):"))
	modeLines = append(modeLines, "")

	// Show all modes with consistent formatting
	for _, mode := range modes {
		if mode.mode == p.selectedMode {
			// Selected mode - highlighted with arrow and color
			indicator := t.S().Base.
				Foreground(mode.color).
				Bold(true).
				Render("▶")

			textStyle := t.S().Base.
				Foreground(mode.color).
				Bold(true)

			modeLines = append(modeLines, "  "+indicator+" "+textStyle.Render(mode.text))
		} else {
			// Non-selected modes - muted with colored dot
			indicator := t.S().Base.
				Foreground(mode.color).
				Render("●")

			textStyle := t.S().Base.
				Foreground(t.FgMuted)

			modeLines = append(modeLines, "  "+indicator+" "+textStyle.Render(mode.text))
		}
	}

	// Container without border - clean look
	containerStyle := t.S().Base.
		Padding(0, 2).
		Width(70)

	return containerStyle.Render(lipgloss.JoinVertical(lipgloss.Left, modeLines...))
}

func (p *loggingPage) getModeColor(t *styles.Theme) color.Color {
	switch p.selectedMode {
	case ModeLogToday:
		return t.Blue
	case ModeEvenDistribution:
		return t.Green
	case ModeSingleTicket:
		return t.Citron
	case ModeCustomDistribution:
		return t.FgMuted
	}
	return t.Blue
}

func (p *loggingPage) renderMissingDaysInfo(t *styles.Theme) string {
	if len(p.missingDays) == 0 {
		return t.S().Base.Foreground(t.Green).Render("✓ No missing days found!")
	}

	infoLines := []string{
		t.S().Base.Foreground(t.Citron).Bold(true).Render(fmt.Sprintf("Missing Days: %d", len(p.missingDays))),
		"",
	}

	// Show first few missing days
	maxShow := 5
	for i, day := range p.missingDays {
		if i >= maxShow {
			remaining := len(p.missingDays) - maxShow
			infoLines = append(infoLines, t.S().Base.Foreground(t.FgMuted).Render(fmt.Sprintf("... and %d more", remaining)))
			break
		}
		infoLines = append(infoLines, t.S().Base.Foreground(t.FgBase).Render("  "+day))
	}

	totalHours := float64(len(p.missingDays)) * 8.0
	infoLines = append(infoLines, "")
	infoLines = append(infoLines, t.S().Base.Foreground(t.Blue).Render(fmt.Sprintf("Total Hours to Log: %.1f", totalHours)))

	return lipgloss.JoinVertical(lipgloss.Left, infoLines...)
}

func (p *loggingPage) renderEvenDistributionPlan(t *styles.Theme) string {
	if len(p.availableTickets) == 0 || len(p.missingDays) == 0 {
		return t.S().Base.Foreground(t.FgMuted).Render("No tickets or missing days available for distribution")
	}

	planLines := []string{
		t.S().Base.Foreground(t.Blue).Bold(true).Render("Distribution Plan Preview:"),
		"",
	}

	// Calculate the distribution plan
	distributionMap := p.calculateEvenDistribution()

	// Show the plan grouped by ticket
	for _, ticket := range p.availableTickets {
		if days, exists := distributionMap[ticket]; exists && len(days) > 0 {
			// Get ticket name for display
			ticketDisplay := ticket
			if ticketData, exists := p.ticketLogs[ticket]; exists && ticketData.Summary != "" {
				ticketName := ticketData.Summary
				if len(ticketName) > 30 {
					ticketName = ticketName[:27] + "..."
				}
				ticketDisplay = fmt.Sprintf("%s - %s", ticket, ticketName)
			}

			// Ticket header with total hours
			totalHours := float64(len(days)) * 8.0
			ticketHeader := t.S().Base.Foreground(t.Green).Bold(true).
				Render(fmt.Sprintf("📋 %s: %.1f hours (%d days)", ticketDisplay, totalHours, len(days)))
			planLines = append(planLines, ticketHeader)

			// Show first few days for this ticket
			maxDaysToShow := 3
			for j, day := range days {
				if j >= maxDaysToShow {
					remaining := len(days) - maxDaysToShow
					planLines = append(planLines, t.S().Base.Foreground(t.FgMuted).
						Render(fmt.Sprintf("    ... and %d more days", remaining)))
					break
				}
				dayEntry := t.S().Base.Foreground(t.FgBase).
					Render(fmt.Sprintf("    • %s → 8.0 hours", day))
				planLines = append(planLines, dayEntry)
			}
			planLines = append(planLines, "")
		}
	}

	// Summary - count only tickets that actually get hours in the distribution
	totalDays := len(p.missingDays)
	totalHours := float64(totalDays) * 8.0
	ticketsWithHours := 0
	for _, days := range distributionMap {
		if len(days) > 0 {
			ticketsWithHours++
		}
	}
	summary := t.S().Base.Foreground(t.Citron).Bold(true).
		Render(fmt.Sprintf("Total: %.1f hours across %d days to %d tickets",
			totalHours, totalDays, ticketsWithHours))
	planLines = append(planLines, summary)

	return lipgloss.JoinVertical(lipgloss.Left, planLines...)
}

func (p *loggingPage) calculateEvenDistribution() map[string][]string {
	if len(p.availableTickets) == 0 || len(p.missingDays) == 0 {
		return make(map[string][]string)
	}

	distributionMap := make(map[string][]string)

	// Initialize all tickets
	for _, ticket := range p.availableTickets {
		distributionMap[ticket] = make([]string, 0)
	}

	// Distribute days round-robin across tickets
	for dayIndex, day := range p.missingDays {
		ticketIndex := dayIndex % len(p.availableTickets)
		ticket := p.availableTickets[ticketIndex]
		distributionMap[ticket] = append(distributionMap[ticket], day)
	}

	return distributionMap
}

func (p *loggingPage) nextTicket() {
	if len(p.availableTickets) == 0 {
		return
	}
	p.selectedTicketIndex++
	if p.selectedTicketIndex >= len(p.availableTickets) {
		p.selectedTicketIndex = 0
	}
}

func (p *loggingPage) prevTicket() {
	if len(p.availableTickets) == 0 {
		return
	}
	p.selectedTicketIndex--
	if p.selectedTicketIndex < 0 {
		p.selectedTicketIndex = len(p.availableTickets) - 1
	}
}

func (p *loggingPage) renderTicketList(t *styles.Theme) string {
	if len(p.availableTickets) == 0 {
		return t.S().Base.Foreground(t.FgMuted).Render("No tickets available")
	}

	var ticketLines []string

	maxShow := 6 // Show up to 6 tickets at a time for better readability
	start := 0
	if p.selectedTicketIndex >= maxShow {
		start = p.selectedTicketIndex - maxShow + 1
	}

	for i := start; i < len(p.availableTickets) && i < start+maxShow; i++ {
		ticketKey := p.availableTickets[i]

		// Get ticket details
		ticketName := ticketKey
		hoursLogged := "0.0h"
		if ticketData, exists := p.ticketLogs[ticketKey]; exists {
			if ticketData.Summary != "" {
				ticketName = ticketData.Summary
			}
			hoursLogged = fmt.Sprintf("%.1fh", ticketData.Total)
		}

		// Truncate long ticket names for display but show full information
		truncatedName := ticketName
		if len(truncatedName) > 40 {
			truncatedName = truncatedName[:37] + "..."
		}

		// Format: PT-1234 - Full Ticket Name [8.5h logged]
		displayText := fmt.Sprintf("%s - %s [%s logged]", ticketKey, truncatedName, hoursLogged)

		if i == p.selectedTicketIndex {
			// Selected ticket - highlighted
			selectedStyle := t.S().Base.
				Background(t.Blue).
				Foreground(t.White).
				Bold(true).
				Padding(0, 1).
				Width(70)
			ticketLines = append(ticketLines, "  "+selectedStyle.Render("▶ "+displayText))
		} else {
			// Regular ticket
			normalStyle := t.S().Base.
				Foreground(t.FgBase).
				Padding(0, 1)
			ticketLines = append(ticketLines, "    "+normalStyle.Render(displayText))
		}
	}

	// Add navigation hint if there are more tickets
	if len(p.availableTickets) > maxShow {
		hint := t.S().Base.Foreground(t.FgMuted).Render(
			fmt.Sprintf("  (%d/%d tickets - use ↑/↓ to navigate)",
				p.selectedTicketIndex+1, len(p.availableTickets)))
		ticketLines = append(ticketLines, "", hint)
	}

	return lipgloss.JoinVertical(lipgloss.Left, ticketLines...)
}

func (p *loggingPage) renderTodayInfo(t *styles.Theme) string {
	today := time.Now().Format("Monday, January 2, 2006")

	infoLines := []string{
		t.S().Base.Foreground(t.Blue).Bold(true).Render("Today's Logging:"),
		"",
		t.S().Base.Foreground(t.FgBase).Render("📅 Date: " + today),
		t.S().Base.Foreground(t.FgBase).Render("⏰ Hours: 8.0 (full workday)"),
		"",
		t.S().Base.Foreground(t.FgMuted).Render("This will log a full 8-hour workday to the selected ticket."),
	}

	return lipgloss.JoinVertical(lipgloss.Left, infoLines...)
}

func (p *loggingPage) renderProjectList(t *styles.Theme) string {
	if len(p.availableProjects) == 0 {
		return t.S().Base.Foreground(t.FgMuted).Render("Loading projects...")
	}

	var projectLines []string

	maxShow := 6 // Show up to 6 projects at a time for better readability
	start := 0
	if p.selectedProjectIndex >= maxShow {
		start = p.selectedProjectIndex - maxShow + 1
	}

	for i := start; i < len(p.availableProjects) && i < start+maxShow; i++ {
		projectKey := p.availableProjects[i]

		// Format: PROJECT_KEY
		displayText := projectKey

		if i == p.selectedProjectIndex {
			// Selected project - highlighted
			selectedStyle := t.S().Base.
				Background(t.Blue).
				Foreground(t.White).
				Bold(true).
				Padding(0, 1).
				Width(20)
			projectLines = append(projectLines, "  "+selectedStyle.Render("▶ "+displayText))
		} else {
			// Regular project
			normalStyle := t.S().Base.
				Foreground(t.FgBase).
				Padding(0, 1)
			projectLines = append(projectLines, "    "+normalStyle.Render(displayText))
		}
	}

	// Add navigation hint if there are more projects
	if len(p.availableProjects) > maxShow {
		hint := t.S().Base.Foreground(t.FgMuted).Render(
			fmt.Sprintf("  (%d/%d projects - use ↑/↓ to navigate)",
				p.selectedProjectIndex+1, len(p.availableProjects)))
		projectLines = append(projectLines, "", hint)
	}

	return lipgloss.JoinVertical(lipgloss.Left, projectLines...)
}

func (p *loggingPage) nextProject() {
	if len(p.availableProjects) == 0 {
		return
	}
	p.selectedProjectIndex++
	if p.selectedProjectIndex >= len(p.availableProjects) {
		p.selectedProjectIndex = 0
	}
}

func (p *loggingPage) prevProject() {
	if len(p.availableProjects) == 0 {
		return
	}
	p.selectedProjectIndex--
	if p.selectedProjectIndex < 0 {
		p.selectedProjectIndex = len(p.availableProjects) - 1
	}
}

func (p *loggingPage) nextCustomTicket() {
	if len(p.availableTickets) == 0 {
		return
	}
	p.selectedTicketIdx++
	if p.selectedTicketIdx >= len(p.availableTickets) {
		p.selectedTicketIdx = 0
	}
}

func (p *loggingPage) prevCustomTicket() {
	if len(p.availableTickets) == 0 {
		return
	}
	p.selectedTicketIdx--
	if p.selectedTicketIdx < 0 {
		p.selectedTicketIdx = len(p.availableTickets) - 1
	}
}

func (p *loggingPage) nextCustomDay() {
	if len(p.missingDays) == 0 {
		return
	}
	p.selectedDayIdx++
	if p.selectedDayIdx >= len(p.missingDays) {
		p.selectedDayIdx = 0
	}
}

func (p *loggingPage) prevCustomDay() {
	if len(p.missingDays) == 0 {
		return
	}
	p.selectedDayIdx--
	if p.selectedDayIdx < 0 {
		p.selectedDayIdx = len(p.missingDays) - 1
	}
}

func (p *loggingPage) handleCustomDistributionInput(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "left", "h":
		p.prevCustomDay()
	case "right", "l":
		p.nextCustomDay()
	case "1", "2", "3", "4", "5", "6", "7", "8":
		// Set hours for current ticket/day combination
		hours, _ := strconv.ParseFloat(msg.String(), 64)
		p.setCustomHours(p.selectedTicketIdx, p.selectedDayIdx, hours)
	case "0":
		// Clear hours for current ticket/day combination
		p.setCustomHours(p.selectedTicketIdx, p.selectedDayIdx, 0)
	case "backspace", "delete":
		// Clear current selection
		p.setCustomHours(p.selectedTicketIdx, p.selectedDayIdx, 0)
	}
	return nil
}

func (p *loggingPage) setCustomHours(ticketIdx, dayIdx int, hours float64) {
	if ticketIdx >= len(p.availableTickets) || dayIdx >= len(p.missingDays) {
		return
	}

	ticketKey := p.availableTickets[ticketIdx]
	dayDate := p.missingDays[dayIdx]

	// Initialize ticket map if needed
	if p.customDistribution[ticketKey] == nil {
		p.customDistribution[ticketKey] = make(map[string]float64)
	}

	if hours > 0 {
		p.customDistribution[ticketKey][dayDate] = hours
	} else {
		delete(p.customDistribution[ticketKey], dayDate)
		if len(p.customDistribution[ticketKey]) == 0 {
			delete(p.customDistribution, ticketKey)
		}
	}
}

func (p *loggingPage) getCustomHours(ticketIdx, dayIdx int) float64 {
	if ticketIdx >= len(p.availableTickets) || dayIdx >= len(p.missingDays) {
		return 0
	}

	ticketKey := p.availableTickets[ticketIdx]
	dayDate := p.missingDays[dayIdx]

	if dayHours, exists := p.customDistribution[ticketKey]; exists {
		return dayHours[dayDate]
	}
	return 0
}

func (p *loggingPage) renderCustomDistributionMatrix(t *styles.Theme) string {
	if len(p.availableTickets) == 0 || len(p.missingDays) == 0 {
		return t.S().Base.Foreground(t.FgMuted).Render("No tickets or missing days available")
	}

	var matrixLines []string

	// Add instructions
	instructions := []string{
		"Use ↑/↓ to select ticket, ←/→ to select day",
		"Press 1-8 to set hours, 0 to clear, Enter to submit",
		"",
	}
	for _, instruction := range instructions {
		matrixLines = append(matrixLines, t.S().Base.Foreground(t.FgMuted).Render(instruction))
	}

	// Header with days (show first 7 days max for readability)
	maxDays := 7
	if len(p.missingDays) > maxDays {
		maxDays = len(p.missingDays)
		if maxDays > 10 { // Hard limit
			maxDays = 10
		}
	}

	dayHeaders := []string{"Ticket"}
	for i := 0; i < maxDays && i < len(p.missingDays); i++ {
		dayStr := p.missingDays[i][5:] // Show MM-DD only
		dayHeaders = append(dayHeaders, dayStr)
	}

	// Create header row
	headerCells := []string{}
	for i, header := range dayHeaders {
		width := 8
		if i == 0 { // Ticket column wider
			width = 20
		}

		style := t.S().Base.
			Foreground(t.Blue).
			Bold(true).
			Width(width).
			Align(lipgloss.Center)

		headerCells = append(headerCells, style.Render(header))
	}
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, headerCells...)
	matrixLines = append(matrixLines, headerRow)

	// Separator
	matrixLines = append(matrixLines, t.S().Base.Foreground(t.Border).Render(strings.Repeat("─", 80)))

	// Scrollable ticket list - show 6 tickets at a time
	maxVisibleTickets := 6
	startIdx := 0

	// Calculate scroll position to keep selected ticket visible
	if p.selectedTicketIdx >= maxVisibleTickets {
		startIdx = p.selectedTicketIdx - maxVisibleTickets + 1
	}

	endIdx := startIdx + maxVisibleTickets
	if endIdx > len(p.availableTickets) {
		endIdx = len(p.availableTickets)
		startIdx = endIdx - maxVisibleTickets
		if startIdx < 0 {
			startIdx = 0
		}
	}

	// Matrix rows (scrollable)
	for i := startIdx; i < endIdx; i++ {
		ticketKey := p.availableTickets[i]

		// Get full ticket name and display properly truncated version
		ticketName := ticketKey
		if ticketData, exists := p.ticketLogs[ticketKey]; exists && ticketData.Summary != "" {
			ticketName = ticketData.Summary
			if len(ticketName) > 18 {
				ticketName = ticketName[:15] + "..."
			}
		}

		// Style ticket cell based on selection
		ticketStyle := t.S().Base.Width(20)
		if i == p.selectedTicketIdx {
			ticketStyle = ticketStyle.Foreground(t.Blue).Bold(true)
		} else {
			ticketStyle = ticketStyle.Foreground(t.FgBase)
		}

		rowCells := []string{ticketStyle.Render(ticketName)}

		// Day cells
		for j := 0; j < maxDays && j < len(p.missingDays); j++ {
			hours := p.getCustomHours(i, j)
			hoursText := ""
			if hours > 0 {
				hoursText = fmt.Sprintf("%.1f", hours)
			}

			// Style based on selection and value
			cellStyle := t.S().Base.Width(8).Align(lipgloss.Center)

			if i == p.selectedTicketIdx && j == p.selectedDayIdx {
				// Currently selected cell
				cellStyle = cellStyle.Background(t.Blue).Foreground(t.White).Bold(true)
			} else if hours > 0 {
				// Cell with hours
				cellStyle = cellStyle.Foreground(t.Green).Bold(true)
			} else if i == p.selectedTicketIdx {
				// Selected row
				cellStyle = cellStyle.Foreground(t.Blue)
			} else if j == p.selectedDayIdx {
				// Selected column
				cellStyle = cellStyle.Foreground(t.Citron)
			} else {
				cellStyle = cellStyle.Foreground(t.FgMuted)
			}

			rowCells = append(rowCells, cellStyle.Render(hoursText))
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, rowCells...)
		matrixLines = append(matrixLines, row)
	}

	// Add scroll indicators if needed
	if len(p.availableTickets) > maxVisibleTickets {
		scrollInfo := fmt.Sprintf("Showing tickets %d-%d of %d",
			startIdx+1, endIdx, len(p.availableTickets))

		var scrollIndicators []string
		if startIdx > 0 {
			scrollIndicators = append(scrollIndicators, "▲ More tickets above")
		}
		if endIdx < len(p.availableTickets) {
			scrollIndicators = append(scrollIndicators, "▼ More tickets below")
		}

		if len(scrollIndicators) > 0 {
			matrixLines = append(matrixLines, "")
			for _, indicator := range scrollIndicators {
				matrixLines = append(matrixLines, t.S().Base.Foreground(t.FgMuted).Render("  "+indicator))
			}
		}

		matrixLines = append(matrixLines, "")
		matrixLines = append(matrixLines, t.S().Base.Foreground(t.FgMuted).Render("  "+scrollInfo))
	}

	// Show totals
	matrixLines = append(matrixLines, "")
	totalHours := 0.0
	totalEntries := 0
	for _, dayHours := range p.customDistribution {
		for _, hours := range dayHours {
			if hours > 0 {
				totalHours += hours
				totalEntries++
			}
		}
	}

	totalText := fmt.Sprintf("Total Hours: %.1f (%d entries)", totalHours, totalEntries)
	matrixLines = append(matrixLines, t.S().Base.Foreground(t.Blue).Bold(true).Render(totalText))

	return lipgloss.JoinVertical(lipgloss.Left, matrixLines...)
}

type WorklogSubmittedMsg struct {
	Success bool
	Message string
}

func (p *loggingPage) SetSize(width, height int) tea.Cmd {
	p.width = width
	p.height = height
	shared.LogErrorf("LOGGING_SIZE", "Logging page size set to %dx%d", width, height)

	// Update input widths based on available space
	if width > 100 {
		p.commentInput.SetWidth(60)
		p.ticketKeyInput.SetWidth(30)
	} else if width > 80 {
		p.commentInput.SetWidth(40)
		p.ticketKeyInput.SetWidth(25)
	} else {
		p.commentInput.SetWidth(30)
		p.ticketKeyInput.SetWidth(20)
	}

	return nil
}

func (p *loggingPage) Bindings() []key.Binding {
	return []key.Binding{
		p.keyMap.NextField,
		p.keyMap.PrevField,
		p.keyMap.Submit,
		p.keyMap.Cancel,
		p.keyMap.Clear,
		p.keyMap.Help,
		p.keyMap.Quit,
	}
}

func (p *loggingPage) Help() help.KeyMap {
	var shortList []key.Binding
	var fullList [][]key.Binding

	// Context-aware help based on current mode and focused field
	if p.focusedField == FieldMode {
		shortList = append(shortList,
			key.NewBinding(
				key.WithKeys("↑/↓"),
				key.WithHelp("↑/↓", "change mode"),
			),
		)
	} else if p.focusedField == FieldTicketKey && p.selectedMode == ModeLogToday {
		shortList = append(shortList,
			key.NewBinding(
				key.WithKeys("↑/↓"),
				key.WithHelp("↑/↓", "select ticket"),
			),
		)
	} else if p.focusedField == FieldTicketKey && p.selectedMode == ModeCustomDistribution {
		shortList = append(shortList,
			key.NewBinding(
				key.WithKeys("↑/↓"),
				key.WithHelp("↑/↓", "ticket"),
			),
			key.NewBinding(
				key.WithKeys("←/→"),
				key.WithHelp("←/→", "day"),
			),
			key.NewBinding(
				key.WithKeys("1-8"),
				key.WithHelp("1-8", "set hours"),
			),
			key.NewBinding(
				key.WithKeys("0"),
				key.WithHelp("0", "clear"),
			),
		)
	}

	shortList = append(shortList,
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev field"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "clear"),
		),
		key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c/q", "quit"),
		),
	)

	for _, v := range shortList {
		fullList = append(fullList, []key.Binding{v})
	}

	return core.NewSimpleHelp(shortList, fullList)
}
