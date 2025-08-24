package tui

import (
	"fmt"

	"LogS/shared"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressPopup represents a modal progress overlay
type ProgressPopup struct {
	visible     bool
	title       string
	message     string
	progress    progress.Model
	spinner     spinner.Model
	percent     float64
	steps       []string
	currentStep int
	showSpinner bool
	width       int
	height      int

	// Action buttons
	buttons     []string
	selectedBtn int
	showButtons bool
}

// Popup styles using True Color RGB from shared
var (
	popupBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(shared.PrimaryColor).
			Background(shared.DialogBg).
			Padding(2, 3).
			Width(65)

	popupTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(shared.BrightText).
			Background(shared.BgLight).
			Padding(0, 2).
			MarginBottom(1).
			Align(lipgloss.Center)

	popupMessageStyle = lipgloss.NewStyle().
				Foreground(shared.SubduedText).
				Align(lipgloss.Center).
				MarginBottom(1)

	popupStepStyle = lipgloss.NewStyle().
			Foreground(shared.DimText).
			Italic(true)

	popupActiveStepStyle = lipgloss.NewStyle().
				Foreground(shared.InfoColor).
				Bold(true)

	popupCompletedStepStyle = lipgloss.NewStyle().
				Foreground(shared.SuccessColor)

	// Button styles with gradients
	buttonStyle = lipgloss.NewStyle().
			Foreground(shared.SubduedText).
			Background(shared.BgLight).
			Padding(0, 3).
			MarginTop(1).
			MarginRight(1)

	// Active button with gradient effect
	activeButtonStyle = lipgloss.NewStyle().
				Foreground(shared.BrightText).
				Background(shared.ButtonGradientStart).
				Padding(0, 3).
				MarginTop(1).
				MarginRight(1).
				Bold(true)
)

// NewProgressPopup creates a new progress popup
func NewProgressPopup() ProgressPopup {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(shared.PrimaryColor)

	p := progress.New(
		progress.WithGradient(string(shared.GradientStart), string(shared.GradientEnd)),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return ProgressPopup{
		visible:     false,
		progress:    p,
		spinner:     s,
		showSpinner: true,
		buttons:     []string{},
		selectedBtn: 0,
		showButtons: false,
	}
}

// Show displays the progress popup with a title and message
func (p *ProgressPopup) Show(title, message string) {
	p.visible = true
	p.title = title
	p.message = message
	p.percent = 0.0
	p.currentStep = 0
	p.steps = []string{}
	p.showButtons = false
}

// ShowWithButtons displays popup with action buttons
func (p *ProgressPopup) ShowWithButtons(title, message string, buttons []string) {
	p.visible = true
	p.title = title
	p.message = message
	p.buttons = buttons
	p.selectedBtn = 0
	p.showButtons = true
	p.showSpinner = false
}

// ShowWithSteps displays the progress popup with step tracking
func (p *ProgressPopup) ShowWithSteps(title string, steps []string) {
	p.visible = true
	p.title = title
	p.message = ""
	p.percent = 0.0
	p.currentStep = 0
	p.steps = steps
}

// SetComplete marks the progress as complete and shows OK button
func (p *ProgressPopup) SetComplete(message string) {
	p.percent = 1.0
	p.progress.SetPercent(1.0)
	p.message = message
	p.showSpinner = false
	p.buttons = []string{"OK"}
	p.selectedBtn = 0
	p.showButtons = true
}

// SetError shows error state with retry/cancel buttons
func (p *ProgressPopup) SetError(errorMsg string) {
	p.message = "❌ " + errorMsg
	p.showSpinner = false
	p.buttons = []string{"Retry", "Cancel"}
	p.selectedBtn = 0
	p.showButtons = true
}

// Hide hides the popup (only called after button action)
func (p *ProgressPopup) Hide() {
	p.visible = false
	p.percent = 0.0
	p.currentStep = 0
	p.showButtons = false
}

// SetProgress updates the progress percentage
func (p *ProgressPopup) SetProgress(percent float64) {
	p.percent = percent
	p.progress.SetPercent(percent)
}

// SetMessage updates the message
func (p *ProgressPopup) SetMessage(message string) {
	p.message = message
}

// NextStep moves to the next step
func (p *ProgressPopup) NextStep() {
	if p.currentStep < len(p.steps)-1 {
		p.currentStep++
		if len(p.steps) > 0 {
			p.percent = float64(p.currentStep+1) / float64(len(p.steps))
			p.progress.SetPercent(p.percent)
		}
	}
}

// Update handles the popup's internal updates
func (p *ProgressPopup) Update(msg tea.Msg) (ProgressPopup, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height

	case tea.KeyMsg:
		if p.visible && p.showButtons {
			switch msg.String() {
			case "left", "h", "shift+tab":
				if p.selectedBtn > 0 {
					p.selectedBtn--
				}
			case "right", "l", "tab":
				if p.selectedBtn < len(p.buttons)-1 {
					p.selectedBtn++
				}
			case "enter", " ":
				if p.selectedBtn < len(p.buttons) {
					action := p.buttons[p.selectedBtn]
					p.Hide() // Hide after action
					return *p, PopupActionCmd(action)
				}
			}
		}

	case spinner.TickMsg:
		if p.visible && p.showSpinner {
			var cmd tea.Cmd
			p.spinner, cmd = p.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case progress.FrameMsg:
		if p.visible {
			progressModel, cmd := p.progress.Update(msg)
			p.progress = progressModel.(progress.Model)
			cmds = append(cmds, cmd)
		}
	}

	return *p, tea.Batch(cmds...)
}

// View renders the progress popup
func (p *ProgressPopup) View() string {
	if !p.visible {
		return ""
	}

	var content []string

	// Title
	if p.title != "" {
		content = append(content, popupTitleStyle.Width(56).Render(p.title))
		content = append(content, "")
	}

	// Message or current step
	if p.message != "" {
		content = append(content, popupMessageStyle.Width(56).Render(p.message))
	} else if len(p.steps) > 0 && p.currentStep < len(p.steps) {
		currentStepText := p.steps[p.currentStep]
		if p.showSpinner {
			currentStepText = p.spinner.View() + " " + currentStepText
		}
		content = append(content, popupActiveStepStyle.Width(56).Render(currentStepText))
	}

	content = append(content, "")

	// Progress bar
	progressBar := p.progress.View()
	percentText := fmt.Sprintf("%.0f%%", p.percent*100)
	progressLine := lipgloss.JoinHorizontal(
		lipgloss.Top,
		progressBar,
		lipgloss.NewStyle().
			Foreground(shared.PrimaryColor).
			Bold(true).
			MarginLeft(2).
			Render(percentText),
	)
	content = append(content, progressLine)

	// Steps list (if available)
	if len(p.steps) > 0 {
		content = append(content, "")
		for i, step := range p.steps {
			prefix := "  ○ "
			style := popupStepStyle
			if i < p.currentStep {
				prefix = "  ✓ "
				style = popupCompletedStepStyle
			} else if i == p.currentStep {
				prefix = "  ▶ "
				style = popupActiveStepStyle
			}
			content = append(content, style.Render(prefix+step))
		}
	}

	// Render action buttons if available
	if p.showButtons && len(p.buttons) > 0 {
		content = append(content, "")
		var buttons []string
		for i, btn := range p.buttons {
			var style lipgloss.Style
			if i == p.selectedBtn {
				style = activeButtonStyle
			} else {
				style = buttonStyle
			}
			buttons = append(buttons, style.Render(btn))
		}
		buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, buttons...)
		content = append(content, lipgloss.PlaceHorizontal(59, lipgloss.Center, buttonRow))
	}

	// Join content
	boxContent := lipgloss.JoinVertical(lipgloss.Center, content...)

	// Create the popup box
	box := popupBoxStyle.Render(boxContent)

	// Return just the box, overlay is handled by app.go
	return box
}

// Commands for progress popup

// ShowProgressCmd shows the progress popup
func ShowProgressCmd(title, message string) tea.Cmd {
	return func() tea.Msg {
		return ShowProgressMsg{
			Title:   title,
			Message: message,
		}
	}
}

// ShowProgressWithStepsCmd shows progress popup with steps
func ShowProgressWithStepsCmd(title string, steps []string) tea.Cmd {
	return func() tea.Msg {
		return ShowProgressMsg{
			Title: title,
			Steps: steps,
		}
	}
}

// ShowProgressWithButtonsCmd shows popup with action buttons
func ShowProgressWithButtonsCmd(title, message string, buttons []string) tea.Cmd {
	return func() tea.Msg {
		return ShowProgressWithButtonsMsg{
			Title:   title,
			Message: message,
			Buttons: buttons,
		}
	}
}

// UpdateProgressCmd updates the progress
func UpdateProgressCmd(percent float64, message string) tea.Cmd {
	return func() tea.Msg {
		return UpdateProgressMsg{
			Percent: percent,
			Message: message,
		}
	}
}

// CompleteProgressCmd marks progress as complete with OK button
func CompleteProgressCmd(message string) tea.Cmd {
	return func() tea.Msg {
		return CompleteProgressMsg{
			Message: message,
		}
	}
}

// ErrorProgressCmd shows error state with retry/cancel buttons
func ErrorProgressCmd(errorMsg string) tea.Cmd {
	return func() tea.Msg {
		return ErrorProgressMsg{
			Error: errorMsg,
		}
	}
}

// PopupActionCmd triggers when a button is pressed
func PopupActionCmd(action string) tea.Cmd {
	return func() tea.Msg {
		return PopupActionMsg{
			Action: action,
		}
	}
}

// Messages for progress popup

type ShowProgressMsg struct {
	Title   string
	Message string
	Steps   []string
}

type ShowProgressWithButtonsMsg struct {
	Title   string
	Message string
	Buttons []string
}

type UpdateProgressMsg struct {
	Percent float64
	Message string
}

type CompleteProgressMsg struct {
	Message string
}

type ErrorProgressMsg struct {
	Error string
}

type PopupActionMsg struct {
	Action string
}

type NextStepMsg struct{}
