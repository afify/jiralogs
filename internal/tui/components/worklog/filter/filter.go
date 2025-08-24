package filter

import (
	"strings"

	"LogS/internal/app"
	"LogS/internal/tui/components/core/layout"
	"LogS/internal/tui/styles"
	"LogS/internal/tui/util"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type (
	FilterChangedMsg struct {
		Filter string
	}
	PeriodChangedMsg struct {
		Period string
	}
)

type Filter interface {
	util.Model
	layout.Sizeable
	SetFilter(filter string) tea.Cmd
	GetFilter() string
}

type filter struct {
	width, height  int
	app            *app.App
	filterText     string
	period         string
	periods        []string
	selectedPeriod int
}

func New(app *app.App) Filter {
	return &filter{
		app: app,
		periods: []string{
			"Week to date",
			"Month to date",
			"Last 30 days",
			"Last 3 months",
			"Year to date",
		},
		selectedPeriod: 1, // Default to "Month to date"
	}
}

func (m *filter) Init() tea.Cmd {
	return nil
}

func (m *filter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Cycle through periods
			m.selectedPeriod = (m.selectedPeriod + 1) % len(m.periods)
			return m, func() tea.Msg {
				return PeriodChangedMsg{Period: m.periods[m.selectedPeriod]}
			}
		}
	}
	return m, nil
}

func (m *filter) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	t := styles.CurrentTheme()
	s := t.S()

	var content strings.Builder

	// Title
	title := s.Title.Render("Filter & Period")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Current filter
	if m.filterText != "" {
		content.WriteString(s.Subtitle.Render("Filter:"))
		content.WriteString("\n")
		content.WriteString(s.Base.Render(m.filterText))
		content.WriteString("\n\n")
	}

	// Period selection
	content.WriteString(s.Subtitle.Render("Time Period:"))
	content.WriteString("\n")

	for i, period := range m.periods {
		prefix := "  "
		style := s.Muted

		if i == m.selectedPeriod {
			prefix = "> "
			style = s.TextSelected
		}

		line := prefix + period
		content.WriteString(style.Render(line))
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\n")
	content.WriteString(s.Subtle.Render("Tab to change period"))

	return s.Base.
		Width(m.width).
		Height(m.height).
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		Render(content.String())
}

func (m *filter) SetFilter(filter string) tea.Cmd {
	m.filterText = filter
	return func() tea.Msg {
		return FilterChangedMsg{Filter: filter}
	}
}

func (m *filter) GetFilter() string {
	return m.filterText
}

// SetSize implements layout.Sizeable.
func (m *filter) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}

// GetSize implements layout.Sizeable.
func (m *filter) GetSize() (int, int) {
	return m.width, m.height
}

func (m *filter) Focus() tea.Cmd {
	return nil
}

func (m *filter) Blur() tea.Cmd {
	return nil
}
