package daydetails

import (
	"fmt"
	"strings"
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
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

var DayDetailsPageID page.PageID = "daydetails"

type DayDetailsPage interface {
	util.Model
	layout.Help
}

type WorklogEntry struct {
	TicketKey string
	WorklogID string
	Hours     float64
	Comment   string
	Summary   string // Ticket summary
}

type DayDetailsData struct {
	Date           time.Time
	WorklogEntries []WorklogEntry
	TotalHours     float64
}

type dayDetailsPage struct {
	width, height int
	app           *app.App
	keyMap        KeyMap

	// Day data
	selectedDate   time.Time
	worklogEntries []WorklogEntry
	totalHours     float64

	// View state
	selectedEntryIdx int
	viewportStart    int
	viewportHeight   int
}

func New(app *app.App) DayDetailsPage {
	shared.LogErrorf("DAYDETAILS_NEW", "Creating new day details page")

	return &dayDetailsPage{
		app:              app,
		keyMap:           DefaultKeyMap(),
		selectedEntryIdx: 0,
		viewportStart:    0,
	}
}

func (p *dayDetailsPage) Init() tea.Cmd {
	shared.LogErrorf("DAYDETAILS_INIT", "Initializing day details page")
	return nil
}

func (p *dayDetailsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	shared.LogErrorf("DAYDETAILS_UPDATE", "Received message type: %T", msg)

	switch msg := msg.(type) {
	case DayDetailsData:
		shared.LogErrorf("DAYDETAILS_UPDATE", "Received DayDetailsData for date: %s", msg.Date.Format("2006-01-02"))
		p.SetDayData(msg)
		return p, nil

	case tea.WindowSizeMsg:
		shared.LogErrorf("DAYDETAILS_UPDATE", "Window resize: %dx%d", msg.Width, msg.Height)
		return p, p.SetSize(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		shared.LogErrorf("DAYDETAILS_UPDATE", "Key press: %s", msg.String())
		switch {
		case key.Matches(msg, p.keyMap.Up):
			if p.selectedEntryIdx > 0 {
				p.selectedEntryIdx--
				p.adjustViewport()
			}
			return p, nil

		case key.Matches(msg, p.keyMap.Down):
			if p.selectedEntryIdx < len(p.worklogEntries)-1 {
				p.selectedEntryIdx++
				p.adjustViewport()
			}
			return p, nil

		case key.Matches(msg, p.keyMap.Edit):
			if len(p.worklogEntries) > 0 && p.selectedEntryIdx < len(p.worklogEntries) {
				shared.LogErrorf("DAYDETAILS_UPDATE", "Edit worklog key pressed for entry: %d", p.selectedEntryIdx)
				return p, p.editWorklogEntry()
			}
			return p, nil

		case key.Matches(msg, p.keyMap.Delete):
			if len(p.worklogEntries) > 0 && p.selectedEntryIdx < len(p.worklogEntries) {
				shared.LogErrorf("DAYDETAILS_UPDATE", "Delete worklog key pressed for entry: %d", p.selectedEntryIdx)
				return p, p.deleteWorklogEntry()
			}
			return p, nil

		case key.Matches(msg, p.keyMap.Back):
			shared.LogErrorf("DAYDETAILS_UPDATE", "Back key pressed")
			// Navigate back to calendar page
			return p, func() tea.Msg {
				return page.PageChangeMsg{ID: "calendar"}
			}

		case key.Matches(msg, p.keyMap.Quit):
			shared.LogErrorf("DAYDETAILS_UPDATE", "Quit key pressed")
			return p, tea.Quit
		}
	}

	return p, nil
}

func (p *dayDetailsPage) View() string {
	shared.LogErrorf("DAYDETAILS_VIEW", "Rendering day details view - dimensions: %dx%d", p.width, p.height)

	if p.width == 0 || p.height == 0 {
		shared.LogErrorf("DAYDETAILS_VIEW", "Dimensions not set, returning empty view")
		return ""
	}

	t := styles.CurrentTheme()
	shared.LogErrorf("DAYDETAILS_VIEW", "Retrieved current theme")

	// Create the full-width top logo
	logoOpts := logo.Opts{
		FieldColor:   t.Primary,
		TitleColorA:  t.Secondary,
		TitleColorB:  t.Primary,
		CharmColor:   t.Secondary,
		VersionColor: t.Primary,
		Width:        p.width - 4,
	}

	logoStr := logo.Render(false, logoOpts)
	shared.LogErrorf("DAYDETAILS_VIEW", "Full-width logo created")

	// Create day details content
	detailsContent := p.renderDayDetails(t)
	shared.LogErrorf("DAYDETAILS_VIEW", "Day details content rendered")

	// Join logo and details vertically
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		logoStr,
		"", // Empty line for spacing
		detailsContent,
	)

	return t.S().Base.
		Width(p.width).
		Height(p.height).
		Padding(1, 2).
		Render(content)
}

func (p *dayDetailsPage) renderDayDetails(t *styles.Theme) string {
	// Date header
	dateStr := p.selectedDate.Format("Monday, January 2, 2006")
	dayOfWeek := p.selectedDate.Weekday().String()

	headerStyle := t.S().Base.
		Foreground(t.Primary).
		Bold(true).
		Align(lipgloss.Center).
		Width(p.width - 8)

	header := headerStyle.Render(fmt.Sprintf("Worklogs for %s", dateStr))

	// Summary info
	summaryStyle := t.S().Base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1, 2).
		Width(p.width - 8)

	var status string
	var statusColor = t.FgBase
	if p.selectedDate.Weekday() == time.Friday || p.selectedDate.Weekday() == time.Saturday {
		status = "Weekend (No work required)"
		statusColor = t.FgMuted
	} else if p.totalHours >= 8.0 {
		status = "✓ Full day logged"
		statusColor = t.Success
	} else if p.totalHours > 0 {
		status = fmt.Sprintf("⚠ Partial day (%.1f/8.0 hours)", p.totalHours)
		statusColor = t.Warning
	} else {
		status = "✗ No hours logged"
		statusColor = t.Error
	}

	summaryContent := lipgloss.JoinVertical(
		lipgloss.Left,
		t.S().Base.Foreground(t.FgMuted).Render("Date: ")+t.S().Base.Foreground(t.FgBase).Render(dateStr),
		t.S().Base.Foreground(t.FgMuted).Render("Day: ")+t.S().Base.Foreground(t.FgBase).Render(dayOfWeek),
		t.S().Base.Foreground(t.FgMuted).Render("Total Hours: ")+t.S().Base.Foreground(t.Primary).Bold(true).Render(fmt.Sprintf("%.1f", p.totalHours)),
		t.S().Base.Foreground(t.FgMuted).Render("Status: ")+t.S().Base.Foreground(statusColor).Render(status),
	)

	summary := summaryStyle.Render(summaryContent)

	// Worklog entries list
	worklogList := p.renderWorklogList(t)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		summary,
		"",
		worklogList,
	)
}

