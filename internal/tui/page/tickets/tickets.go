package tickets

import (
	"fmt"
	"image/color"
	"strings"

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
	"github.com/charmbracelet/x/ansi"
)

var TicketsPageID page.PageID = "tickets"

type TicketsPage interface {
	util.Model
	layout.Help
}

type ticketsPage struct {
	width, height int
	app           *app.App
	keyMap        KeyMap

	// Ticket data
	ticketLogs   map[string]*shared.TicketWorklog
	ticketKeys   []string // Ordered list of ticket keys for navigation
	selectedIdx  int      // Currently selected ticket index
	
	// Scrolling state
	viewportStart int // First visible row index
	viewportHeight int // Number of visible rows
}

func New(app *app.App) TicketsPage {
	shared.LogErrorf("TICKETS_NEW", "Creating new tickets page")
	
	return &ticketsPage{
		app:           app,
		keyMap:        DefaultKeyMap(),
		selectedIdx:   0,
		viewportStart: 0,
	}
}

func (p *ticketsPage) Init() tea.Cmd {
	shared.LogErrorf("TICKETS_INIT", "Initializing tickets page")
	return nil
}

func (p *ticketsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	shared.LogErrorf("TICKETS_UPDATE", "Received message type: %T", msg)
	
	switch msg := msg.(type) {
	case welcome.WorklogDataMsg:
		shared.LogErrorf("TICKETS_UPDATE", "Received WorklogDataMsg with %d tickets", len(msg.TicketLogs))
		p.SetTicketData(msg.TicketLogs)
		return p, nil
		
	case tea.WindowSizeMsg:
		shared.LogErrorf("TICKETS_UPDATE", "Window resize: %dx%d", msg.Width, msg.Height)
		return p, p.SetSize(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		shared.LogErrorf("TICKETS_UPDATE", "Key press: %s", msg.String())
		switch {
		case key.Matches(msg, p.keyMap.Up):
			if p.selectedIdx > 0 {
				p.selectedIdx--
				p.adjustViewport()
			}
			return p, nil
		case key.Matches(msg, p.keyMap.Down):
			if p.selectedIdx < len(p.ticketKeys)-1 {
				p.selectedIdx++
				p.adjustViewport()
			}
			return p, nil
		case key.Matches(msg, p.keyMap.Select):
			shared.LogErrorf("TICKETS_UPDATE", "Select key pressed for ticket: %s", p.getSelectedTicketKey())
			// TODO: Navigate to ticket detail or log time
			return p, nil
		case key.Matches(msg, p.keyMap.Back):
			shared.LogErrorf("TICKETS_UPDATE", "Back key pressed")
			// Navigate back to stats page
			return p, func() tea.Msg {
				return page.PageChangeMsg{ID: "stats"}
			}
		case key.Matches(msg, p.keyMap.Refresh):
			shared.LogErrorf("TICKETS_UPDATE", "Refresh key pressed")
			// TODO: Refresh ticket data
			return p, nil
		case key.Matches(msg, p.keyMap.LogTime):
			shared.LogErrorf("TICKETS_UPDATE", "Log time key pressed for ticket: %s", p.getSelectedTicketKey())
			// TODO: Open log time dialog
			return p, nil
		case key.Matches(msg, p.keyMap.Quit):
			shared.LogErrorf("TICKETS_UPDATE", "Quit key pressed")
			return p, tea.Quit
		}
	}
	
	return p, nil
}

func (p *ticketsPage) View() string {
	shared.LogErrorf("TICKETS_VIEW", "Rendering tickets view - dimensions: %dx%d", p.width, p.height)
	
	if p.width == 0 || p.height == 0 {
		shared.LogErrorf("TICKETS_VIEW", "Dimensions not set, returning empty view")
		return ""
	}

	t := styles.CurrentTheme()
	shared.LogErrorf("TICKETS_VIEW", "Retrieved current theme")

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
	shared.LogErrorf("TICKETS_VIEW", "Full-width logo created")

	// Create tickets table
	ticketsContent := p.renderSimpleTable(t)
	shared.LogErrorf("TICKETS_VIEW", "Tickets table rendered")

	// Join logo and tickets vertically
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		logoStr,
		"", // Empty line for spacing
		ticketsContent,
	)

	return t.S().Base.
		Width(p.width).
		Height(p.height).
		Padding(1, 2).
		Render(content)
}

