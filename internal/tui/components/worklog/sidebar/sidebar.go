package sidebar

import (
	"fmt"
	"strings"

	"LogS/internal/app"
	"LogS/internal/tui/components/core/layout"
	"LogS/internal/tui/components/logo"
	"LogS/internal/tui/styles"
	"LogS/internal/tui/util"
	"LogS/internal/version"
	"LogS/jira"
	tea "github.com/charmbracelet/bubbletea/v2"
)

const LogoHeightBreakpoint = 30

// Default maximum number of items to show in each section
const (
	DefaultMaxWorklogsShown = 10
	DefaultMaxIssuesShown   = 8
	DefaultMaxProjectsShown = 8
	MinItemsPerSection      = 2 // Minimum items to show per section
)

type WorklogSummary struct {
	Worklog   jira.Worklog
	Issue     jira.Issue
	TimeSpent float64
	Date      string
}

type WorklogSummaryMsg struct {
	Worklogs []WorklogSummary
}

type Sidebar interface {
	util.Model
	layout.Sizeable
	SetWorklog(worklog jira.Worklog) tea.Cmd
	SetCompactMode(bool)
}

type sidebarCmp struct {
	width, height int
	worklog       jira.Worklog
	logo          string
	cwd           string
	app           *app.App
	compactMode   bool
	worklogs      []WorklogSummary
}

func New(app *app.App, compact bool) Sidebar {
	return &sidebarCmp{
		app:         app,
		compactMode: compact,
		worklogs:    []WorklogSummary{},
	}
}

func (m *sidebarCmp) Init() tea.Cmd {
	return nil
}

func (m *sidebarCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WorklogSummaryMsg:
		m.worklogs = msg.Worklogs
		return m, nil
	case jira.WorklogUpdatedMsg:
		m.worklog = msg.Worklog
		return m, nil
	}
	return m, nil
}

func (m *sidebarCmp) View() string {
	t := styles.CurrentTheme()

	var sections []string

	// Logo section
	if !m.compactMode && m.height >= LogoHeightBreakpoint {
		logoOpts := logo.Opts{
			Width: 40,
		}
		m.logo = logo.Render(version.Version, true, logoOpts)
		sections = append(sections, m.logo)
	}

	// Current worklog section
	if m.worklog.ID != "" {
		currentSection := m.renderCurrentWorklog()
		sections = append(sections, currentSection)
	}

	// Recent worklogs section
	if len(m.worklogs) > 0 {
		worklogSection := m.renderRecentWorklogs()
		sections = append(sections, worklogSection)
	}

	// Project info section
	projectSection := m.renderProjectInfo()
	sections = append(sections, projectSection)

	content := strings.Join(sections, "\n\n")

	return t.S().Base.
		Width(m.width).
		Height(m.height).
		Padding(1, 1).
		Render(content)
}

func (m *sidebarCmp) renderCurrentWorklog() string {
	t := styles.CurrentTheme()
	s := t.S()

	title := s.Title.Render("Current Worklog")

	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n")

	if m.worklog.Issue.Key != "" {
		content.WriteString(s.Base.Render(fmt.Sprintf("Issue: %s", m.worklog.Issue.Key)))
		content.WriteString("\n")
	}

	if m.worklog.TimeSpentHours > 0 {
		content.WriteString(s.Muted.Render(fmt.Sprintf("Time: %.1fh", m.worklog.TimeSpentHours)))
		content.WriteString("\n")
	}

	if m.worklog.Started != nil {
		content.WriteString(s.Subtle.Render(fmt.Sprintf("Date: %s", m.worklog.Started.Format("2006-01-02"))))
	}

	return content.String()
}

func (m *sidebarCmp) renderRecentWorklogs() string {
	t := styles.CurrentTheme()
	s := t.S()

	title := s.Title.Render("Recent Worklogs")

	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n")

	maxToShow := min(len(m.worklogs), DefaultMaxWorklogsShown)
	for i, worklog := range m.worklogs[:maxToShow] {
		if i > 0 {
			content.WriteString("\n")
		}

		issueStyle := s.Base
		if worklog.Worklog.ID == m.worklog.ID {
			issueStyle = s.TextSelected
		}

		content.WriteString(issueStyle.Render(fmt.Sprintf("• %s", worklog.Issue.Key)))
		content.WriteString(" ")
		content.WriteString(s.Muted.Render(fmt.Sprintf("%.1fh", worklog.TimeSpent)))
	}

	return content.String()
}

func (m *sidebarCmp) renderProjectInfo() string {
	t := styles.CurrentTheme()
	s := t.S()

	title := s.Title.Render("JIRA Info")

	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n")

	// Show JIRA connection status
	content.WriteString(s.Muted.Render("Status: Connected"))
	content.WriteString("\n")

	// Show total worklogs count
	if len(m.worklogs) > 0 {
		content.WriteString(s.Subtle.Render(fmt.Sprintf("Total: %d worklogs", len(m.worklogs))))
	}

	return content.String()
}

// SetWorklog implements Sidebar.
func (m *sidebarCmp) SetWorklog(worklog jira.Worklog) tea.Cmd {
	m.worklog = worklog
	return nil
}

// SetCompactMode implements Sidebar.
func (m *sidebarCmp) SetCompactMode(compact bool) {
	m.compactMode = compact
}

// SetSize implements layout.Sizeable.
func (m *sidebarCmp) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// GetSize implements layout.Sizeable.
func (m *sidebarCmp) GetSize() (int, int) {
	return m.width, m.height
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
