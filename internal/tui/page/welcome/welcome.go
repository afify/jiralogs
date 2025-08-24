package welcome

import (
	"time"

	"LogS/internal/app"
	"LogS/internal/tui/components/core"
	"LogS/internal/tui/components/core/layout"
	"LogS/internal/tui/components/logo"
	"LogS/internal/tui/page"
	"LogS/internal/tui/styles"
	"LogS/internal/tui/util"
	"LogS/shared"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

var WelcomePageID page.PageID = "welcome"

// Messages for worklog fetching
type WorklogFetchStartedMsg struct{}
type WorklogFetchCompletedMsg struct {
	Success      bool
	Error        error
	TicketLogs   map[string]*shared.TicketWorklog
	DailyHours   map[string]float64
	DailyTickets map[string][]string
}

// Message to pass worklog data to stats page
type WorklogDataMsg struct {
	TicketLogs   map[string]*shared.TicketWorklog
	DailyHours   map[string]float64
	DailyTickets map[string][]string
}

type welcomeState int

const (
	stateShowingLogo welcomeState = iota
	stateFetchingWorklogs
	stateComplete
	stateError
)

type WelcomePage interface {
	util.Model
	layout.Help
}

type welcomePage struct {
	width, height int
	app           *app.App
	keyMap        KeyMap
	state         welcomeState
	statusText    string
	autoFetch     bool
	
	// Spinner for loading animation  
	spinner spinner.Model
	
	// Worklog data to pass to stats page
	ticketLogs   map[string]*shared.TicketWorklog
	dailyHours   map[string]float64
	dailyTickets map[string][]string
}

func New(app *app.App) WelcomePage {
	shared.LogErrorf("WELCOME_NEW", "Creating new welcome page")
	isConfigured := app.Config().IsConfigured()
	shared.LogErrorf("WELCOME_NEW", "JIRA configuration status: %v", isConfigured)
	
	// Create spinner with Crush's style
	spinnerModel := spinner.New()
	spinnerModel.Spinner = spinner.Dot
	shared.LogErrorf("WELCOME_NEW", "Spinner created with Crush styling")
	
	page := &welcomePage{
		app:       app,
		keyMap:    DefaultKeyMap(),
		state:     stateShowingLogo,
		autoFetch: true, // Always auto-fetch
		spinner:   spinnerModel,
	}
	
	shared.LogErrorf("WELCOME_NEW", "Welcome page created with state: %d, autoFetch: %v", page.state, page.autoFetch)
	return page
}

func (p *welcomePage) Init() tea.Cmd {
	shared.LogErrorf("WELCOME_INIT", "Initializing welcome page with state: %d", p.state)
	shared.LogErrorf("WELCOME_INIT", "AutoFetch enabled: %v", p.autoFetch)
	
	if p.autoFetch {
		shared.LogErrorf("WELCOME_INIT", "Scheduling immediate worklog fetch")
		// Start fetching worklogs immediately
		return func() tea.Msg {
			shared.LogErrorf("WELCOME_INIT", "Returning WorklogFetchStartedMsg")
			return WorklogFetchStartedMsg{}
		}
	}
	
	shared.LogErrorf("WELCOME_INIT", "AutoFetch disabled, no fetch scheduled")
	return nil
}

func (p *welcomePage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	shared.LogErrorf("WELCOME_UPDATE", "Received message type: %T", msg)
	
	// Update spinner animation when fetching
	if p.state == stateFetchingWorklogs {
		var cmd tea.Cmd
		p.spinner, cmd = p.spinner.Update(msg)
		
		// Handle spinner-specific messages
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			shared.LogErrorf("WELCOME_UPDATE", "Window resize: %dx%d", msg.Width, msg.Height)
			return p, tea.Batch(cmd, p.SetSize(msg.Width, msg.Height))
		case spinner.TickMsg:
			// Just handle spinner animation, don't log every tick
			return p, cmd
		}
	}
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		shared.LogErrorf("WELCOME_UPDATE", "Window resize: %dx%d", msg.Width, msg.Height)
		return p, p.SetSize(msg.Width, msg.Height)

	case WorklogFetchStartedMsg:
		shared.LogErrorf("WELCOME_UPDATE", "Processing WorklogFetchStartedMsg")
		shared.LogErrorf("WELCOME_FETCH", "Starting worklog fetch - changing state from %d to %d", p.state, stateFetchingWorklogs)
		p.state = stateFetchingWorklogs
		p.statusText = "Fetching worklogs from JIRA..."
		shared.LogErrorf("WELCOME_FETCH", "State changed to fetching, status: %s", p.statusText)
		shared.LogErrorf("WELCOME_FETCH", "Starting spinner animation")
		// Start spinner and fetch worklogs  
		return p, tea.Batch(p.spinner.Tick, p.fetchWorklogs())

	case WorklogFetchCompletedMsg:
		shared.LogErrorf("WELCOME_UPDATE", "Processing WorklogFetchCompletedMsg")
		shared.LogErrorf("WELCOME_FETCH", "Worklog fetch completed: success=%v, error=%v", msg.Success, msg.Error)
		if msg.Success {
			shared.LogErrorf("WELCOME_FETCH", "Fetch successful - changing state from %d to %d", p.state, stateComplete)
			p.state = stateComplete
			p.statusText = "Worklogs fetched successfully!"
			
			// Store the worklog data
			p.ticketLogs = msg.TicketLogs
			p.dailyHours = msg.DailyHours
			p.dailyTickets = msg.DailyTickets
			shared.LogErrorf("WELCOME_FETCH", "Stored worklog data - TicketLogs: %d, DailyHours: %d, DailyTickets: %d",
				len(p.ticketLogs), len(p.dailyHours), len(p.dailyTickets))
			
			shared.LogErrorf("WELCOME_FETCH", "Navigating to stats page with worklog data")
			// Navigate to stats page with data
			return p, func() tea.Msg {
				shared.LogErrorf("WELCOME_NAVIGATE", "Sending PageChangeMsg to stats page with worklog data")
				shared.LogErrorf("WELCOME_NAVIGATE", "Worklog data - TicketLogs: %d, DailyHours: %d, DailyTickets: %d",
					len(p.ticketLogs), len(p.dailyHours), len(p.dailyTickets))
				return page.PageChangeMsg{
					ID: "stats",
					Data: WorklogDataMsg{
						TicketLogs:   p.ticketLogs,
						DailyHours:   p.dailyHours,
						DailyTickets: p.dailyTickets,
					},
				}
			}
		} else {
			shared.LogErrorf("WELCOME_FETCH", "Fetch failed - changing state from %d to %d", p.state, stateError)
			p.state = stateError
			if msg.Error != nil {
				p.statusText = "Error: " + msg.Error.Error()
				shared.LogErrorf("WELCOME_FETCH", "Error details: %s", msg.Error.Error())
			} else {
				p.statusText = "Error fetching worklogs"
				shared.LogErrorf("WELCOME_FETCH", "No specific error provided")
			}
		}
		return p, nil

	case tea.KeyPressMsg:
		shared.LogErrorf("WELCOME_UPDATE", "Key press: %s", msg.String())
		switch {
		case key.Matches(msg, p.keyMap.Continue):
			shared.LogErrorf("WELCOME_UPDATE", "Continue key pressed - navigating to stats page")
			// Navigate to stats page
			return p, func() tea.Msg {
				shared.LogErrorf("WELCOME_NAVIGATE", "Manual navigation to stats page")
				return page.PageChangeMsg{ID: "stats"}
			}
		}
	}
	shared.LogErrorf("WELCOME_UPDATE", "No handler for message type %T", msg)
	return p, nil
}

