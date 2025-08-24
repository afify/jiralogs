package stats

import (
	"fmt"
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
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

var StatsPageID page.PageID = "stats"

type StatsPage interface {
	util.Model
	layout.Help
}

type statsPage struct {
	width, height int
	app           *app.App
	keyMap        KeyMap
	
	// User data
	userEmail string
	userID    string
	
	// Period data
	currentPeriod shared.Period
	daysLogged    int
	totalDays     int
	
	// Stats data
	ticketLogs   map[string]*shared.TicketWorklog
	dailyHours   map[string]float64
	dailyTickets map[string][]string
	totalHours   float64
	totalTickets int
}

func New(app *app.App) StatsPage {
	shared.LogErrorf("STATS_NEW", "Creating new stats page")
	
	// Get user data from config and JIRA
	config := app.Config()
	userEmail := config.JiraEmail
	userID := ""
	
	// Try to get user ID from JIRA
	if app.JiraClient() != nil {
		if user, err := app.JiraClient().GetCurrentUser(); err == nil {
			userID = user.AccountID
			shared.LogErrorf("STATS_NEW", "Got user ID from JIRA: %s", userID)
		} else {
			shared.LogError("STATS_NEW", err)
			userID = "N/A"
		}
	} else {
		userID = "N/A"
	}
	
	// Set default period: beginning of year to today
	now := time.Now()
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	defaultPeriod := shared.Period{
		Start: yearStart,
		End:   now,
	}
	
	shared.LogErrorf("STATS_NEW", "User: %s, UserID: %s, Period: %s to %s", userEmail, userID,
		defaultPeriod.Start.Format("2006-01-02"), defaultPeriod.End.Format("2006-01-02"))
	
	return &statsPage{
		app:           app,
		keyMap:        DefaultKeyMap(),
		userEmail:     userEmail,
		userID:        userID,
		currentPeriod: defaultPeriod,
	}
}

func (p *statsPage) Init() tea.Cmd {
	shared.LogErrorf("STATS_INIT", "Initializing stats page")
	return nil
}

func (p *statsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	shared.LogErrorf("STATS_UPDATE", "Received message type: %T", msg)
	
	switch msg := msg.(type) {
	case welcome.WorklogDataMsg:
		shared.LogErrorf("STATS_UPDATE", "Received WorklogDataMsg")
		p.SetWorklogData(msg.TicketLogs, msg.DailyHours, msg.DailyTickets)
		return p, nil
		
	case tea.WindowSizeMsg:
		shared.LogErrorf("STATS_UPDATE", "Window resize: %dx%d", msg.Width, msg.Height)
		return p, p.SetSize(msg.Width, msg.Height)
	
	case tea.KeyPressMsg:
		shared.LogErrorf("STATS_UPDATE", "Key press: %s", msg.String())
		switch {
		case key.Matches(msg, p.keyMap.LogTime):
			shared.LogErrorf("STATS_UPDATE", "Log time key pressed")
			// TODO: Open log time dialog
			return p, nil
		case key.Matches(msg, p.keyMap.Refresh):
			shared.LogErrorf("STATS_UPDATE", "Refresh key pressed")
			// TODO: Refresh worklog data
			return p, nil
		case key.Matches(msg, p.keyMap.ListTickets):
			shared.LogErrorf("STATS_UPDATE", "List tickets key pressed")
			// Navigate to tickets page with current data
			worklogData := welcome.WorklogDataMsg{
				TicketLogs:   p.ticketLogs,
				DailyHours:   p.dailyHours,
				DailyTickets: p.dailyTickets,
			}
			return p, func() tea.Msg {
				return page.PageChangeMsg{
					ID:   "tickets",
					Data: worklogData,
				}
			}
		case key.Matches(msg, p.keyMap.Calendar):
			shared.LogErrorf("STATS_UPDATE", "Calendar key pressed")
			// Navigate to calendar page with current data
			worklogData := welcome.WorklogDataMsg{
				TicketLogs:   p.ticketLogs,
				DailyHours:   p.dailyHours,
				DailyTickets: p.dailyTickets,
			}
			return p, func() tea.Msg {
				return page.PageChangeMsg{
					ID:   "calendar",
					Data: worklogData,
				}
			}
		case key.Matches(msg, p.keyMap.Quit):
			shared.LogErrorf("STATS_UPDATE", "Quit key pressed")
			return p, tea.Quit
		}
	}
	
	shared.LogErrorf("STATS_UPDATE", "No handler for message type %T", msg)
	return p, nil
}

func (p *statsPage) View() string {
	shared.LogErrorf("STATS_VIEW", "Rendering stats view - dimensions: %dx%d", p.width, p.height)
	
	if p.width == 0 || p.height == 0 {
		shared.LogErrorf("STATS_VIEW", "Dimensions not set, returning empty view")
		return ""
	}

	t := styles.CurrentTheme()
	shared.LogErrorf("STATS_VIEW", "Retrieved current theme")

	// Create the full-width top logo with Crush's exact colors
	logoOpts := logo.Opts{
		FieldColor:   t.Primary,
		TitleColorA:  t.Secondary,
		TitleColorB:  t.Primary,
		CharmColor:   t.Secondary,
		VersionColor: t.Primary,
		Width:        p.width - 4,
	}
	
	logoStr := logo.Render("v1.0.0", false, logoOpts) // false for full logo
	shared.LogErrorf("STATS_VIEW", "Full-width logo created")

	// Create stats content
	statsContent := p.renderStats(t)
	shared.LogErrorf("STATS_VIEW", "Stats content rendered")

	// Join logo and stats vertically
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		logoStr,
		"", // Empty line for spacing
		statsContent,
	)

	return t.S().Base.
		Width(p.width).
		Height(p.height).
		Padding(1, 2).
		Render(content)
}

