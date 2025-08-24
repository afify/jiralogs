package calendar

import (
	"fmt"
	"image/color"
	"strconv"
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

var CalendarPageID page.PageID = "calendar"

type CalendarPage interface {
	util.Model
	layout.Help
}

type DayStatus int

const (
	DayFuture  DayStatus = iota // Upcoming days (dimmed)
	DayNoLog                    // Days not logged (different background)
	DayLogged                   // Days with logged hours
	DayToday                    // Today
	DayWeekend                  // Weekend days
)

type calendarPage struct {
	width, height int
	app           *app.App
	keyMap        KeyMap

	// Calendar state
	currentYear   int
	selectedMonth int
	selectedDay   int
	today         time.Time

	// Data
	dailyHours map[string]float64 // Date string -> hours logged
}

func New(app *app.App) CalendarPage {
	shared.LogErrorf("CALENDAR_NEW", "Creating new calendar page")

	now := time.Now()
	return &calendarPage{
		app:           app,
		keyMap:        DefaultKeyMap(),
		currentYear:   now.Year(),
		selectedMonth: int(now.Month()),
		selectedDay:   now.Day(),
		today:         now,
		dailyHours:    make(map[string]float64),
	}
}

func (p *calendarPage) Init() tea.Cmd {
	shared.LogErrorf("CALENDAR_INIT", "Initializing calendar page")
	return nil
}

func (p *calendarPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	shared.LogErrorf("CALENDAR_UPDATE", "Received message type: %T", msg)

	switch msg := msg.(type) {
	case welcome.WorklogDataMsg:
		shared.LogErrorf("CALENDAR_UPDATE", "Received WorklogDataMsg")
		p.SetWorklogData(msg.DailyHours)
		return p, nil

	case tea.WindowSizeMsg:
		shared.LogErrorf("CALENDAR_UPDATE", "Window resize: %dx%d", msg.Width, msg.Height)
		return p, p.SetSize(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		shared.LogErrorf("CALENDAR_UPDATE", "Key press: %s", msg.String())
		switch {
		case key.Matches(msg, p.keyMap.Up):
			p.selectedMonth--
			if p.selectedMonth < 1 {
				p.selectedMonth = 12
				p.currentYear--
			}
			return p, nil
		case key.Matches(msg, p.keyMap.Down):
			p.selectedMonth++
			if p.selectedMonth > 12 {
				p.selectedMonth = 1
				p.currentYear++
			}
			return p, nil
		case key.Matches(msg, p.keyMap.Left):
			p.selectedMonth--
			if p.selectedMonth < 1 {
				p.selectedMonth = 12
				p.currentYear--
			}
			return p, nil
		case key.Matches(msg, p.keyMap.Right):
			p.selectedMonth++
			if p.selectedMonth > 12 {
				p.selectedMonth = 1
				p.currentYear++
			}
			return p, nil
		case key.Matches(msg, p.keyMap.PrevYear):
			p.currentYear--
			return p, nil
		case key.Matches(msg, p.keyMap.NextYear):
			p.currentYear++
			return p, nil
		case key.Matches(msg, p.keyMap.Today):
			p.currentYear = p.today.Year()
			p.selectedMonth = int(p.today.Month())
			p.selectedDay = p.today.Day()
			return p, nil
		case key.Matches(msg, p.keyMap.Back):
			shared.LogErrorf("CALENDAR_UPDATE", "Back key pressed")
			// Navigate back to stats page
			return p, func() tea.Msg {
				return page.PageChangeMsg{ID: "stats"}
			}
		case key.Matches(msg, p.keyMap.Quit):
			shared.LogErrorf("CALENDAR_UPDATE", "Quit key pressed")
			return p, tea.Quit
		}
	}

	return p, nil
}

func (p *calendarPage) View() string {
	shared.LogErrorf("CALENDAR_VIEW", "Rendering calendar view - dimensions: %dx%d", p.width, p.height)

	if p.width == 0 || p.height == 0 {
		shared.LogErrorf("CALENDAR_VIEW", "Dimensions not set, returning empty view")
		return ""
	}

	t := styles.CurrentTheme()
	shared.LogErrorf("CALENDAR_VIEW", "Retrieved current theme")

	// Create the full-width top logo with Crush's exact colors
	logoOpts := logo.Opts{
		FieldColor:   t.Primary,
		TitleColorA:  t.Secondary,
		TitleColorB:  t.Primary,
		CharmColor:   t.Secondary,
		VersionColor: t.Primary,
		Width:        p.width - 6,
	}

	logoStr := logo.Render(false, logoOpts) // false for full logo
	shared.LogErrorf("CALENDAR_VIEW", "Full-width logo created")

	// Create calendar year view
	calendarContent := p.renderYearView(t)
	shared.LogErrorf("CALENDAR_VIEW", "Calendar content rendered")

	// Join logo and calendar vertically with proper spacing
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		logoStr,
		"", // Empty line for spacing
		calendarContent,
	)

	// Apply base styling
	baseStyle := t.S().Base.
		Width(p.width).
		Height(p.height).
		Padding(1, 2)

	return baseStyle.Render(content)
}

func (p *calendarPage) renderYearView(t *styles.Theme) string {
	// Create year header
	yearHeader := p.renderYearHeader(t)

	// Create months grid (4 rows x 3 columns)
	monthsGrid := p.renderMonthsGrid(t)

	// Legend for day status colors
	legend := p.renderLegend(t)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		yearHeader,
		"",
		monthsGrid,
		"",
		legend,
	)
}

