package details

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"LogS/internal/app"
	"LogS/jira"
	"LogS/internal/tui/styles"
	"LogS/internal/tui/util"
	"LogS/internal/tui/components/core/layout"
	"github.com/charmbracelet/lipgloss/v2"
)

type (
	OpenDetailsMsg struct {
		WorklogID string
	}
	WorklogDetailsMsg struct {
		Worklog jira.Worklog
	}
)

type Details interface {
	util.Model
	layout.Sizeable
	SetWorklog(worklogID string) tea.Cmd
}

type details struct {
	width, height int
	app           *app.App
	worklog       jira.Worklog
}

func New(app *app.App) Details {
	return &details{
		app: app,
	}
}

func (m *details) Init() tea.Cmd {
	return nil
}

func (m *details) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WorklogDetailsMsg:
		m.worklog = msg.Worklog
	case OpenDetailsMsg:
		return m, m.SetWorklog(msg.WorklogID)
	}
	return m, nil
}

func (m *details) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	t := styles.CurrentTheme()
	s := t.S()

	var content strings.Builder

	// Header
	header := s.Title.Render("Worklog Details")
	content.WriteString(header)
	content.WriteString("\n\n")

	if m.worklog.ID == "" {
		content.WriteString(s.Muted.Render("Select a worklog to view details"))
	} else {
		// Issue information
		content.WriteString(s.Subtitle.Render("Issue Information"))
		content.WriteString("\n")
		content.WriteString(s.Base.Render(fmt.Sprintf("Key: %s", m.worklog.Issue.Key)))
		content.WriteString("\n")
		content.WriteString(s.Base.Render(fmt.Sprintf("Summary: %s", m.worklog.Issue.Fields.Summary)))
		content.WriteString("\n")
		content.WriteString(s.Base.Render(fmt.Sprintf("Status: %s", m.worklog.Issue.Fields.Status.Name)))
		content.WriteString("\n")
		content.WriteString(s.Base.Render(fmt.Sprintf("Assignee: %s", getAssigneeName(m.worklog.Issue))))
		content.WriteString("\n\n")

		// Worklog information
		content.WriteString(s.Subtitle.Render("Worklog Information"))
		content.WriteString("\n")
		content.WriteString(s.Base.Render(fmt.Sprintf("Time Spent: %.1f hours", m.worklog.TimeSpentHours)))
		content.WriteString("\n")
		if m.worklog.Started != nil {
			content.WriteString(s.Base.Render(fmt.Sprintf("Date: %s", m.worklog.Started.Format("2006-01-02 15:04"))))
			content.WriteString("\n")
		}
		if m.worklog.Author.DisplayName != "" {
			content.WriteString(s.Base.Render(fmt.Sprintf("Author: %s", m.worklog.Author.DisplayName)))
			content.WriteString("\n")
		}
		content.WriteString("\n")

		// Comment section
		if m.worklog.Comment != "" {
			content.WriteString(s.Subtitle.Render("Comment"))
			content.WriteString("\n")
			content.WriteString(s.Base.Render(m.worklog.Comment))
			content.WriteString("\n\n")
		}

		// Actions section
		content.WriteString(s.Subtitle.Render("Actions"))
		content.WriteString("\n")
		content.WriteString(s.Muted.Render("• Press 'l' to log more time"))
		content.WriteString("\n")
		content.WriteString(s.Muted.Render("• Press 'e' to export worklog"))
		content.WriteString("\n")
		content.WriteString(s.Muted.Render("• Press 'r' to refresh"))
	}

	return s.Base.
		Width(m.width).
		Height(m.height).
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		Render(content.String())
}

func (m *details) SetWorklog(worklogID string) tea.Cmd {
	return func() tea.Msg {
		// TODO: Fetch worklog details from JIRA service
		// For now, return empty worklog
		return WorklogDetailsMsg{Worklog: jira.Worklog{ID: worklogID}}
	}
}

// SetSize implements layout.Sizeable.
func (m *details) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// GetSize implements layout.Sizeable.
func (m *details) GetSize() (int, int) {
	return m.width, m.height
}

func (m *details) Focus() tea.Cmd {
	return nil
}

func (m *details) Blur() tea.Cmd {
	return nil
}

func getAssigneeName(issue jira.Issue) string {
	if issue.Fields.Assignee != nil && issue.Fields.Assignee.DisplayName != "" {
		return issue.Fields.Assignee.DisplayName
	}
	return "Unassigned"
}