func (p *statsPage) renderStats(t *styles.Theme) string {
	shared.LogErrorf("STATS_VIEW", "Rendering stats data")

	// User Information Section
	userInfoStyle := t.S().Base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1, 2)

	userInfo := userInfoStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			t.S().Base.Foreground(t.Primary).Bold(true).Render("User Information"),
			"",
			t.S().Base.Foreground(t.FgMuted).Render("Email: ") + t.S().Base.Foreground(t.FgBase).Render(p.userEmail),
			t.S().Base.Foreground(t.FgMuted).Render("User ID: ") + t.S().Base.Foreground(t.FgBase).Render(p.userID),
		),
	)

	// Period Information Section with Progress Bar
	periodInfoStyle := t.S().Base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1, 2)

	// Create progress bar for days logged with Crush's beautiful gradient styling
	var daysProgressPercentage float64
	if p.totalDays > 0 {
		daysProgressPercentage = float64(p.daysLogged) / float64(p.totalDays) * 100
	} else {
		daysProgressPercentage = 0
	}
	progressBarWidth := 30
	filledWidth := int(float64(progressBarWidth) * daysProgressPercentage / 100)
	
	// Create the filled and empty parts separately
	filledPart := ""
	emptyPart := ""
	
	for i := 0; i < filledWidth && i < progressBarWidth; i++ {
		filledPart += "█"
	}
	
	for i := filledWidth; i < progressBarWidth; i++ {
		emptyPart += "░"
	}
	
	// Apply Crush's gradient styling to the filled portion and muted color to empty portion
	var styledProgressBar string
	if len(filledPart) > 0 {
		// Apply gradient to filled part using Crush's ApplyForegroundGrad
		styledFilled := styles.ApplyForegroundGrad(filledPart, t.Primary, t.Secondary)
		styledEmpty := t.S().Base.Foreground(t.Border).Render(emptyPart)
		styledProgressBar = styledFilled + styledEmpty
	} else {
		// All empty - use muted border color
		styledProgressBar = t.S().Base.Foreground(t.Border).Render(emptyPart)
	}

	periodInfo := periodInfoStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			t.S().Base.Foreground(t.Primary).Bold(true).Render("Period Information"),
			"",
			t.S().Base.Foreground(t.FgMuted).Render("From: ") + t.S().Base.Foreground(t.FgBase).Render(p.currentPeriod.Start.Format("2006-01-02")),
			t.S().Base.Foreground(t.FgMuted).Render("To: ") + t.S().Base.Foreground(t.FgBase).Render(p.currentPeriod.End.Format("2006-01-02")),
			"",
			t.S().Base.Foreground(t.FgMuted).Render("Days Progress:"),
			styledProgressBar,
			t.S().Base.Foreground(t.FgBase).Render(fmt.Sprintf("%d/%d days", p.daysLogged, p.totalDays)) + 
				t.S().Base.Foreground(t.FgMuted).Render(fmt.Sprintf(" (%.0f%%)", daysProgressPercentage)),
		),
	)

	// Overall Progress Section
	progressStyle := t.S().Base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1, 2)

	progress := progressStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			t.S().Base.Foreground(t.Primary).Bold(true).Render("Overall Progress"),
			"",
			t.S().Base.Foreground(t.FgMuted).Render("Total Hours: ") + t.S().Base.Foreground(t.FgBase).Render(fmt.Sprintf("%.1f", p.totalHours)),
			t.S().Base.Foreground(t.FgMuted).Render("Total Tickets: ") + t.S().Base.Foreground(t.FgBase).Render(fmt.Sprintf("%d", p.totalTickets)),
			"",
			t.S().Base.Foreground(t.FgMuted).Render("Completion: ") + t.S().Base.Foreground(t.Primary).Bold(true).Render(fmt.Sprintf("%.0f%%", daysProgressPercentage)),
		),
	)

	// Combine all sections
	return lipgloss.JoinVertical(
		lipgloss.Left,
		userInfo,
		"",
		periodInfo,
		"",
		progress,
	)
}