func (p *welcomePage) View() string {
	shared.LogErrorf("WELCOME_VIEW", "Rendering view - dimensions: %dx%d, state: %d", p.width, p.height, p.state)
	
	if p.width == 0 || p.height == 0 {
		shared.LogErrorf("WELCOME_VIEW", "Dimensions not set, returning empty view")
		return ""
	}

	t := styles.CurrentTheme()
	shared.LogErrorf("WELCOME_VIEW", "Retrieved current theme")

	// Create the logo with Crush's exact colors
	logoOpts := logo.Opts{
		FieldColor:   t.Primary,
		TitleColorA:  t.Secondary,
		TitleColorB:  t.Primary,
		CharmColor:   t.Secondary,
		VersionColor: t.Primary,
		Width:        p.width - 4, // Leave some padding
	}

	logoStr := logo.Render("v1.0.0", false, logoOpts)

	// Create centered content
	centeredLogo := t.S().Base.
		Width(p.width).
		Height(p.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(logoStr)

	// Status message and spinner based on state
	shared.LogErrorf("WELCOME_VIEW", "Determining status message for state: %d", p.state)
	var statusContent string
	switch p.state {
	case stateShowingLogo:
		statusMessage := "Press Enter to continue..."
		shared.LogErrorf("WELCOME_VIEW", "State: showing logo, message: %s", statusMessage)
		statusContent = t.S().Base.
			Foreground(t.FgMuted).
			Render(statusMessage)
	case stateFetchingWorklogs:
		shared.LogErrorf("WELCOME_VIEW", "State: fetching worklogs, showing spinner")
		// Show spinner with Crush's beautiful animation
		spinnerView := p.spinner.View()
		statusContent = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Render(spinnerView + " Fetching worklogs from JIRA...")
	case stateComplete:
		statusMessage := p.statusText
		shared.LogErrorf("WELCOME_VIEW", "State: complete, message: %s", statusMessage)
		statusContent = t.S().Base.
			Foreground(t.Success).
			Render(statusMessage)
	case stateError:
		statusMessage := p.statusText
		shared.LogErrorf("WELCOME_VIEW", "State: error, message: %s", statusMessage)
		statusContent = t.S().Base.
			Foreground(t.Error).
			Render(statusMessage)
	}

	// Position content at the bottom center
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		centeredLogo,
		"", // Empty line for spacing
		statusContent,
	)

	return t.S().Base.
		Width(p.width).
		Height(p.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

func (p *welcomePage) fetchWorklogs() tea.Cmd {
	shared.LogErrorf("WELCOME_FETCH", "Creating fetchWorklogs command")
	return func() tea.Msg {
		shared.LogErrorf("WELCOME_FETCH", "Executing fetchWorklogs command")
		
		worklogService := p.app.WorklogService()
		shared.LogErrorf("WELCOME_FETCH", "Retrieved worklog service: %v", worklogService != nil)
		
		if worklogService == nil {
			shared.LogErrorf("WELCOME_FETCH", "No worklog service available - likely JIRA not configured")
			return WorklogFetchCompletedMsg{Success: false, Error: nil}
		}

		// Create a period for year-to-date (beginning of year to today)
		now := time.Now()
		shared.LogErrorf("WELCOME_FETCH", "Current time: %s", now.Format("2006-01-02 15:04:05"))
		
		period := shared.Period{
			Start: time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()),
			End:   now,
		}

		shared.LogErrorf("WELCOME_FETCH", "Created period - Start: %s, End: %s", 
			period.Start.Format("2006-01-02"), period.End.Format("2006-01-02"))

		shared.LogErrorf("WELCOME_FETCH", "Calling worklogService.FetchWorklogs")
		ticketLogs, dailyHours, dailyTickets, err := worklogService.FetchWorklogs(period)
		
		if err != nil {
			shared.LogError("WELCOME_FETCH", err)
			shared.LogErrorf("WELCOME_FETCH", "FetchWorklogs failed with error: %s", err.Error())
			return WorklogFetchCompletedMsg{Success: false, Error: err}
		}

		shared.LogErrorf("WELCOME_FETCH", "FetchWorklogs succeeded")
		shared.LogErrorf("WELCOME_FETCH", "Results - TicketLogs: %d, DailyHours entries: %d, DailyTickets entries: %d", 
			len(ticketLogs), len(dailyHours), len(dailyTickets))
		
		return WorklogFetchCompletedMsg{
			Success:      true, 
			Error:        nil,
			TicketLogs:   ticketLogs,
			DailyHours:   dailyHours,
			DailyTickets: dailyTickets,
		}
	}
}

func (p *welcomePage) SetSize(width, height int) tea.Cmd {
	p.width = width
	p.height = height
	shared.LogErrorf("WELCOME_SIZE", "Welcome page size set to %dx%d", width, height)
	return nil
}

func (p *welcomePage) Bindings() []key.Binding {
	return []key.Binding{
		p.keyMap.Continue,
	}
}

func (p *welcomePage) Help() help.KeyMap {
	var shortList []key.Binding
	var fullList [][]key.Binding

	shortList = append(shortList,
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "continue"),
		),
		key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	)

	for _, v := range shortList {
		fullList = append(fullList, []key.Binding{v})
	}

	return core.NewSimpleHelp(shortList, fullList)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
