package tui

import (
	"time"

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
	MainMenuView ViewType = iota
	PeriodSelectionView
	WorklogDisplayView
	TimeLoggingView
	TicketCreationView
	LoadingView
	ErrorView
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

	mainMenu    list.Model
	periodList  list.Model
	ticketTable table.Model
	textInputs  []textinput.Model
	activeInput int
	spinner     spinner.Model
	progressBar progress.Model
	viewport    viewport.Model

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

func NewApp(client JiraClientInterface, service WorklogServiceInterface) *AppModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

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

	prog := progress.New(progress.WithDefaultGradient())

	now := time.Now()
	start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	return &AppModel{
		currentView:  MainMenuView,
		client:       client,
		service:      service,
		mainMenu:     mainMenuList,
		periodList:   periodList,
		spinner:      s,
		progressBar:  prog,
		period:       shared.Period{Start: start, End: now},
		ticketLogs:   make(map[string]*shared.TicketLogData),
		loadingSteps: []string{},
	}
}

func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadCurrentUser(),
	)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		frameH, frameV := ContentStyle.GetFrameSize()
		contentWidth := msg.Width - frameH
		contentHeight := msg.Height - frameV - 12

		if contentHeight < 5 {
			contentHeight = 5
		}
		if contentWidth < 20 {
			contentWidth = 20
		}

		m.mainMenu.SetSize(contentWidth, contentHeight)
		m.periodList.SetSize(contentWidth, contentHeight)

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

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
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
		case WorklogDisplayView:
			cmd = m.handleWorklogDisplayKeys(msg)
		case TimeLoggingView:
			cmd = m.handleTimeLoggingKeys(msg)
		}

		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case errMsg:
		m.error = msg.err
		m.errorDetails = msg.details
		m.loading = false
		m.currentView = ErrorView

	case ticketsLoadedMsg:
		m.ticketLogs = msg.ticketLogs
		m.summary = msg.summary
		m.loading = false
		m.progressPercent = 1.0
		m.currentView = MainMenuView

	case userLoadedMsg:
		m.currentUser = msg.user

	case timeLoggingReadyMsg:
		m.loading = false
		m.currentView = TimeLoggingView

	case ticketCreationReadyMsg:
		m.loading = false
		m.currentView = TicketCreationView

	case periodSelectionReadyMsg:
		m.loading = false
		m.currentView = PeriodSelectionView

	case progressMsg:
		m.progressPercent = msg.percent
		m.loadingMsg = msg.step
		cmd := m.progressBar.SetPercent(msg.percent)
		cmds = append(cmds, cmd)

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
	}

	switch m.currentView {
	case MainMenuView:
		m.mainMenu, cmd = m.mainMenu.Update(msg)
		cmds = append(cmds, cmd)
	case PeriodSelectionView:
		m.periodList, cmd = m.periodList.Update(msg)
		cmds = append(cmds, cmd)
	case WorklogDisplayView:
		m.ticketTable, cmd = m.ticketTable.Update(msg)
		cmds = append(cmds, cmd)
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *AppModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	banner := RenderBanner()
	header := m.renderHeaderInfo()

	var content string
	switch m.currentView {
	case LoadingView:
		content = m.renderLoadingView()
	case ErrorView:
		content = m.renderErrorView()
	case MainMenuView:
		content = m.renderMainMenuView()
	case PeriodSelectionView:
		content = m.renderPeriodSelectionView()
	case WorklogDisplayView:
		content = m.renderWorklogDisplayView()
	case TimeLoggingView:
		content = m.renderTimeLoggingView()
	case TicketCreationView:
		content = m.renderTicketCreationView()
	}

	help := m.renderHelpLine()

	bannerWithHeader := lipgloss.JoinHorizontal(lipgloss.Top, banner, header)

	fullContent := lipgloss.JoinVertical(lipgloss.Left,
		bannerWithHeader,
		"",
		content,
		"",
		help,
	)

	return FrameStyle.Render(fullContent)
}

func (m *AppModel) pushView(view ViewType) {
	m.viewStack = append(m.viewStack, m.currentView)
	m.previousView = m.currentView
	m.currentView = view
}