func (p *calendarPage) renderYearHeader(t *styles.Theme) string {
	yearStr := strconv.Itoa(p.currentYear)

	// Create styled year with navigation arrows
	leftArrow := t.S().Base.Foreground(t.FgMuted).Render("◀")
	rightArrow := t.S().Base.Foreground(t.FgMuted).Render("▶")
	year := t.S().Base.Foreground(t.Primary).Bold(true).Render(yearStr)

	header := fmt.Sprintf("  %s    %s    %s", leftArrow, year, rightArrow)

	// Center the header
	headerStyle := t.S().Base.
		Align(lipgloss.Center).
		Width(p.width - 8)

	return headerStyle.Render(header)
}

func (p *calendarPage) renderMonthsGrid(t *styles.Theme) string {
	monthWidth := (p.width - 12) / 3 // 3 columns with margins
	if monthWidth < 20 {
		monthWidth = 20
	}

	var rows []string

	// Render 4 rows of 3 months each
	for row := 0; row < 4; row++ {
		var monthsInRow []string

		for col := 0; col < 3; col++ {
			monthNum := row*3 + col + 1
			if monthNum <= 12 {
				month := p.renderMonth(t, monthNum, monthWidth)
				monthsInRow = append(monthsInRow, month)
			}
		}

		if len(monthsInRow) > 0 {
			rowStr := lipgloss.JoinHorizontal(lipgloss.Top, monthsInRow...)
			rows = append(rows, rowStr)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (p *calendarPage) renderMonth(t *styles.Theme, month int, width int) string {
	monthName := time.Month(month).String()

	// Month header
	headerStyle := t.S().Base.
		Foreground(t.Primary).
		Bold(true).
		Align(lipgloss.Center).
		Width(width)

	header := headerStyle.Render(monthName)

	// Days of week header
	daysHeader := p.renderDaysOfWeekHeader(t, width)

	// Calendar days
	monthDays := p.renderMonthDays(t, p.currentYear, month, width)

	// Month container with border
	monthStyle := t.S().Base.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Width(width).
		Padding(0, 1).
		Margin(0, 1)

	monthContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		daysHeader,
		monthDays,
	)

	return monthStyle.Render(monthContent)
}

func (p *calendarPage) renderDaysOfWeekHeader(t *styles.Theme, width int) string {
	days := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}

	dayWidth := (width - 8) / 7 // 7 days with some padding
	if dayWidth < 2 {
		dayWidth = 2
	}

	var styledDays []string
	for _, day := range days {
		dayStyle := t.S().Base.
			Foreground(t.FgMuted).
			Width(dayWidth).
			Align(lipgloss.Center)
		styledDays = append(styledDays, dayStyle.Render(day))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, styledDays...)
}

func (p *calendarPage) renderMonthDays(t *styles.Theme, year, month int, width int) string {
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)

	dayWidth := (width - 8) / 7
	if dayWidth < 2 {
		dayWidth = 2
	}

	var weeks []string
	var currentWeek []string

	// Add empty cells for days before the first day of the month
	for i := 0; i < int(firstDay.Weekday()); i++ {
		emptyStyle := t.S().Base.Width(dayWidth).Align(lipgloss.Center)
		currentWeek = append(currentWeek, emptyStyle.Render(""))
	}

	// Add days of the month
	for day := 1; day <= lastDay.Day(); day++ {
		dayDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		dayStatus := p.getDayStatus(dayDate)

		dayStr := strconv.Itoa(day)
		dayStyle := p.getDayStyle(t, dayStatus, dayWidth)

		styledDay := dayStyle.Render(dayStr)
		currentWeek = append(currentWeek, styledDay)

		// If we have 7 days or it's the last day, complete the week
		if len(currentWeek) == 7 {
			weekStr := lipgloss.JoinHorizontal(lipgloss.Top, currentWeek...)
			weeks = append(weeks, weekStr)
			currentWeek = []string{}
		}
	}

	// Fill remaining days of the last week
	for len(currentWeek) < 7 && len(currentWeek) > 0 {
		emptyStyle := t.S().Base.Width(dayWidth).Align(lipgloss.Center)
		currentWeek = append(currentWeek, emptyStyle.Render(""))
	}

	if len(currentWeek) > 0 {
		weekStr := lipgloss.JoinHorizontal(lipgloss.Top, currentWeek...)
		weeks = append(weeks, weekStr)
	}

	return lipgloss.JoinVertical(lipgloss.Left, weeks...)
}

