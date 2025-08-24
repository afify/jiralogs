package tui

import (
	"fmt"
	"strings"
	"time"

	"LogS/internal/tui/components/logo"
	"LogS/internal/tui/styles"
	"LogS/shared"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewType int

const (
	InitializationView ViewType = iota
	MainMenuView
	PeriodSelectionView
	StatsView // Main stats dashboard
	TicketDetailsView
	TimeLoggingView
	TicketCreationView
)

type AppModel struct {
	currentView  ViewType
	previousView ViewType
	viewStack    []ViewType

	client      JiraClientInterface
	service     WorklogServiceInterface
	ticketLogs  map[string]*shared.TicketLogData
	summary     *shared.SummaryData
	period      shared.Period
	currentUser *shared.JiraUser

	mainMenu          list.Model
	periodList        list.Model
	ticketTable       table.Model
	textInputs        []textinput.Model
	activeInput       int
	spinner           spinner.Model
	viewport          viewport.Model
	progressPopup     ProgressPopup     // Progress popup overlay
	notificationPopup NotificationPopup // Notification popup overlay

	loading         bool
	loadingMsg      string
	loadingSteps    []string
	currentStep     int
	progressPercent float64
	error           error
	errorDetails    string
	width           int
	height          int
	ready           bool

	// Initialization tracking
	initSteps       []string
	initCurrentStep int
	initError       error
	initialized     bool
	initStarted     bool
	initFunc        InitFunc
}

type errMsg struct {
	err     error
	details string
}

type ticketsLoadedMsg struct {
	ticketLogs map[string]*shared.TicketLogData
	summary    *shared.SummaryData
}

type progressMsg struct {
	percent float64
	step    string
}

type stepCompleteMsg struct{}

type userLoadedMsg struct {
	user *shared.JiraUser
}

type timeLoggingReadyMsg struct{}

type ticketCreationReadyMsg struct{}

type periodSelectionReadyMsg struct{}

func (e errMsg) Error() string { return e.err.Error() }

type InitFunc func() (JiraClientInterface, WorklogServiceInterface, error)

func NewApp(initFunc InitFunc) *AppModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return &AppModel{
		currentView:       InitializationView,
		spinner:           s,
		progressPopup:     NewProgressPopup(),
		notificationPopup: NewNotificationPopup(),
		initSteps: []string{
			"Loading environment configuration",
			"Connecting to JIRA API",
			"Authenticating user",
			"Initializing worklog service",
			"Setting up UI components",
		},
		initCurrentStep: 0,
		initialized:     false,
		initFunc:        initFunc,
	}
}

func (m *AppModel) setupUIComponents() {
	menuItems := []list.Item{
		menuItem{title: "📊 View Worklogs", desc: "Check your current worklog status"},
		menuItem{title: "⏰ Log Time", desc: "Log missing time to tickets"},
		menuItem{title: "🎫 Create Ticket", desc: "Create a new JIRA ticket"},
		menuItem{title: "📅 Filter Period", desc: "Change the date range filter"},
		menuItem{title: "🔄 Refresh", desc: "Reload data from JIRA"},
		menuItem{title: "❌ Quit", desc: "Exit the application"},
	}

	mainMenuList := list.New(menuItems, menuItemDelegate{}, 0, 0)
	mainMenuList.Title = "Main Menu"
	mainMenuList.SetShowStatusBar(false)
	mainMenuList.SetFilteringEnabled(false)
	mainMenuList.Styles.Title = ListTitleStyle

	periodItems := []list.Item{
		periodItem{id: "1", title: "Month to date", desc: "From 1st of current month"},
		periodItem{id: "3", title: "Last 3 months", desc: "Previous 3 months"},
		periodItem{id: "6", title: "Last 6 months", desc: "Previous 6 months"},
		periodItem{id: "9", title: "Last 9 months", desc: "Previous 9 months"},
		periodItem{id: "w", title: "Week to date", desc: "From Monday"},
		periodItem{id: "y", title: "Year to date", desc: "From January 1st"},
	}

	periodList := list.New(periodItems, periodItemDelegate{}, 0, 0)
	periodList.Title = "Select Period"
	periodList.SetShowStatusBar(false)
	periodList.SetFilteringEnabled(false)
	periodList.Styles.Title = ListTitleStyle

	now := time.Now()
	start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	m.mainMenu = mainMenuList
	m.periodList = periodList
	m.period = shared.Period{Start: start, End: now}
	m.ticketLogs = make(map[string]*shared.TicketLogData)
}

