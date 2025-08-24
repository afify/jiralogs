package tui

import (
	"fmt"
	"strings"
	"time"

	"LogS/shared"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define gradient colors for a modern look
var (
	// Use gradient colors from shared package
	gradientStart = shared.GradientStart
	gradientMid   = shared.GradientMid
	gradientEnd   = shared.GradientEnd

	// Base card style with modern look
	baseCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 3).
			MarginRight(2).
			MarginBottom(1)

	// Header styles for cards with better contrast
	cardHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	cardValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 0, 0, 0)

	cardSubtextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E0E0E0")).
				Italic(true)

	// Section title style with gradient underline
	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(shared.PrimaryColor).
				MarginBottom(1)

	// Beautiful badge styles with proper contrast
	completeBadge = lipgloss.NewStyle().
			Background(shared.Card2GradientStart).
			Foreground(shared.White).
			Bold(true).
			Padding(0, 2).
			MarginRight(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(shared.Card2GradientMid)

	partialBadge = lipgloss.NewStyle().
			Background(shared.WarningColor).
			Foreground(shared.BgDark).
			Bold(true).
			Padding(0, 2).
			MarginRight(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(shared.Card4GradientMid)

	missingBadge = lipgloss.NewStyle().
			Background(shared.ErrorColor).
			Foreground(shared.White).
			Bold(true).
			Padding(0, 2).
			MarginRight(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(shared.WarmGradientStart)
)

// RenderStatsPage renders the main stats dashboard with modern design
func (m *AppModel) RenderStatsPage() string {
	// Wait for terminal size
	if m.width == 0 || m.height == 0 {
		return "\n  Loading stats..."
	}

	shared.LogError("STATS_PAGE_START", fmt.Errorf("Rendering stats page - Width: %d, Height: %d", m.width, m.height))

	var content []string

	// Stats cards row (no title needed, shown in app.go)
	statsCards := m.renderStatsCards()
	cardsHeight := strings.Count(statsCards, "\n") + 1
	shared.LogError("STATS_CARDS", fmt.Errorf("Stats cards height: %d lines", cardsHeight))
	content = append(content, statsCards)
	content = append(content, "")

	// Progress visualization
	progressSection := m.renderAnimatedProgress()
	progressHeight := strings.Count(progressSection, "\n") + 1
	shared.LogError("STATS_PROGRESS", fmt.Errorf("Progress section height: %d lines", progressHeight))
	content = append(content, progressSection)

	// Join all content
	fullContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	totalStatsHeight := strings.Count(fullContent, "\n") + 1
	shared.LogError("STATS_TOTAL", fmt.Errorf("Total stats content height: %d lines", totalStatsHeight))

	// Apply padding and return
	result := lipgloss.NewStyle().
		Padding(0, 2).
		Render(fullContent)

	finalStatsHeight := strings.Count(result, "\n") + 1
	shared.LogError("STATS_FINAL", fmt.Errorf("Final stats height with padding: %d lines", finalStatsHeight))

	return result
}

// renderStatsCards creates beautiful stat cards with gradients
func (m *AppModel) renderStatsCards() string {
	if m.summary == nil {
		return InfoStyle.Render("📊 Loading statistics...")
	}

	// Calculate available width and card sizes
	availableWidth := m.width - 8     // Account for padding
	cardWidth := availableWidth/4 - 2 // Divide equally among 4 cards with margins
	if cardWidth < 20 {
		cardWidth = 20
	}

	cards := []string{}

	// Period Card with purple gradient
	periodCard := baseCardStyle.
		BorderForeground(shared.Card1GradientMid).
		Background(shared.Card1GradientStart).
		Width(cardWidth).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				cardHeaderStyle.Copy().Foreground(shared.White).Render("📅 PERIOD"),
				cardValueStyle.Copy().Foreground(shared.BrightText).Render(m.summary.StartDate),
				cardSubtextStyle.Copy().Foreground(shared.LightGray).Render("to"),
				cardValueStyle.Copy().Foreground(shared.BrightText).Render(m.summary.EndDate),
			),
		)
	cards = append(cards, periodCard)

	// Days Card with green gradient
	daysRatio := float64(m.summary.LoggedDays) / float64(m.summary.Workdays)
	daysColor := shared.Card2GradientStart
	if daysRatio < 0.8 {
		daysColor = shared.WarningColor
	}
	if daysRatio < 0.5 {
		daysColor = shared.ErrorColor
	}

	daysCard := baseCardStyle.
		BorderForeground(shared.Card2GradientMid).
		Background(daysColor).
		Width(cardWidth).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				cardHeaderStyle.Copy().Foreground(shared.White).Render("📊 DAYS"),
				cardValueStyle.Copy().Foreground(shared.BrightText).Bold(true).Render(fmt.Sprintf("%d/%d", m.summary.LoggedDays, m.summary.Workdays)),
				cardSubtextStyle.Copy().Foreground(shared.LightGray).Render(fmt.Sprintf("%.0f%% logged", daysRatio*100)),
			),
		)
	cards = append(cards, daysCard)

	// Hours Card with cyan gradient
	hoursRatio := m.summary.LoggedHours / m.summary.RequiredHours
	hoursColor := shared.Card3GradientStart
	if hoursRatio < 0.8 {
		hoursColor = shared.WarningColor
	}
	if hoursRatio < 0.5 {
		hoursColor = shared.ErrorColor
	}

	hoursCard := baseCardStyle.
		BorderForeground(shared.Card3GradientMid).
		Background(hoursColor).
		Width(cardWidth).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				cardHeaderStyle.Copy().Foreground(shared.White).Render("⏱️  HOURS"),
				cardValueStyle.Copy().Foreground(shared.BrightText).Bold(true).Render(fmt.Sprintf("%.1fh", m.summary.LoggedHours)),
				cardSubtextStyle.Copy().Foreground(shared.LightGray).Render(fmt.Sprintf("of %.0fh", m.summary.RequiredHours)),
			),
		)
	cards = append(cards, hoursCard)

	// Tickets Card with orange gradient
	ticketsCard := baseCardStyle.
		BorderForeground(shared.Card4GradientMid).
		Background(shared.Card4GradientStart).
		Width(cardWidth).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				cardHeaderStyle.Copy().Foreground(shared.White).Render("🎫 TICKETS"),
				cardValueStyle.Copy().Foreground(shared.BrightText).Bold(true).Render(fmt.Sprintf("%d", m.summary.TotalTickets)),
				cardSubtextStyle.Copy().Foreground(shared.LightGray).Render("tracked"),
			),
		)
	cards = append(cards, ticketsCard)

	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