func (p *dayDetailsPage) renderWorklogList(t *styles.Theme) string {
	if len(p.worklogEntries) == 0 {
		emptyStyle := t.S().Base.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			Padding(2, 4).
			Align(lipgloss.Center)

		return emptyStyle.Render(
			t.S().Base.Foreground(t.FgMuted).Render("No worklogs for this day"),
		)
	}

	// Calculate viewport
	p.calculateViewport()

	// Table container
	containerStyle := t.S().Base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1)

	// Table header
	headerStyle := t.S().Base.Foreground(t.Primary).Bold(true)

	header := fmt.Sprintf("%s  %s  %s  %s",
		headerStyle.Width(15).Render("TICKET"),
		headerStyle.Width(8).Align(lipgloss.Right).Render("HOURS"),
		headerStyle.Width(25).Render("COMMENT"),
		headerStyle.Width(30).Render("SUMMARY"),
	)

	// Separator
	separator := strings.Repeat("─", p.width-12)
	styledSeparator := styles.ApplyForegroundGrad(separator, t.Primary, t.Secondary)

	// Build table rows
	var tableRows []string
	tableRows = append(tableRows, header, styledSeparator)

	// Add data rows
	visibleEnd := p.viewportStart + p.viewportHeight
	if visibleEnd > len(p.worklogEntries) {
		visibleEnd = len(p.worklogEntries)
	}

	for i := p.viewportStart; i < visibleEnd; i++ {
		entry := p.worklogEntries[i]
		isSelected := i == p.selectedEntryIdx

		row := p.renderWorklogRow(t, entry, isSelected)
		tableRows = append(tableRows, row)
	}

	// Add footer if scrollable
	if len(p.worklogEntries) > p.viewportHeight {
		tableRows = append(tableRows, "")
		footerText := fmt.Sprintf("Showing %d-%d of %d worklogs",
			p.viewportStart+1,
			visibleEnd,
			len(p.worklogEntries))
		footer := t.S().Base.Foreground(t.FgMuted).Align(lipgloss.Center).Render(footerText)
		tableRows = append(tableRows, footer)
	}

	tableContent := lipgloss.JoinVertical(lipgloss.Left, tableRows...)
	return containerStyle.Render(tableContent)
}

func (p *dayDetailsPage) renderWorklogRow(t *styles.Theme, entry WorklogEntry, isSelected bool) string {
	// Truncate long text
	comment := entry.Comment
	if len(comment) > 25 {
		comment = comment[:22] + "..."
	}

	summary := entry.Summary
	if len(summary) > 30 {
		summary = summary[:27] + "..."
	}

	// Apply styling based on selection
	var keyStyle, hoursStyle, commentStyle, summaryStyle lipgloss.Style

	if isSelected {
		baseStyle := t.S().Base.Background(t.Secondary).Foreground(t.White)
		keyStyle = baseStyle.Width(15).Bold(true)
		hoursStyle = baseStyle.Width(8).Align(lipgloss.Right)
		commentStyle = baseStyle.Width(25)
		summaryStyle = baseStyle.Width(30)
	} else {
		keyStyle = t.S().Base.Width(15).Foreground(t.Primary)
		hoursStyle = t.S().Base.Width(8).Foreground(t.FgBase).Align(lipgloss.Right)
		commentStyle = t.S().Base.Width(25).Foreground(t.FgMuted)
		summaryStyle = t.S().Base.Width(30).Foreground(t.FgBase)
	}

	return fmt.Sprintf("%s  %s  %s  %s",
		keyStyle.Render(entry.TicketKey),
		hoursStyle.Render(fmt.Sprintf("%.1fh", entry.Hours)),
		commentStyle.Render(comment),
		summaryStyle.Render(summary),
	)
}

