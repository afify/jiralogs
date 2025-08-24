package list

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"LogS/internal/app"
	"LogS/jira"
	"LogS/internal/tui/styles"
	"LogS/internal/tui/util"
	"LogS/internal/tui/components/core/layout"
	"github.com/charmbracelet/lipgloss/v2"
)

type (
	WorklogSelectedMsg struct {
		WorklogID string
	}
	WorklogSelectionMsg struct {
		WorklogID string
	}
	RefreshMsg struct {
		Period string
	}
	WorklogsRefreshedMsg struct{}
)

type WorklogListCmp interface {
	util.Model
	layout.Sizeable
	Refresh() tea.Cmd
	SetWorklogs(worklogs []jira.Worklog) tea.Cmd
}

type worklogList struct {
	width, height int
	app           *app.App
	worklogs      []jira.Worklog
	selected      int
	keyMap        KeyMap
}

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Select   key.Binding
	Refresh  key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
	}
}

func New(app *app.App) WorklogListCmp {
	return &worklogList{
		app:    app,
		keyMap: DefaultKeyMap(),
	}
}

func (m *worklogList) Init() tea.Cmd {
	return m.Refresh()
}

func (m *worklogList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, m.keyMap.Down):
			if m.selected < len(m.worklogs)-1 {
				m.selected++
			}
		case key.Matches(msg, m.keyMap.Select):
			if len(m.worklogs) > 0 && m.selected < len(m.worklogs) {
				selectedWorklog := m.worklogs[m.selected]
				return m, func() tea.Msg {
					return WorklogSelectedMsg{WorklogID: selectedWorklog.ID}
				}
			}
		case key.Matches(msg, m.keyMap.Refresh):
			return m, m.Refresh()
		}
	case tea.MouseClickMsg:
		// Handle mouse clicks for selection
		if msg.X >= 0 && msg.X < m.width && msg.Y >= 0 {
			if msg.Y < len(m.worklogs) {
				m.selected = msg.Y
				if len(m.worklogs) > 0 {
					selectedWorklog := m.worklogs[m.selected]
					return m, func() tea.Msg {
						return WorklogSelectedMsg{WorklogID: selectedWorklog.ID}
					}
				}
			}
		}
	case WorklogsRefreshedMsg:
		// Worklogs have been refreshed
		return m, nil
	}
	return m, nil
}

func (m *worklogList) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	t := styles.CurrentTheme()
	s := t.S()

	var content strings.Builder
	
	// Header
	header := s.Title.Render("Worklogs")
	content.WriteString(header)
	content.WriteString("\n")

	if len(m.worklogs) == 0 {
		content.WriteString(s.Muted.Render("No worklogs found"))
		content.WriteString("\n")
		content.WriteString(s.Subtle.Render("Press 'r' to refresh"))
	} else {
		// Worklog items
		for i, worklog := range m.worklogs {
			if i > 0 {
				content.WriteString("\n")
			}

			var itemStyle lipgloss.Style
			prefix := "  "
			
			if i == m.selected {
				itemStyle = s.TextSelected
				prefix = "> "
			} else {
				itemStyle = s.Base
			}

			// Format: "KEY - Summary (2.5h) - 2024-01-15"
			line := fmt.Sprintf("%s%s - %s (%.1fh) - %s",
				prefix,
				worklog.Issue.Key,
				truncateString(worklog.Issue.Fields.Summary, 40),
				worklog.TimeSpentHours,
				worklog.Started.Format("2006-01-02"),
			)

			content.WriteString(itemStyle.Render(line))
		}
	}

	return s.Base.
		Width(m.width).
		Height(m.height).
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderRight(true).
		Render(content.String())
}

func (m *worklogList) Refresh() tea.Cmd {
	return func() tea.Msg {
		// TODO: Integrate with JIRA service to fetch worklogs
		// For now, return a refresh completion message
		return WorklogsRefreshedMsg{}
	}
}

func (m *worklogList) SetWorklogs(worklogs []jira.Worklog) tea.Cmd {
	m.worklogs = worklogs
	if m.selected >= len(worklogs) {
		m.selected = len(worklogs) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
	return nil
}

// SetSize implements layout.Sizeable.
func (m *worklogList) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// GetSize implements layout.Sizeable.
func (m *worklogList) GetSize() (int, int) {
	return m.width, m.height
}

func (m *worklogList) Focus() tea.Cmd {
	return nil
}

func (m *worklogList) Blur() tea.Cmd {
	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}