func (p *ticketsPage) renderSimpleTable(t *styles.Theme) string {
	shared.LogErrorf("TICKETS_VIEW", "Rendering simple ticket table with %d tickets", len(p.ticketLogs))

	if len(p.ticketKeys) == 0 {
		// Empty state
		emptyStyle := t.S().Base.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary).
			Padding(2, 4).
			Align(lipgloss.Center)

		return emptyStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				t.S().Base.Foreground(t.Primary).Bold(true).Render("No Tickets Found"),
				"",
				t.S().Base.Foreground(t.FgMuted).Render("No ticket data available."),
				t.S().Base.Foreground(t.FgMuted).Render("Try refreshing or check your JIRA connection."),
			),
		)
	}

	// Calculate viewport
	p.calculateViewport()

	// Table container with border
	containerStyle := t.S().Base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1)

	// Calculate column widths for a proper JIRA ticket table
	availableWidth := p.width - 8 // borders and padding
	if availableWidth < 70 {
		availableWidth = 70
	}
	
	keyWidth := 15        // JIRA ticket key (e.g., PT-6165)
	statusWidth := 12     // Status (e.g., In Progress, Done)
	hoursWidth := 8       // Logged hours (e.g., 48.0h)
	priorityWidth := 8    // Priority (e.g., High, Medium)
	summaryWidth := availableWidth - keyWidth - statusWidth - hoursWidth - priorityWidth - 8 // spaces between columns

	// Create header
	header := p.createTableHeader(t, keyWidth, statusWidth, hoursWidth, priorityWidth, summaryWidth)
	
	// Create rows
	var tableRows []string
	tableRows = append(tableRows, header)
	
	// Add separator
	separator := strings.Repeat("─", availableWidth-2)
	styledSeparator := styles.ApplyForegroundGrad(separator, t.Primary, t.Secondary)
	tableRows = append(tableRows, styledSeparator)
	
	// Add data rows
	visibleEnd := p.viewportStart + p.viewportHeight
	if visibleEnd > len(p.ticketKeys) {
		visibleEnd = len(p.ticketKeys)
	}

	for i := p.viewportStart; i < visibleEnd; i++ {
		ticketKey := p.ticketKeys[i]
		ticketData := p.ticketLogs[ticketKey]
		isSelected := i == p.selectedIdx
		
		row := p.createTableRow(t, ticketKey, ticketData, keyWidth, statusWidth, hoursWidth, priorityWidth, summaryWidth, isSelected)
		tableRows = append(tableRows, row)
	}
	
	// Add footer if needed
	if len(p.ticketKeys) > p.viewportHeight {
		tableRows = append(tableRows, "")
		totalItems := len(p.ticketKeys)
		visibleStart := p.viewportStart + 1
		actualVisibleEnd := p.viewportStart + p.viewportHeight
		if actualVisibleEnd > totalItems {
			actualVisibleEnd = totalItems
		}
		
		footerText := fmt.Sprintf("Showing %d-%d of %d tickets", visibleStart, actualVisibleEnd, totalItems)
		footer := t.S().Base.Foreground(t.FgMuted).Align(lipgloss.Center).Render(footerText)
		tableRows = append(tableRows, footer)
	}

	tableContent := lipgloss.JoinVertical(lipgloss.Left, tableRows...)
	return containerStyle.Render(tableContent)
}