func (m *AppModel) Init() tea.Cmd {
	// Always get window size first, then start other operations
	return tea.Batch(
		tea.WindowSize(), // Get terminal size immediately
		m.spinner.Tick,
		// Don't start initialization yet - wait for window size
	)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Always handle window size first, regardless of view
	if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsMsg.Width
		m.height = wsMsg.Height
		m.ready = true

		// Log terminal dimensions
		shared.LogError("TERMINAL_SIZE", fmt.Errorf("Terminal resized - Width: %d, Height: %d", m.width, m.height))

		frameH, frameV := ContentStyle.GetFrameSize()
		contentWidth := wsMsg.Width - frameH
		contentHeight := wsMsg.Height - frameV - 12

		shared.LogError("CONTENT_AREA", fmt.Errorf("Content area after frame - Width: %d, Height: %d (frameH: %d, frameV: %d)", contentWidth, contentHeight, frameH, frameV))

		if contentHeight < 5 {
			contentHeight = 5
		}
		if contentWidth < 20 {
			contentWidth = 20
		}

		// Only update sizes if components are initialized
		if m.initialized {
			m.mainMenu.SetSize(contentWidth, contentHeight)
			m.periodList.SetSize(contentWidth, contentHeight)
		}

		if m.viewport.Width == 0 {
			m.viewport = viewport.New(contentWidth, contentHeight-5)
		} else {
			m.viewport.Width = contentWidth
			m.viewport.Height = contentHeight - 5
		}

		if m.ticketTable.Columns() != nil {
			tableHeight := contentHeight - 8
			if tableHeight < 5 {
				tableHeight = 5
			}
			m.ticketTable.SetHeight(tableHeight)
		}

		// Start initialization after we have window dimensions
		if m.currentView == InitializationView && !m.initStarted {
			m.initStarted = true
			return m, m.startInitialization()
		}
	}

	// Handle other message types
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't process keys if popups are visible
		if m.progressPopup.visible || m.notificationPopup.visible {
			if m.progressPopup.visible {
				popup, cmd := m.progressPopup.Update(msg)
				m.progressPopup = popup
				return m, cmd
			}
			if m.notificationPopup.visible {
				popup, cmd := m.notificationPopup.Update(msg)
				m.notificationPopup = popup
				return m, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q", "Q":
			// Quit the application globally
			return m, tea.Quit
		case "h", "H":
			// Return to home (stats page)
			if m.currentView != StatsView && m.currentView != InitializationView {
				m.viewStack = []ViewType{}
				m.currentView = StatsView
			}
		case "esc":
			if len(m.viewStack) > 0 && m.currentView != MainMenuView {
				m.previousView = m.currentView
				m.currentView = m.viewStack[len(m.viewStack)-1]
				m.viewStack = m.viewStack[:len(m.viewStack)-1]
			}
		case "?":
			m.pushView(MainMenuView)
		}

		switch m.currentView {
		case MainMenuView:
			cmd = m.handleMainMenuKeys(msg)
		case PeriodSelectionView:
			cmd = m.handlePeriodSelectionKeys(msg)
		case StatsView:
			cmd = m.HandleStatsInput(msg)
		case TicketDetailsView:
			cmd = m.HandleTicketDetailsInput(msg)
		case TimeLoggingView:
			cmd = m.handleTimeLoggingKeys(msg)
		}

		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case initStepMsg:
		m.initCurrentStep++
		m.progressPercent = float64(m.initCurrentStep) / float64(len(m.initSteps))
		return m, nil

	case initCompleteMsg:
		m.client = msg.client
		m.service = msg.service
		m.initialized = true
		m.setupUIComponents()

		// Set up year-to-date period
		now := time.Now()
		yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		m.period = shared.Period{Start: yearStart, End: now}

		// Load worklogs immediately with progress popup
		m.currentView = StatsView // Set to stats view immediately
		m.progressPopup.ShowWithSteps("Loading Worklogs", []string{
			"Fetching user information",
			"Loading year-to-date worklogs",
			"Processing ticket data",
			"Calculating statistics",
		})

		return m, tea.Batch(
			m.loadCurrentUser(),
			LoadWorklogsWithProgress(m.service, m.period),
		)

	case initErrorMsg:
		m.initError = msg.err
		m.error = msg.err
		m.errorDetails = msg.details
		m.progressPopup.Hide()
		m.notificationPopup.ShowError("Initialization Error", msg.err.Error())

	case errMsg:
		m.error = msg.err
		m.errorDetails = msg.details
		m.loading = false
		m.progressPopup.Hide()
		m.notificationPopup.ShowError("Error", msg.err.Error())

	case ticketsLoadedMsg:
		m.ticketLogs = msg.ticketLogs
		m.summary = msg.summary
		m.loading = false
		m.progressPercent = 1.0
		// Hide progress popup
		m.progressPopup.Hide()
		// Go directly to stats view instead of menu
		m.currentView = StatsView

	case userLoadedMsg:
		m.currentUser = msg.user

	case timeLoggingReadyMsg:
		m.loading = false
		m.progressPopup.Hide()
		m.currentView = TimeLoggingView

	case ticketCreationReadyMsg:
		m.loading = false
		m.progressPopup.Hide()
		m.currentView = TicketCreationView

	case periodSelectionReadyMsg:
		m.loading = false
		m.progressPopup.Hide()
		m.currentView = PeriodSelectionView

	case ShowProgressMsg:
		if len(msg.Steps) > 0 {
			m.progressPopup.ShowWithSteps(msg.Title, msg.Steps)
		} else {
			m.progressPopup.Show(msg.Title, msg.Message)
		}

	case ShowProgressWithButtonsMsg:
		m.progressPopup.ShowWithButtons(msg.Title, msg.Message, msg.Buttons)

	case UpdateProgressMsg:
		m.progressPopup.SetProgress(msg.Percent)
		if msg.Message != "" {
			m.progressPopup.SetMessage(msg.Message)
		}

	case CompleteProgressMsg:
		m.progressPopup.SetComplete(msg.Message)

	case ErrorProgressMsg:
		m.progressPopup.SetError(msg.Error)

	case PopupActionMsg:
		// Handle popup button actions
		switch msg.Action {
		case "OK":
			// Just hide the popup for OK
			m.progressPopup.Hide()
		case "Retry":
			// Retry the last operation
			m.progressPopup.Hide()
			// Show progress popup and retry loading worklogs
			m.progressPopup.ShowWithSteps("Retrying", []string{
				"Reconnecting to JIRA",
				"Fetching worklogs",
				"Processing data",
			})
			return m, LoadWorklogsWithProgress(m.service, m.period)
		case "Cancel":
			// Cancel and go back
			m.progressPopup.Hide()
			if len(m.viewStack) > 0 {
				m.currentView = m.viewStack[len(m.viewStack)-1]
				m.viewStack = m.viewStack[:len(m.viewStack)-1]
			}
		}

	case NextStepMsg:
		m.progressPopup.NextStep()

	// Notification popup messages
	case ShowNotificationMsg:
		switch msg.Type {
		case NotificationSuccess:
			m.notificationPopup.ShowSuccess(msg.Title, msg.Message)
		case NotificationError:
			m.notificationPopup.ShowError(msg.Title, msg.Message)
		case NotificationWarning:
			m.notificationPopup.ShowWarning(msg.Title, msg.Message)
		case NotificationInfo:
			m.notificationPopup.ShowInfo(msg.Title, msg.Message)
		}

	case ShowNotificationWithButtonsMsg:
		m.notificationPopup.ShowWithButtons(msg.Type, msg.Title, msg.Message, msg.Buttons)

	case NotificationActionMsg:
		// Handle notification button actions
		switch msg.Action {
		case "OK":
			// Just hide the notification
			m.notificationPopup.Hide()
		case "Retry":
			// Retry the last operation
			m.notificationPopup.Hide()
			// Show progress popup and retry loading worklogs
			m.progressPopup.ShowWithSteps("Retrying", []string{
				"Reconnecting to JIRA",
				"Fetching worklogs",
				"Processing data",
			})
			return m, LoadWorklogsWithProgress(m.service, m.period)
		case "Cancel":
			// Cancel and go back
			m.notificationPopup.Hide()
			if len(m.viewStack) > 0 {
				m.currentView = m.viewStack[len(m.viewStack)-1]
				m.viewStack = m.viewStack[:len(m.viewStack)-1]
			}
		}

	case progressMsg:
		m.progressPercent = msg.percent
		m.loadingMsg = msg.step
		// Update progress popup if visible
		m.progressPopup.SetProgress(msg.percent)
		if msg.step != "" {
			m.progressPopup.SetMessage(msg.step)
		}

	case stepCompleteMsg:
		m.currentStep++
		if m.currentStep < len(m.loadingSteps) {
			m.loadingMsg = m.loadingSteps[m.currentStep]
		}

	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		// Also update progress popup spinner
		if m.progressPopup.visible {
			popup, cmd := m.progressPopup.Update(msg)
			m.progressPopup = popup
			cmds = append(cmds, cmd)
		}

	case progress.FrameMsg:
		// Update progress popup's progress bar
		if m.progressPopup.visible {
			popup, cmd := m.progressPopup.Update(msg)
			m.progressPopup = popup
			cmds = append(cmds, cmd)
		}
	}

	switch m.currentView {
	case InitializationView:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case MainMenuView:
		m.mainMenu, cmd = m.mainMenu.Update(msg)
		cmds = append(cmds, cmd)
	case PeriodSelectionView:
		m.periodList, cmd = m.periodList.Update(msg)
		cmds = append(cmds, cmd)
	case StatsView:
		// Stats view no longer has table or viewport
	case TicketDetailsView:
		// Ticket details handles its own table updates
		m.ticketTable, cmd = m.ticketTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *AppModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// Special handling for initialization view
	if m.currentView == InitializationView {
		return m.renderInitializationView()
	}

	shared.LogError("VIEW_RENDER", fmt.Errorf("Starting render for view: %v, Width: %d, Height: %d", m.currentView, m.width, m.height))

	// Use the Crush-style logo with Crush's theme
	t := styles.CurrentTheme()
	logoOpts := logo.Opts{
		FieldColor:   t.Primary,
		TitleColorA:  t.Secondary,
		TitleColorB:  t.Primary,
		CharmColor:   t.Primary,
		VersionColor: t.FgMuted,
		Width:        0,
	}
	banner := logo.Render("v1.0.0", false, logoOpts)
	bannerHeight := strings.Count(banner, "\n") + 1
	shared.LogError("COMPONENT_BANNER", fmt.Errorf("Banner height: %d lines", bannerHeight))

	// Get page title based on current view
	var pageTitle string
	switch m.currentView {
	case StatsView:
		pageTitle = "📊 WORKLOG STATISTICS"
	case MainMenuView:
		pageTitle = "🏠 MAIN MENU"
	case PeriodSelectionView:
		pageTitle = "📅 SELECT PERIOD"
	case TicketDetailsView:
		pageTitle = "📋 TICKET DETAILS"
	case TimeLoggingView:
		pageTitle = "⏰ LOG TIME"
	case TicketCreationView:
		pageTitle = "🎫 CREATE TICKET"
	}

	var content string
	switch m.currentView {
	case MainMenuView:
		content = m.renderMainMenuView()
	case PeriodSelectionView:
		content = m.renderPeriodSelectionView()
	case StatsView:
		content = m.renderStatsView()
	case TicketDetailsView:
		content = m.renderTicketDetailsView()
	case TimeLoggingView:
		content = m.renderTimeLoggingView()
	case TicketCreationView:
		content = m.renderTicketCreationView()
	}

	contentHeight := strings.Count(content, "\n") + 1
	shared.LogError("COMPONENT_CONTENT", fmt.Errorf("Content height for %v: %d lines", m.currentView, contentHeight))

	help := m.renderHelpLine()
	helpHeight := strings.Count(help, "\n") + 1
	shared.LogError("COMPONENT_HELP", fmt.Errorf("Help line height: %d lines", helpHeight))

	// Calculate available width for proper spacing, accounting for frame
	// Frame adds 2 chars for left/right borders + 2 chars for padding on each side = 4 total
	frameHorizontalSpace := 4
	availableWidth := m.width - frameHorizontalSpace

	shared.LogError("WIDTH_CALC", fmt.Errorf("Terminal width: %d, Frame H space: %d, Available width: %d",
		m.width, frameHorizontalSpace, availableWidth))

	// Use just the banner
	topBar := banner
	topBarHeight := strings.Count(topBar, "\n") + 1
	shared.LogError("COMPONENT_TOPBAR", fmt.Errorf("TopBar height: %d lines", topBarHeight))

	// Create a nice separator line
	separatorStyle := lipgloss.NewStyle().
		Foreground(shared.DarkGray).
		Width(availableWidth)
	separator := separatorStyle.Render(strings.Repeat("─", availableWidth))
	separatorHeight := strings.Count(separator, "\n") + 1
	shared.LogError("COMPONENT_SEPARATOR", fmt.Errorf("Separator height: %d lines", separatorHeight))

	// Create page title
	pageTitleRendered := ""
	if pageTitle != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(shared.PrimaryColor).
			Bold(true).
			Background(shared.BgLight).
			Padding(0, 3)
		pageTitleRendered = lipgloss.PlaceHorizontal(availableWidth, lipgloss.Center, titleStyle.Render(pageTitle))
	}
	pageTitleHeight := strings.Count(pageTitleRendered, "\n") + 1
	shared.LogError("COMPONENT_PAGE_TITLE", fmt.Errorf("Page title height: %d lines", pageTitleHeight))

	// Calculate space needed to push help to bottom
	// Don't add 1 to the last component since we're counting newlines
	usedHeight := strings.Count(topBar, "\n") + 1 +
		strings.Count(separator, "\n") + 1 +
		strings.Count(pageTitleRendered, "\n") + 1 +
		strings.Count(content, "\n") + 1 +
		strings.Count(help, "\n") // No +1 for the last item

	// Account for frame (adds 4 lines total - top/bottom borders and padding)
	frameHeight := 4 // Frame adds top border, bottom border, and padding lines
	availableHeight := m.height - frameHeight
	spacerLines := availableHeight - usedHeight - 2 // Subtract 2 more to account for content being 2 lines too tall

	shared.LogError("SPACER_CALC", fmt.Errorf("Used: %d, Available: %d, Spacer lines needed: %d",
		usedHeight, availableHeight, spacerLines))

	// Ensure we don't overflow the terminal
	if spacerLines < 0 {
		spacerLines = 0
	}

	// Create vertical spacer to push help to bottom
	verticalSpacer := ""
	if spacerLines > 0 {
		verticalSpacer = strings.Repeat("\n", spacerLines)
	}

	// Log the exact position where footer will appear
	footerStartLine := usedHeight + spacerLines
	shared.LogError("FOOTER_LOCATION", fmt.Errorf("Footer will be at line %d of %d (with %d spacer lines)",
		footerStartLine, availableHeight, spacerLines))

	fullContent := lipgloss.JoinVertical(lipgloss.Left,
		topBar,
		separator,
		pageTitleRendered,
		content,
		verticalSpacer,
		help,
	)

	totalHeight := strings.Count(fullContent, "\n") + 1
	shared.LogError("TOTAL_HEIGHT", fmt.Errorf("Total content height before frame: %d lines", totalHeight))

	// Render the base content
	baseView := FrameStyle.Render(fullContent)

	finalHeight := strings.Count(baseView, "\n") + 1
	shared.LogError("FINAL_HEIGHT", fmt.Errorf("Final height after frame: %d lines, Terminal height: %d", finalHeight, m.height))

	// If progress popup is visible, overlay it on top
	if m.progressPopup.visible {
		return m.overlayPopup(baseView, m.progressPopup.View())
	}

	// If notification popup is visible, overlay it on top
	if m.notificationPopup.visible {
		return m.overlayPopup(baseView, m.notificationPopup.View())
	}

	return baseView
}