func (p *statsPage) SetSize(width, height int) tea.Cmd {
	p.width = width
	p.height = height
	shared.LogErrorf("STATS_SIZE", "Stats page size set to %dx%d", width, height)
	return nil
}

func (p *statsPage) SetWorklogData(ticketLogs map[string]*shared.TicketWorklog, dailyHours map[string]float64, dailyTickets map[string][]string) {
	shared.LogErrorf("STATS_DATA", "Setting worklog data - TicketLogs: %d, DailyHours: %d, DailyTickets: %d", 
		len(ticketLogs), len(dailyHours), len(dailyTickets))
	
	p.ticketLogs = ticketLogs
	p.dailyHours = dailyHours
	p.dailyTickets = dailyTickets
	
	// Calculate totals
	p.totalHours = 0
	for _, hours := range dailyHours {
		p.totalHours += hours
	}
	
	p.totalTickets = len(ticketLogs)
	
	// Calculate days logged and total days
	p.daysLogged = len(dailyHours) // Days with logged hours
	
	// Calculate total working days (excluding weekends - Friday/Saturday for Middle East)
	totalDays := 0
	current := p.currentPeriod.Start
	for !current.After(p.currentPeriod.End) {
		// Skip Friday (5) and Saturday (6) - Middle East weekend
		if current.Weekday() != time.Friday && current.Weekday() != time.Saturday {
			totalDays++
		}
		current = current.AddDate(0, 0, 1)
	}
	p.totalDays = totalDays
	
	shared.LogErrorf("STATS_DATA", "Calculated totals - Hours: %.1f, Tickets: %d, DaysLogged: %d, TotalDays: %d", 
		p.totalHours, p.totalTickets, p.daysLogged, p.totalDays)
}

func (p *statsPage) Bindings() []key.Binding {
	return []key.Binding{
		p.keyMap.LogTime,
		p.keyMap.Refresh,
		p.keyMap.ListTickets,
		p.keyMap.Calendar,
		p.keyMap.Quit,
	}
}

func (p *statsPage) Help() help.KeyMap {
	var shortList []key.Binding
	var fullList [][]key.Binding

	shortList = append(shortList,
		key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "log time"),
		),
		key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "list tickets"),
		),
		key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "calendar"),
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