// renderAnimatedProgress creates beautiful progress bars with gradients
func (m *AppModel) renderAnimatedProgress() string {
	if m.summary == nil {
		return ""
	}

	var lines []string

	// Section title with icon
	titleWithIcon := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(shared.GradientStart).Render("▸ "),
		sectionTitleStyle.Copy().Render("Progress Overview"),
	)
	lines = append(lines, titleWithIcon)
	lines = append(lines, "")

	// Main progress bar with gradient
	progressRatio := m.summary.LoggedHours / m.summary.RequiredHours
	if progressRatio > 1.0 {
		progressRatio = 1.0
	}

	// Calculate bar width based on terminal width
	availableWidth := m.width - 8
	barWidth := availableWidth - 35 // Account for label and percentage
	if barWidth < 40 {
		barWidth = 40
	}

	// Create a beautiful gradient progress bar
	prog := progress.New(
		progress.WithGradient(string(shared.GradientStart), string(shared.GradientEnd)),
		progress.WithWidth(barWidth),
		progress.WithoutPercentage(),
	)
	prog.SetPercent(progressRatio)

	// Progress bar with styled labels
	progressLabel := lipgloss.NewStyle().
		Foreground(shared.SubduedText).
		Width(20).
		Render("⚡ Overall Progress")

	progressPercent := lipgloss.NewStyle().
		Bold(true).
		Foreground(getProgressColor(progressRatio)).
		Width(10).
		Align(lipgloss.Right).
		Render(fmt.Sprintf("%.0f%%", progressRatio*100))

	progressLine := lipgloss.JoinHorizontal(lipgloss.Top,
		progressLabel,
		prog.View(),
		progressPercent,
	)
	lines = append(lines, progressLine)
	lines = append(lines, "")

	// Days breakdown with mini progress bars
	workdaysProgress := float64(m.summary.LoggedDays) / float64(m.summary.Workdays)
	miniProg := progress.New(
		progress.WithGradient(string(shared.Card2GradientStart), string(shared.Card2GradientEnd)),
		progress.WithWidth(30),
		progress.WithoutPercentage(),
	)
	miniProg.SetPercent(workdaysProgress)

	daysLine := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(shared.SubduedText).Width(20).Render("📅 Days Logged"),
		miniProg.View(),
		lipgloss.NewStyle().Foreground(shared.Card2GradientStart).Width(15).Align(lipgloss.Right).
			Render(fmt.Sprintf("%d/%d days", m.summary.LoggedDays, m.summary.Workdays)),
	)
	lines = append(lines, daysLine)
	lines = append(lines, "")

	// Beautiful status badges with icons
	badges := lipgloss.JoinHorizontal(lipgloss.Top,
		completeBadge.Copy().Render(fmt.Sprintf("✨ Complete: %d", len(m.summary.CompliantDays))),
		partialBadge.Copy().Render(fmt.Sprintf("⚡ Partial: %d", len(m.summary.PartialDays))),
		missingBadge.Copy().Render(fmt.Sprintf("⚠️  Missing: %d", len(m.summary.MissingDays))),
	)
	lines = append(lines, badges)

	// Create a beautiful box with gradient border
	boxWidth := availableWidth
	if boxWidth > 120 {
		boxWidth = 120
	}

	progressBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(shared.PrimaryColor).
		Background(lipgloss.Color("#0A0A0F")).
		Padding(1, 3).
		Width(boxWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))

	return progressBox
}