// overlayPopup renders a popup over the base view
func (m *AppModel) overlayPopup(baseView, popupView string) string {
	if m.width == 0 || m.height == 0 || popupView == "" {
		return baseView
	}

	// Split views into lines
	baseLines := strings.Split(baseView, "\n")
	popupLines := strings.Split(popupView, "\n")

	// Center the popup
	popupHeight := len(popupLines)
	popupWidth := 0
	for _, line := range popupLines {
		if w := lipgloss.Width(line); w > popupWidth {
			popupWidth = w
		}
	}

	// Calculate popup position
	startY := (m.height - popupHeight) / 2
	startX := (m.width - popupWidth) / 2

	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	// Create result lines, starting with base view
	result := make([]string, len(baseLines))
	copy(result, baseLines)

	// Ensure we have enough lines
	for len(result) < m.height {
		result = append(result, "")
	}

	// Overlay popup lines
	for i, popupLine := range popupLines {
		targetY := startY + i
		if targetY >= 0 && targetY < len(result) {
			// Get the base line
			baseLine := result[targetY]
			baseWidth := lipgloss.Width(baseLine)

			// Pad base line if needed
			if baseWidth < startX {
				baseLine = baseLine + strings.Repeat(" ", startX-baseWidth)
			}

			// Create the overlaid line
			if startX > 0 && len(baseLine) > 0 {
				// Keep part of base before popup
				before := baseLine
				if lipgloss.Width(before) > startX {
					before = string([]rune(before)[:startX])
				}
				result[targetY] = before + popupLine
			} else {
				result[targetY] = popupLine
			}
		}
	}

	return strings.Join(result, "\n")
}

func (m *AppModel) pushView(view ViewType) {
	m.viewStack = append(m.viewStack, m.currentView)
	m.previousView = m.currentView
	m.currentView = view
}