func (p *ticketsPage) createTableHeader(t *styles.Theme, keyWidth, statusWidth, hoursWidth, priorityWidth, summaryWidth int) string {
	// Create a proper JIRA ticket table header with all relevant columns
	headerStyle := t.S().Base.Foreground(t.Primary).Bold(true)
	
	keyHeader := headerStyle.Width(keyWidth).Render("TICKET")
	statusHeader := headerStyle.Width(statusWidth).Render("STATUS")
	hoursHeader := headerStyle.Width(hoursWidth).Align(lipgloss.Right).Render("LOGGED")
	priorityHeader := headerStyle.Width(priorityWidth).Render("PRIORITY")
	summaryHeader := headerStyle.Width(summaryWidth).Render("SUMMARY")
	
	return fmt.Sprintf("%s  %s  %s  %s  %s", keyHeader, statusHeader, hoursHeader, priorityHeader, summaryHeader)
}

func (p *ticketsPage) createTableRow(t *styles.Theme, ticketKey string, ticketData *shared.TicketWorklog, keyWidth, statusWidth, hoursWidth, priorityWidth, summaryWidth int, isSelected bool) string {
	// Prepare row data with JIRA-like structure
	key := ansi.Truncate(ticketKey, keyWidth, "…")
	status := p.getTicketStatus(ticketKey) // Derive status from worklog data
	hours := fmt.Sprintf("%.1fh", ticketData.Total)
	priority := p.getTicketPriority(ticketKey) // Derive priority from ticket key patterns
	summary := ansi.Truncate(ticketData.Summary, summaryWidth, "…")
	
	// Apply styling based on selection
	var keyStyle, statusStyle, hoursStyle, priorityStyle, summaryStyle lipgloss.Style
	
	if isSelected {
		// Selected row styling with Crush's secondary color
		baseStyle := t.S().Base.Background(t.Secondary).Foreground(t.White)
		keyStyle = baseStyle.Width(keyWidth).Bold(true)
		statusStyle = baseStyle.Width(statusWidth)
		hoursStyle = baseStyle.Width(hoursWidth).Align(lipgloss.Right)
		priorityStyle = baseStyle.Width(priorityWidth)
		summaryStyle = baseStyle.Width(summaryWidth)
	} else {
		// Normal row styling
		keyStyle = t.S().Base.Width(keyWidth).Foreground(t.Primary)
		statusStyle = t.S().Base.Width(statusWidth).Foreground(p.getStatusColor(status, t))
		hoursStyle = t.S().Base.Width(hoursWidth).Foreground(t.FgBase).Align(lipgloss.Right)
		priorityStyle = t.S().Base.Width(priorityWidth).Foreground(p.getPriorityColor(priority, t))
		summaryStyle = t.S().Base.Width(summaryWidth).Foreground(t.FgBase)
	}
	
	// Render cells
	keyCell := keyStyle.Render(key)
	statusCell := statusStyle.Render(status)
	hoursCell := hoursStyle.Render(hours)
	priorityCell := priorityStyle.Render(priority)
	summaryCell := summaryStyle.Render(summary)
	
	return fmt.Sprintf("%s  %s  %s  %s  %s", keyCell, statusCell, hoursCell, priorityCell, summaryCell)
}

// getTicketStatus derives status based on worklog data
func (p *ticketsPage) getTicketStatus(ticketKey string) string {
	ticketData := p.ticketLogs[ticketKey]
	if ticketData.Total > 0 {
		// If there are recent logs, assume it's in progress
		if len(ticketData.Logs) > 0 {
			return "In Progress"
		}
		return "Done"
	}
	return "To Do"
}

// getTicketPriority derives priority from ticket patterns or defaults
func (p *ticketsPage) getTicketPriority(ticketKey string) string {
	// Simple heuristic: tickets with higher numbers tend to be newer/higher priority
	// In a real implementation, this would come from JIRA API
	if len(p.ticketLogs[ticketKey].Logs) > 3 {
		return "High"
	} else if len(p.ticketLogs[ticketKey].Logs) > 1 {
		return "Medium"  
	}
	return "Low"
}

