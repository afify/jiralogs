package tui

import (
	"fmt"
	"time"

	"LogS/shared"
	"github.com/charmbracelet/lipgloss/v2"
)

func (m *AppModel) renderInitializationView() string {
	return m.RenderInitializationPage()
}

func (m *AppModel) renderMainMenuView() string {
	title := renderPageTitle("MAIN MENU", m.width)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		m.mainMenu.View(),
	)
	return ContentStyle.Render(content)
}

func (m *AppModel) renderPeriodSelectionView() string {
	title := renderPageTitle("SELECT PERIOD", m.width)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		m.periodList.View(),
	)
	return ContentStyle.Render(content)
}

// renderPageTitle renders a centered title below the top bar
func renderPageTitle(title string, width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(shared.PrimaryColor).
		Bold(true).
		Background(shared.BgLight).
		Padding(0, 3).
		MarginBottom(1)

	return lipgloss.PlaceHorizontal(width, lipgloss.Center, titleStyle.Render(title))
}

func (m *AppModel) renderStatsView() string {
	// Use the modern stats page from page_stats.go
	return m.RenderStatsPage()
}

func (m *AppModel) renderTimeLoggingView() string {
	title := renderPageTitle("LOG TIME", m.width)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"Time logging form will be shown here",
	)
	return ContentStyle.Render(content)
}

func (m *AppModel) renderTicketCreationView() string {
	title := renderPageTitle("CREATE TICKET", m.width)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"Ticket creation form will be shown here",
	)
	return ContentStyle.Render(content)
}

func (m *AppModel) renderTicketDetailsView() string {
	// Use the ticket details page from page_ticket_details.go
	return m.RenderTicketDetailsPage()
}

func (m *AppModel) renderHelpLine() string {
	// Create styled key hints
	keyStyle := lipgloss.NewStyle().
		Foreground(shared.Orange).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(shared.LightGray)

	separator := lipgloss.NewStyle().
		Foreground(shared.DarkGray).
		Render(" • ")

	var helpItems []string

	switch m.currentView {
	case MainMenuView:
		helpItems = []string{
			keyStyle.Render("↑/↓") + descStyle.Render(" Navigate"),
			keyStyle.Render("Enter") + descStyle.Render(" Select"),
			keyStyle.Render("Q") + descStyle.Render(" Quit"),
		}
	case PeriodSelectionView:
		helpItems = []string{
			keyStyle.Render("↑/↓") + descStyle.Render(" Navigate"),
			keyStyle.Render("Enter") + descStyle.Render(" Select"),
			keyStyle.Render("ESC") + descStyle.Render(" Back"),
			keyStyle.Render("Q") + descStyle.Render(" Quit"),
		}
	case StatsView:
		helpItems = []string{
			keyStyle.Render("T") + descStyle.Render(" Tickets"),
			keyStyle.Render("L") + descStyle.Render(" Log"),
			keyStyle.Render("C") + descStyle.Render(" Create"),
			keyStyle.Render("P") + descStyle.Render(" Period"),
			keyStyle.Render("R") + descStyle.Render(" Refresh"),
			keyStyle.Render("M") + descStyle.Render(" Menu"),
			keyStyle.Render("Q") + descStyle.Render(" Quit"),
		}
	case TicketDetailsView:
		helpItems = []string{
			keyStyle.Render("↑/↓") + descStyle.Render(" Scroll"),
			keyStyle.Render("H") + descStyle.Render(" Home"),
			keyStyle.Render("P") + descStyle.Render(" Period"),
			keyStyle.Render("R") + descStyle.Render(" Refresh"),
			keyStyle.Render("ESC") + descStyle.Render(" Back"),
			keyStyle.Render("Q") + descStyle.Render(" Quit"),
		}
	case TimeLoggingView, TicketCreationView:
		helpItems = []string{
			keyStyle.Render("Tab") + descStyle.Render(" Next Field"),
			keyStyle.Render("Shift+Tab") + descStyle.Render(" Previous"),
			keyStyle.Render("Enter") + descStyle.Render(" Submit"),
			keyStyle.Render("ESC") + descStyle.Render(" Cancel"),
			keyStyle.Render("Q") + descStyle.Render(" Quit"),
		}
	default:
		helpItems = []string{
			keyStyle.Render("ESC") + descStyle.Render(" Back"),
			keyStyle.Render("?") + descStyle.Render(" Help"),
			keyStyle.Render("Q") + descStyle.Render(" Quit"),
		}
	}

	help := lipgloss.JoinHorizontal(lipgloss.Left, append([]string{helpItems[0]},
		func() []string {
			var result []string
			for _, item := range helpItems[1:] {
				result = append(result, separator, item)
			}
			return result
		}()...)...)

	// Account for frame borders and padding (2 for borders + 2 for padding = 4)
	availableWidth := m.width - 4
	return HelpStyle.Width(availableWidth).Align(lipgloss.Center).Render(help)
}

func (m *AppModel) renderHeaderInfo() string {
	now := time.Now()

	zone, offset := now.Zone()
	offsetHours := offset / 3600
	offsetMins := (offset % 3600) / 60

	dateTime := now.Format("Jan 02, 2006 • 15:04")
	timezone := fmt.Sprintf("%s UTC%+d:%02d", zone, offsetHours, offsetMins)

	var userEmail, userID string
	if m.currentUser != nil {
		userEmail = m.currentUser.EmailAddress
		userID = m.currentUser.AccountID

		if userEmail == "" && m.client != nil {
			userEmail = m.getEmailFromConfig()
		}
	}

	if userEmail == "" {
		userEmail = "Loading..."
	}
	if userID == "" {
		userID = "Loading..."
	}

	// Create styled user info card
	userCardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(shared.Purple).
		Padding(0, 2).
		MarginLeft(2).
		Background(shared.BgAccent)

	emailStyle := lipgloss.NewStyle().
		Foreground(shared.LightBlue).
		Bold(true)

	idStyle := lipgloss.NewStyle().
		Foreground(shared.Gray)

	dateStyle := lipgloss.NewStyle().
		Foreground(shared.LightGray)

	periodStyle := lipgloss.NewStyle().
		Foreground(shared.Pink).
		Bold(true)

	var headerLines []string
	headerLines = append(headerLines, emailStyle.Render("👤 "+userEmail))
	headerLines = append(headerLines, idStyle.Render("   ID: "+userID[:8]+"..."))
	headerLines = append(headerLines, "")
	headerLines = append(headerLines, dateStyle.Render("📅 "+dateTime))
	headerLines = append(headerLines, dateStyle.Render("🌍 "+timezone))

	if m.period.Start.Year() > 1 {
		periodStart := m.period.Start.Format("Jan 02")
		periodEnd := m.period.End.Format("Jan 02")
		headerLines = append(headerLines, "")
		headerLines = append(headerLines, periodStyle.Render("📊 "+periodStart+" → "+periodEnd))
	}

	headerContent := lipgloss.JoinVertical(lipgloss.Left, headerLines...)

	return userCardStyle.Render(headerContent)
}