// LoadStatsData loads data for the stats page
func (m *AppModel) LoadStatsData() tea.Cmd {
	// Set up year-to-date period if not already set
	if m.period.Start.IsZero() {
		now := time.Now()
		yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		m.period = shared.Period{Start: yearStart, End: now}
	}

	// Load worklogs data
	return LoadWorklogsWithProgress(m.service, m.period)
}

// HandleStatsInput handles keyboard input for the stats page
func (m *AppModel) HandleStatsInput(msg tea.KeyMsg) tea.Cmd {
	// Handle stats-specific keys
	switch msg.String() {
	case "q", "Q":
		// Quit the application
		return tea.Quit

	case "t", "T":
		// View ticket details
		m.currentView = TicketDetailsView
		return nil

	case "p", "P":
		// Change period
		m.currentView = PeriodSelectionView
		return nil

	case "l", "L":
		// Log time
		m.progressPopup.Show("Loading Time Logging", "Preparing time logging interface...")
		return m.loadTimeLoggingWithProgress()

	case "c", "C":
		// Create ticket
		m.progressPopup.Show("Loading Ticket Creation", "Preparing ticket creation form...")
		return m.loadTicketCreationWithProgress()

	case "r", "R":
		// Refresh data
		m.progressPopup.ShowWithSteps("🔄 Refreshing Statistics", []string{
			"Fetching latest data from JIRA",
			"Processing worklogs",
			"Updating statistics",
		})
		return LoadWorklogsWithProgress(m.service, m.period)

	case "h", "H":
		// Return to stats page (already here)
		return nil

	case "m", "M":
		// Go to main menu
		m.currentView = MainMenuView
		return nil
	}

	return nil
}

// Helper function to get color based on progress ratio
func getProgressColor(ratio float64) lipgloss.Color {
	if ratio >= 1.0 {
		return shared.Card2GradientStart // Beautiful green
	} else if ratio >= 0.75 {
		return shared.Card3GradientStart // Cyan
	} else if ratio >= 0.5 {
		return shared.WarningColor // Yellow
	} else if ratio >= 0.25 {
		return shared.Card4GradientStart // Orange
	}
	return shared.ErrorColor // Red
}
