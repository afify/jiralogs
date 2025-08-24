package header

import (
	"fmt"
	"strings"

	"LogS/internal/tui/styles"
	"LogS/internal/tui/util"
	"LogS/jira"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type Header interface {
	util.Model
	SetWorklog(worklog jira.Worklog) tea.Cmd
	SetWidth(width int) tea.Cmd
	SetDetailsOpen(open bool)
	ShowingDetails() bool
}

type header struct {
	width       int
	worklog     jira.Worklog
	detailsOpen bool
}

func New() Header {
	return &header{
		width: 0,
	}
}

func (h *header) Init() tea.Cmd {
	return nil
}

func (h *header) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case jira.WorklogUpdatedMsg:
		h.worklog = msg.Worklog
	}
	return h, nil
}

func (h *header) View() string {
	const (
		gap          = " "
		diag         = "╱"
		minDiags     = 3
		leftPadding  = 1
		rightPadding = 1
	)

	t := styles.CurrentTheme()

	var b strings.Builder

	b.WriteString(t.S().Base.Foreground(t.Secondary).Render("JIRA™"))
	b.WriteString(gap)
	b.WriteString(styles.ApplyBoldForegroundGrad("LOGS", t.Secondary, t.Primary))
	b.WriteString(gap)

	availDetailWidth := h.width - leftPadding - rightPadding - lipgloss.Width(b.String()) - minDiags
	details := h.details(availDetailWidth)

	remainingWidth := h.width -
		lipgloss.Width(b.String()) -
		lipgloss.Width(details) -
		leftPadding -
		rightPadding

	if remainingWidth > 0 {
		b.WriteString(t.S().Base.Foreground(t.Primary).Render(
			strings.Repeat(diag, max(minDiags, remainingWidth)),
		))
		b.WriteString(gap)
	}

	b.WriteString(details)

	return t.S().Base.Padding(0, rightPadding, 0, leftPadding).Render(b.String())
}

func (h *header) details(availWidth int) string {
	s := styles.CurrentTheme().S()

	var parts []string

	// Show worklog count or selected worklog info
	if h.worklog.ID != "" {
		parts = append(parts, s.Muted.Render(fmt.Sprintf("Issue: %s", h.worklog.Issue.Key)))
		parts = append(parts, s.Muted.Render(fmt.Sprintf("%.1fh", h.worklog.TimeSpentHours)))
	} else {
		parts = append(parts, s.Muted.Render("No worklog selected"))
	}

	const keystroke = "ctrl+d"
	if h.detailsOpen {
		parts = append(parts, s.Muted.Render(keystroke)+s.Subtle.Render(" close"))
	} else {
		parts = append(parts, s.Muted.Render(keystroke)+s.Subtle.Render(" open "))
	}

	dot := s.Subtle.Render(" • ")
	metadata := strings.Join(parts, dot)
	metadata = dot + metadata

	return metadata
}

func (h *header) SetDetailsOpen(open bool) {
	h.detailsOpen = open
}

// SetWorklog implements Header.
func (h *header) SetWorklog(worklog jira.Worklog) tea.Cmd {
	h.worklog = worklog
	return nil
}

// SetWidth implements Header.
func (h *header) SetWidth(width int) tea.Cmd {
	h.width = width
	return nil
}

// ShowingDetails implements Header.
func (h *header) ShowingDetails() bool {
	return h.detailsOpen
}