func (p *dayDetailsPage) SetSize(width, height int) tea.Cmd {
	p.width = width
	p.height = height
	shared.LogErrorf("DAYDETAILS_SIZE", "Day details page size set to %dx%d", width, height)
	return nil
}

func (p *dayDetailsPage) SetDayData(data DayDetailsData) {
	p.selectedDate = data.Date
	p.worklogEntries = data.WorklogEntries
	p.totalHours = data.TotalHours
	p.selectedEntryIdx = 0
	p.viewportStart = 0

	shared.LogErrorf("DAYDETAILS_DATA", "Set day data for %s - %d worklog entries, %.1f hours",
		data.Date.Format("2006-01-02"), len(data.WorklogEntries), data.TotalHours)
}

func (p *dayDetailsPage) Bindings() []key.Binding {
	return []key.Binding{
		p.keyMap.Up,
		p.keyMap.Down,
		p.keyMap.Edit,
		p.keyMap.Delete,
		p.keyMap.Back,
		p.keyMap.Quit,
	}
}

func (p *dayDetailsPage) Help() help.KeyMap {
	var shortList []key.Binding
	var fullList [][]key.Binding

	shortList = append(shortList,
		key.NewBinding(
			key.WithKeys("↑/↓"),
			key.WithHelp("↑/↓", "navigate"),
		),
		key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit worklog"),
		),
		key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete worklog"),
		),
		key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "back"),
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

// calculateViewport calculates the viewport dimensions
func (p *dayDetailsPage) calculateViewport() {
	// Calculate available height for table rows (subtract header, footer, borders)
	p.viewportHeight = p.height - 16 // Logo, headers, summary, borders
	if p.viewportHeight < 1 {
		p.viewportHeight = 1
	}
}

// adjustViewport ensures the selected item is visible in the viewport
func (p *dayDetailsPage) adjustViewport() {
	if len(p.worklogEntries) == 0 {
		return
	}

	// Ensure selected item is within viewport
	if p.selectedEntryIdx < p.viewportStart {
		p.viewportStart = p.selectedEntryIdx
	} else if p.selectedEntryIdx >= p.viewportStart+p.viewportHeight {
		p.viewportStart = p.selectedEntryIdx - p.viewportHeight + 1
	}

	// Ensure viewport start is valid
	if p.viewportStart < 0 {
		p.viewportStart = 0
	}
	maxStart := len(p.worklogEntries) - p.viewportHeight
	if maxStart < 0 {
		maxStart = 0
	}
	if p.viewportStart > maxStart {
		p.viewportStart = maxStart
	}
}

// editWorklogEntry handles editing the selected worklog entry
func (p *dayDetailsPage) editWorklogEntry() tea.Cmd {
	if p.selectedEntryIdx >= len(p.worklogEntries) {
		return nil
	}

	entry := p.worklogEntries[p.selectedEntryIdx]
	shared.LogErrorf("DAYDETAILS_EDIT", "Editing worklog entry: %s - %s (%.1fh)", entry.TicketKey, entry.WorklogID, entry.Hours)

	// For now, just log the action - you can implement an edit dialog/form here
	// This could navigate to an edit form page or show an inline editor
	return func() tea.Msg {
		// TODO: Implement edit form navigation
		// For now, just return nil to stay on current page
		return nil
	}
}

// deleteWorklogEntry handles deleting the selected worklog entry
func (p *dayDetailsPage) deleteWorklogEntry() tea.Cmd {
	if p.selectedEntryIdx >= len(p.worklogEntries) {
		return nil
	}

	entry := p.worklogEntries[p.selectedEntryIdx]
	shared.LogErrorf("DAYDETAILS_DELETE", "Deleting worklog entry: %s - %s (%.1fh)", entry.TicketKey, entry.WorklogID, entry.Hours)

	// Call JIRA API to delete the worklog
	return func() tea.Msg {
		jiraClient := p.app.JiraClient()
		err := jiraClient.DeleteWorklog(entry.TicketKey, entry.WorklogID)
		if err != nil {
			shared.LogError("DAYDETAILS_DELETE", err)
			// TODO: Return error message to display to user
			return nil
		}

		// TODO: Refresh the day details data after successful deletion
		// For now, just log success
		shared.LogErrorf("DAYDETAILS_DELETE", "Successfully deleted worklog %s from %s", entry.WorklogID, entry.TicketKey)
		return nil
	}
}
