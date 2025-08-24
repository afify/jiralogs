package tui

import (
	"LogS/shared"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NotificationType represents the type of notification
type NotificationType int

const (
	NotificationSuccess NotificationType = iota
	NotificationError
	NotificationWarning
	NotificationInfo
)

// NotificationPopup represents a modal notification overlay
type NotificationPopup struct {
	visible     bool
	notifType   NotificationType
	title       string
	message     string
	buttons     []string
	selectedBtn int
	width       int
	height      int
}

// Notification styles using True Color RGB from shared
var (
	notifBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Background(shared.DialogBg).
			Padding(2, 3).
			Width(60)

	notifTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(shared.BrightText).
			Background(shared.BgLight).
			Padding(0, 2).
			MarginBottom(1).
			Align(lipgloss.Center)

	notifMessageStyle = lipgloss.NewStyle().
				Foreground(shared.SubduedText).
				Align(lipgloss.Center).
				MarginBottom(1)

	notifButtonStyle = lipgloss.NewStyle().
				Foreground(shared.SubduedText).
				Background(shared.BgLight).
				Padding(0, 3).
				MarginTop(1).
				MarginRight(1)

	// Active button with gradient effect
	notifActiveButtonStyle = lipgloss.NewStyle().
				Foreground(shared.BrightText).
				Background(shared.ButtonGradientStart).
				Padding(0, 3).
				MarginTop(1).
				MarginRight(1).
				Bold(true)
)

// NewNotificationPopup creates a new notification popup
func NewNotificationPopup() NotificationPopup {
	return NotificationPopup{
		visible:     false,
		buttons:     []string{"OK"},
		selectedBtn: 0,
	}
}

// ShowSuccess displays a success notification
func (n *NotificationPopup) ShowSuccess(title, message string) {
	n.visible = true
	n.notifType = NotificationSuccess
	n.title = "✅ " + title
	n.message = message
	n.buttons = []string{"OK"}
	n.selectedBtn = 0
}

// ShowError displays an error notification with retry option
func (n *NotificationPopup) ShowError(title, message string) {
	n.visible = true
	n.notifType = NotificationError
	n.title = "❌ " + title
	n.message = message
	n.buttons = []string{"Retry", "Cancel"}
	n.selectedBtn = 0
}

// ShowWarning displays a warning notification
func (n *NotificationPopup) ShowWarning(title, message string) {
	n.visible = true
	n.notifType = NotificationWarning
	n.title = "⚠️ " + title
	n.message = message
	n.buttons = []string{"OK"}
	n.selectedBtn = 0
}

// ShowInfo displays an info notification
func (n *NotificationPopup) ShowInfo(title, message string) {
	n.visible = true
	n.notifType = NotificationInfo
	n.title = "ℹ️ " + title
	n.message = message
	n.buttons = []string{"OK"}
	n.selectedBtn = 0
}

// ShowWithButtons displays notification with custom buttons
func (n *NotificationPopup) ShowWithButtons(notifType NotificationType, title, message string, buttons []string) {
	n.visible = true
	n.notifType = notifType

	// Add icon based on type
	switch notifType {
	case NotificationSuccess:
		n.title = "✅ " + title
	case NotificationError:
		n.title = "❌ " + title
	case NotificationWarning:
		n.title = "⚠️ " + title
	case NotificationInfo:
		n.title = "ℹ️ " + title
	default:
		n.title = title
	}

	n.message = message
	n.buttons = buttons
	n.selectedBtn = 0
}

// Hide hides the notification (only called after button action)
func (n *NotificationPopup) Hide() {
	n.visible = false
	n.selectedBtn = 0
}

// Update handles the popup's internal updates
func (n *NotificationPopup) Update(msg tea.Msg) (NotificationPopup, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		n.width = msg.Width
		n.height = msg.Height

	case tea.KeyMsg:
		if n.visible {
			switch msg.String() {
			case "left", "h", "shift+tab":
				if n.selectedBtn > 0 {
					n.selectedBtn--
				}
			case "right", "l", "tab":
				if n.selectedBtn < len(n.buttons)-1 {
					n.selectedBtn++
				}
			case "enter", " ":
				if n.selectedBtn < len(n.buttons) {
					action := n.buttons[n.selectedBtn]
					n.Hide() // Hide after action
					return *n, NotificationActionCmd(action)
				}
			}
		}
	}

	return *n, nil
}

// View renders the notification popup
func (n *NotificationPopup) View() string {
	if !n.visible {
		return ""
	}

	// Get gradient colors based on notification type
	var gradientStart lipgloss.Color
	switch n.notifType {
	case NotificationSuccess:
		gradientStart = shared.SuccessGradientStart
	case NotificationError:
		gradientStart = shared.ErrorGradientStart
	case NotificationWarning:
		gradientStart = shared.WarningGradientStart
	case NotificationInfo:
		gradientStart = shared.InfoGradientStart
	default:
		gradientStart = shared.GradientStart
	}

	// Update box style with gradient border color
	boxStyle := notifBoxStyle.BorderForeground(gradientStart)

	var content []string

	// Title with gradient-like background
	if n.title != "" {
		titleStyle := notifTitleStyle.Width(52)
		// Use gradient start color for title background based on notification type
		titleStyle = titleStyle.Background(gradientStart)
		content = append(content, titleStyle.Render(n.title))
		content = append(content, "")
	}

	// Message
	if n.message != "" {
		content = append(content, notifMessageStyle.Width(52).Render(n.message))
	}

	// Render action buttons
	if len(n.buttons) > 0 {
		content = append(content, "")
		var buttons []string
		for i, btn := range n.buttons {
			var style lipgloss.Style
			if i == n.selectedBtn {
				style = notifActiveButtonStyle
			} else {
				style = notifButtonStyle
			}
			buttons = append(buttons, style.Render(btn))
		}
		buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, buttons...)
		content = append(content, lipgloss.PlaceHorizontal(54, lipgloss.Center, buttonRow))
	}

	// Join content
	boxContent := lipgloss.JoinVertical(lipgloss.Center, content...)

	// Create the popup box
	box := boxStyle.Render(boxContent)

	// Return just the box, overlay is handled by app.go
	return box
}

// Commands for notification popup

// ShowNotificationCmd shows the notification popup
func ShowNotificationCmd(notifType NotificationType, title, message string) tea.Cmd {
	return func() tea.Msg {
		return ShowNotificationMsg{
			Type:    notifType,
			Title:   title,
			Message: message,
		}
	}
}

// ShowNotificationWithButtonsCmd shows notification with custom buttons
func ShowNotificationWithButtonsCmd(notifType NotificationType, title, message string, buttons []string) tea.Cmd {
	return func() tea.Msg {
		return ShowNotificationWithButtonsMsg{
			Type:    notifType,
			Title:   title,
			Message: message,
			Buttons: buttons,
		}
	}
}

// NotificationActionCmd triggers when a button is pressed
func NotificationActionCmd(action string) tea.Cmd {
	return func() tea.Msg {
		return NotificationActionMsg{
			Action: action,
		}
	}
}

// Messages for notification popup

type ShowNotificationMsg struct {
	Type    NotificationType
	Title   string
	Message string
}

type ShowNotificationWithButtonsMsg struct {
	Type    NotificationType
	Title   string
	Message string
	Buttons []string
}

type NotificationActionMsg struct {
	Action string
}
