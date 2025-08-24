package tui

import (
	"time"

	"LogS/internal/app"
	"LogS/internal/tui/components/core"
	"LogS/internal/tui/components/core/layout"
	"LogS/internal/tui/components/core/status"
	"LogS/internal/tui/components/dialogs"
	"LogS/internal/tui/page"
	"LogS/internal/tui/page/calendar"
	"LogS/internal/tui/page/stats"
	"LogS/internal/tui/page/tickets"
	"LogS/internal/tui/page/welcome"
	"LogS/internal/tui/page/worklog"
	"LogS/internal/tui/styles"
	"LogS/internal/tui/util"
	"LogS/shared"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

var lastMouseEvent time.Time

func MouseEventFilter(m tea.Model, msg tea.Msg) tea.Msg {
	switch msg.(type) {
	case tea.MouseWheelMsg, tea.MouseMotionMsg:
		now := time.Now()
		// trackpad is sending too many requests
		if now.Sub(lastMouseEvent) < 15*time.Millisecond {
			return nil
		}
		lastMouseEvent = now
	}
	return msg
}

// appModel represents the main application model that manages pages, dialogs, and UI state.
type appModel struct {
	wWidth, wHeight int // Window dimensions
	width, height   int
	keyMap          KeyMap

	currentPage  page.PageID
	previousPage page.PageID
	pages        map[page.PageID]util.Model
	loadedPages  map[page.PageID]bool

	// Status
	status          status.StatusCmp
	showingFullHelp bool

	app *app.App

	dialog       dialogs.DialogCmp
	isConfigured bool

	// Worklog Page Specific
	selectedWorklogID string // The ID of the currently selected worklog
}

// Init initializes the application model and returns initial commands.
func (a appModel) Init() tea.Cmd {
	shared.LogErrorf("TUI_INIT", "Initializing TUI application model")
	var cmds []tea.Cmd
	cmd := a.pages[a.currentPage].Init()
	cmds = append(cmds, cmd)
	a.loadedPages[a.currentPage] = true
	shared.LogErrorf("TUI_INIT", "Initialized page: %s", string(a.currentPage))

	cmd = a.status.Init()
	cmds = append(cmds, cmd)
	shared.LogErrorf("TUI_INIT", "Status component initialized")

	cmds = append(cmds, tea.EnableMouseAllMotion)
	shared.LogErrorf("TUI_INIT", "Mouse motion enabled")

	return tea.Batch(cmds...)
}

// Update handles incoming messages and updates the application state.
func (a *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	// TODO: Implement config check similar to crush
	a.isConfigured = true // For now

	switch msg := msg.(type) {
	case tea.KeyboardEnhancementsMsg:
		for id, page := range a.pages {
			m, pageCmd := page.Update(msg)
			a.pages[id] = m.(util.Model)
			if pageCmd != nil {
				cmds = append(cmds, pageCmd)
			}
		}
		return a, tea.Batch(cmds...)
	case tea.WindowSizeMsg:
		a.wWidth, a.wHeight = msg.Width, msg.Height
		return a, a.handleWindowResize(msg.Width, msg.Height)

	// Dialog messages
	case dialogs.OpenDialogMsg, dialogs.CloseDialogMsg:
		u, dialogCmd := a.dialog.Update(msg)
		a.dialog = u.(dialogs.DialogCmp)
		return a, dialogCmd

	// Page change messages
	case page.PageChangeMsg:
		cmd := a.moveToPage(msg)
		return a, cmd

	// Status Messages
	case util.InfoMsg, util.ClearStatusMsg:
		s, statusCmd := a.status.Update(msg)
		a.status = s.(status.StatusCmp)
		cmds = append(cmds, statusCmd)
		return a, tea.Batch(cmds...)

	// Worklog data message for stats page
	case welcome.WorklogDataMsg:
		shared.LogErrorf("TUI_UPDATE", "Received WorklogDataMsg, passing to current page: %s", string(a.currentPage))
		item, ok := a.pages[a.currentPage]
		if ok {
			updated, cmd := item.Update(msg)
			a.pages[a.currentPage] = updated.(util.Model)
			return a, cmd
		}
		return a, nil

	case tea.KeyPressMsg:
		return a, a.handleKeyPressMsg(msg)

	case tea.MouseWheelMsg:
		if a.dialog.HasDialogs() {
			u, dialogCmd := a.dialog.Update(msg)
			a.dialog = u.(dialogs.DialogCmp)
			cmds = append(cmds, dialogCmd)
		} else {
			item, ok := a.pages[a.currentPage]
			if !ok {
				return a, nil
			}

			updated, pageCmd := item.Update(msg)
			a.pages[a.currentPage] = updated.(util.Model)
			cmds = append(cmds, pageCmd)
		}
		return a, tea.Batch(cmds...)
	case tea.PasteMsg:
		if a.dialog.HasDialogs() {
			u, dialogCmd := a.dialog.Update(msg)
			a.dialog = u.(dialogs.DialogCmp)
			cmds = append(cmds, dialogCmd)
		} else {
			item, ok := a.pages[a.currentPage]
			if !ok {
				return a, nil
			}

			updated, pageCmd := item.Update(msg)
			a.pages[a.currentPage] = updated.(util.Model)
			cmds = append(cmds, pageCmd)
		}
		return a, tea.Batch(cmds...)
	}
	
	// Log all other message types for debugging
	shared.LogErrorf("TUI_UPDATE", "Received message type: %T for page: %s", msg, string(a.currentPage))
	
	s, _ := a.status.Update(msg)
	a.status = s.(status.StatusCmp)

	item, ok := a.pages[a.currentPage]
	if !ok {
		return a, nil
	}
	updated, cmd := item.Update(msg)
	a.pages[a.currentPage] = updated.(util.Model)

	if a.dialog.HasDialogs() {
		u, dialogCmd := a.dialog.Update(msg)
		a.dialog = u.(dialogs.DialogCmp)
		cmds = append(cmds, dialogCmd)
	}
	cmds = append(cmds, cmd)
	return a, tea.Batch(cmds...)
}

// handleWindowResize processes window resize events and updates all components.
func (a *appModel) handleWindowResize(width, height int) tea.Cmd {
	var cmds []tea.Cmd
	if a.showingFullHelp {
		height -= 5
	} else {
		height -= 2
	}
	a.width, a.height = width, height
	// Update status bar
	s, cmd := a.status.Update(tea.WindowSizeMsg{Width: width, Height: height})
	a.status = s.(status.StatusCmp)
	cmds = append(cmds, cmd)

	// Update the current page
	for p, page := range a.pages {
		updated, pageCmd := page.Update(tea.WindowSizeMsg{Width: width, Height: height})
		a.pages[p] = updated.(util.Model)
		cmds = append(cmds, pageCmd)
	}

	// Update the dialogs
	dialog, cmd := a.dialog.Update(tea.WindowSizeMsg{Width: width, Height: height})
	a.dialog = dialog.(dialogs.DialogCmp)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// handleKeyPressMsg processes keyboard input and routes to appropriate handlers.
func (a *appModel) handleKeyPressMsg(msg tea.KeyPressMsg) tea.Cmd {
	if a.dialog.HasDialogs() {
		u, dialogCmd := a.dialog.Update(msg)
		a.dialog = u.(dialogs.DialogCmp)
		return dialogCmd
	}
	switch {
	// help
	case key.Matches(msg, a.keyMap.Help):
		a.status.ToggleFullHelp()
		a.showingFullHelp = !a.showingFullHelp
		return a.handleWindowResize(a.wWidth, a.wHeight)
	// dialogs
	case key.Matches(msg, a.keyMap.Quit):
		return tea.Quit
	default:
		item, ok := a.pages[a.currentPage]
		if !ok {
			return nil
		}

		updated, cmd := item.Update(msg)
		a.pages[a.currentPage] = updated.(util.Model)
		return cmd
	}
}

// moveToPage handles navigation between different pages in the application.
func (a *appModel) moveToPage(msg page.PageChangeMsg) tea.Cmd {
	var cmds []tea.Cmd
	pageID := msg.ID
	
	shared.LogErrorf("TUI_NAVIGATE", "Moving to page: %s, hasData: %v", string(pageID), msg.Data != nil)
	
	if _, ok := a.loadedPages[pageID]; !ok {
		cmd := a.pages[pageID].Init()
		cmds = append(cmds, cmd)
		a.loadedPages[pageID] = true
	}
	a.previousPage = a.currentPage
	a.currentPage = pageID
	
	// Handle data passing to the new page
	if msg.Data != nil && (pageID == "stats" || pageID == "tickets" || pageID == "calendar") {
		shared.LogErrorf("TUI_NAVIGATE", "Passing data to %s page", string(pageID))
		if worklogData, ok := msg.Data.(welcome.WorklogDataMsg); ok {
			shared.LogErrorf("TUI_NAVIGATE", "Found WorklogDataMsg, setting %s data - TicketLogs: %d, DailyHours: %d, DailyTickets: %d",
				string(pageID), len(worklogData.TicketLogs), len(worklogData.DailyHours), len(worklogData.DailyTickets))
			// Send the data message to the target page
			cmds = append(cmds, func() tea.Msg {
				return worklogData
			})
		}
	}
	
	if sizable, ok := a.pages[a.currentPage].(layout.Sizeable); ok {
		cmd := sizable.SetSize(a.width, a.height)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// View renders the complete application interface including pages, dialogs, and overlays.
func (a *appModel) View() tea.View {
	var view tea.View
	t := styles.CurrentTheme()
	view.BackgroundColor = t.BgBase
	if a.wWidth < 25 || a.wHeight < 15 {
		view.Layer = lipgloss.NewCanvas(
			lipgloss.NewLayer(
				t.S().Base.Width(a.wWidth).Height(a.wHeight).
					Align(lipgloss.Center, lipgloss.Center).
					Render(
						t.S().Base.
							Padding(1, 4).
							Foreground(t.White).
							BorderStyle(lipgloss.RoundedBorder()).
							BorderForeground(t.Primary).
							Render("Window too small!"),
					),
			),
		)
		return view
	}

	shared.LogErrorf("TUI_VIEW", "Rendering page: %s", string(a.currentPage))
	page := a.pages[a.currentPage]
	if withHelp, ok := page.(core.KeyMapHelp); ok {
		a.status.SetKeyMap(withHelp.Help())
	}
	pageView := page.View()
	components := []string{
		pageView,
	}
	components = append(components, a.status.View())

	appView := lipgloss.JoinVertical(lipgloss.Top, components...)
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(appView),
	}
	if a.dialog.HasDialogs() {
		layers = append(
			layers,
			a.dialog.GetLayers()...,
		)
	}

	var cursor *tea.Cursor
	if v, ok := page.(util.Cursor); ok {
		cursor = v.Cursor()
	}
	activeView := a.dialog.ActiveModel()
	if activeView != nil {
		cursor = nil // Reset cursor if a dialog is active unless it implements util.Cursor
		if v, ok := activeView.(util.Cursor); ok {
			cursor = v.Cursor()
		}
	}

	canvas := lipgloss.NewCanvas(
		layers...,
	)

	view.Layer = canvas
	view.Cursor = cursor
	return view
}

// New creates and initializes a new TUI application model.
func New(app *app.App) tea.Model {
	shared.LogErrorf("TUI_NEW", "Creating new TUI application model")
	welcomePage := welcome.New(app)
	shared.LogErrorf("TUI_NEW", "Welcome page created")
	
	statsPage := stats.New(app)
	shared.LogErrorf("TUI_NEW", "Stats page created")
	
	calendarPage := calendar.New(app)
	shared.LogErrorf("TUI_NEW", "Calendar page created")
	
	ticketsPage := tickets.New(app)
	shared.LogErrorf("TUI_NEW", "Tickets page created")
	
	worklogPage := worklog.New(app)
	shared.LogErrorf("TUI_NEW", "Worklog page created")
	
	keyMap := DefaultKeyMap()
	keyMap.pageBindings = welcomePage.Bindings()
	shared.LogErrorf("TUI_NEW", "Key mappings configured")

	model := &appModel{
		currentPage: welcome.WelcomePageID,
		app:         app,
		status:      status.NewStatusCmp(),
		loadedPages: make(map[page.PageID]bool),
		keyMap:      keyMap,

		pages: map[page.PageID]util.Model{
			welcome.WelcomePageID:   welcomePage,
			stats.StatsPageID:       statsPage,
			calendar.CalendarPageID: calendarPage,
			tickets.TicketsPageID:   ticketsPage,
			worklog.WorklogPageID:   worklogPage,
		},

		dialog: dialogs.NewDialogCmp(),
	}

	shared.LogErrorf("TUI_NEW", "TUI application model created successfully with page: %s", string(welcome.WelcomePageID))
	return model
}