// getStatusColor returns appropriate color for ticket status
func (p *ticketsPage) getStatusColor(status string, t *styles.Theme) color.Color {
	switch status {
	case "Done":
		return t.Success
	case "In Progress":
		return t.Warning
	case "To Do":
		return t.FgMuted
	default:
		return t.FgBase
	}
}

// getPriorityColor returns appropriate color for ticket priority
func (p *ticketsPage) getPriorityColor(priority string, t *styles.Theme) color.Color {
	switch priority {
	case "High":
		return t.Error
	case "Medium":
		return t.Warning
	case "Low":
		return t.FgMuted
	default:
		return t.FgBase
	}
}


func (p *ticketsPage) SetSize(width, height int) tea.Cmd {
	p.width = width
	p.height = height
	shared.LogErrorf("TICKETS_SIZE", "Tickets page size set to %dx%d", width, height)
	return nil
}

func (p *ticketsPage) SetTicketData(ticketLogs map[string]*shared.TicketWorklog) {
	shared.LogErrorf("TICKETS_DATA", "Setting ticket data - TicketLogs: %d", len(ticketLogs))
	
	p.ticketLogs = ticketLogs
	
	// Create ordered list of ticket keys for navigation
	p.ticketKeys = make([]string, 0, len(ticketLogs))
	for key := range ticketLogs {
		p.ticketKeys = append(p.ticketKeys, key)
	}
	
	// Reset selection to first item
	p.selectedIdx = 0
	
	shared.LogErrorf("TICKETS_DATA", "Ticket data set with %d tickets", len(p.ticketKeys))
}

func (p *ticketsPage) getSelectedTicketKey() string {
	if p.selectedIdx >= 0 && p.selectedIdx < len(p.ticketKeys) {
		return p.ticketKeys[p.selectedIdx]
	}
	return ""
}

func (p *ticketsPage) GetSelectedTicket() *shared.TicketWorklog {
	key := p.getSelectedTicketKey()
	if key != "" {
		return p.ticketLogs[key]
	}
	return nil
}

func (p *ticketsPage) Bindings() []key.Binding {
	return []key.Binding{
		p.keyMap.Up,
		p.keyMap.Down,
		p.keyMap.Select,
		p.keyMap.Back,
		p.keyMap.Refresh,
		p.keyMap.LogTime,
		p.keyMap.Quit,
	}
}

func (p *ticketsPage) Help() help.KeyMap {
	var shortList []key.Binding
	var fullList [][]key.Binding

	shortList = append(shortList,
		key.NewBinding(
			key.WithKeys("↑/↓"),
			key.WithHelp("↑/↓", "navigate"),
		),
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "log time"),
		),
		key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "back"),
		),
		key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
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

// calculateViewport calculates the viewport dimensions and adjusts viewport start
func (p *ticketsPage) calculateViewport() {
	// Calculate available height for table rows (subtract header, footer, borders)
	p.viewportHeight = p.height - 12 // Logo, header, footer, borders
	if p.viewportHeight < 1 {
		p.viewportHeight = 1
	}
}

// adjustViewport ensures the selected item is visible in the viewport
func (p *ticketsPage) adjustViewport() {
	if len(p.ticketKeys) == 0 {
		return
	}

	// Ensure selected item is within viewport
	if p.selectedIdx < p.viewportStart {
		p.viewportStart = p.selectedIdx
	} else if p.selectedIdx >= p.viewportStart+p.viewportHeight {
		p.viewportStart = p.selectedIdx - p.viewportHeight + 1
	}

	// Ensure viewport start is valid
	if p.viewportStart < 0 {
		p.viewportStart = 0
	}
	maxStart := len(p.ticketKeys) - p.viewportHeight
	if maxStart < 0 {
		maxStart = 0
	}
	if p.viewportStart > maxStart {
		p.viewportStart = maxStart
	}
}