func (p *calendarPage) getDayStatus(date time.Time) DayStatus {
	dateStr := date.Format("2006-01-02")

	// Check if it's today
	if p.today.Format("2006-01-02") == dateStr {
		return DayToday
	}

	// Check if it's in the future
	if date.After(p.today) {
		return DayFuture
	}

	// Check if it's a weekend (Friday/Saturday for Middle East)
	if date.Weekday() == time.Friday || date.Weekday() == time.Saturday {
		return DayWeekend
	}

	// Check if there are logged hours
	if hours, exists := p.dailyHours[dateStr]; exists && hours > 0 {
		return DayLogged
	}

	// No logged hours on a workday
	return DayNoLog
}

func (p *calendarPage) getDayStyle(t *styles.Theme, status DayStatus, width int) lipgloss.Style {
	baseStyle := t.S().Base.Width(width).Align(lipgloss.Center)

	switch status {
	case DayToday:
		return baseStyle.Background(t.Primary).Foreground(t.White).Bold(true)
	case DayLogged:
		return baseStyle.Background(t.Success).Foreground(t.White)
	case DayNoLog:
		return baseStyle.Background(t.Error).Foreground(t.White)
	case DayWeekend:
		return baseStyle.Background(t.BgSubtle).Foreground(t.FgMuted)
	case DayFuture:
		return baseStyle.Foreground(t.FgMuted)
	default:
		return baseStyle.Foreground(t.FgBase)
	}
}

func (p *calendarPage) renderLegend(t *styles.Theme) string {
	legendItems := []struct {
		label string
		color color.Color
		bg    color.Color
	}{
		{"Today", t.White, t.Primary},
		{"Logged", t.White, t.Success},
		{"Not Logged", t.White, t.Error},
		{"Weekend", t.FgMuted, t.BgSubtle},
		{"Future", t.FgMuted, nil},
	}

	var items []string
	for _, item := range legendItems {
		style := t.S().Base.Foreground(item.color).Padding(0, 1)
		if item.bg != nil {
			style = style.Background(item.bg)
		}

		legendItem := style.Render("██") + " " + t.S().Base.Foreground(t.FgBase).Render(item.label)
		items = append(items, legendItem)
	}

	legend := lipgloss.JoinHorizontal(lipgloss.Top, items...)

	legendStyle := t.S().Base.
		Align(lipgloss.Center).
		Width(p.width - 8)

	return legendStyle.Render(legend)
}

func (p *calendarPage) SetSize(width, height int) tea.Cmd {
	p.width = width
	p.height = height
	shared.LogErrorf("CALENDAR_SIZE", "Calendar page size set to %dx%d", width, height)
	return nil
}

func (p *calendarPage) SetWorklogData(dailyHours map[string]float64) {
	shared.LogErrorf("CALENDAR_DATA", "Setting worklog data - DailyHours: %d", len(dailyHours))
	p.dailyHours = dailyHours
}

func (p *calendarPage) Bindings() []key.Binding {
	return []key.Binding{
		p.keyMap.Up,
		p.keyMap.Down,
		p.keyMap.Left,
		p.keyMap.Right,
		p.keyMap.PrevYear,
		p.keyMap.NextYear,
		p.keyMap.Today,
		p.keyMap.Back,
		p.keyMap.Quit,
	}
}

func (p *calendarPage) Help() help.KeyMap {
	var shortList []key.Binding
	var fullList [][]key.Binding

	shortList = append(shortList,
		key.NewBinding(
			key.WithKeys("↑/↓/←/→"),
			key.WithHelp("arrows", "navigate"),
		),
		key.NewBinding(
			key.WithKeys("[/]"),
			key.WithHelp("[/]", "prev/next year"),
		),
		key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "today"),
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
