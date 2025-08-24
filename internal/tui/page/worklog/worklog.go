package worklog

import (
	"time"

	"LogS/internal/app"
	"LogS/internal/config"
	"LogS/internal/tui/components/core"
	"LogS/internal/tui/components/core/layout"
	"LogS/internal/tui/page"
	"LogS/internal/tui/util"
	"LogS/shared"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
)

var WorklogPageID page.PageID = "worklog"

type (
	WorklogFocusedMsg struct {
		Focused bool
	}
	CancelTimerExpiredMsg struct{}
)

type PanelType string

const (
	PanelTypeList   PanelType = "list"
	PanelTypeDetail PanelType = "detail"
)

const (
	CompactModeWidthBreakpoint  = 120
	CompactModeHeightBreakpoint = 30
	SideBarWidth                = 31
	HeaderHeight                = 1
	BorderWidth                 = 1
	CancelTimerDuration         = 2 * time.Second
)

type WorklogPage interface {
	util.Model
	layout.Help
	IsListFocused() bool
}

type worklogPage struct {
	width, height        int
	app                  *app.App
	keyboardEnhancements tea.KeyboardEnhancementsMsg

	// Layout state
	compact     bool
	focusedPane PanelType

	keyMap KeyMap

	// Components
	// TODO: Add your worklog list, detail view, etc.

	// Simple state flags
	showingDetails bool
	isCanceling    bool
	isConfigured   bool
}

func New(app *app.App) WorklogPage {
	shared.LogErrorf("WORKLOG_NEW", "Creating new worklog page")
	return &worklogPage{
		app:         app,
		keyMap:      DefaultKeyMap(),
		focusedPane: PanelTypeList,
	}
}

func (p *worklogPage) Init() tea.Cmd {
	shared.LogErrorf("WORKLOG_INIT", "Initializing worklog page")
	cfg := config.New()
	p.isConfigured = cfg.IsConfigured()
	shared.LogErrorf("WORKLOG_INIT", "Configuration status: %v", p.isConfigured)

	return tea.Batch(
	// TODO: Initialize your components
	)
}

func (p *worklogPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyboardEnhancementsMsg:
		p.keyboardEnhancements = msg
		return p, nil
	case tea.WindowSizeMsg:
		return p, tea.Batch(p.SetSize(msg.Width, msg.Height))
	case CancelTimerExpiredMsg:
		p.isCanceling = false
		return p, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, p.keyMap.Tab):
			p.changeFocus()
			return p, nil
		case key.Matches(msg, p.keyMap.Details):
			p.toggleDetails()
			return p, nil
		}

		switch p.focusedPane {
		case PanelTypeList:
			// TODO: Handle list navigation
		case PanelTypeDetail:
			// TODO: Handle detail view
		}
	}
	return p, tea.Batch(cmds...)
}

func (p *worklogPage) View() string {
	// TODO: Implement your layout similar to crush chat page
	// This should render your worklog list, detail view, etc.

	return "Worklog Page - TODO: Implement layout"
}

func (p *worklogPage) SetSize(width, height int) tea.Cmd {
	p.width = width
	p.height = height
	var cmds []tea.Cmd

	// TODO: Update component sizes

	return tea.Batch(cmds...)
}

func (p *worklogPage) changeFocus() {
	switch p.focusedPane {
	case PanelTypeList:
		p.focusedPane = PanelTypeDetail
	case PanelTypeDetail:
		p.focusedPane = PanelTypeList
	}
}

func (p *worklogPage) toggleDetails() {
	p.showingDetails = !p.showingDetails
}

func (p *worklogPage) Bindings() []key.Binding {
	bindings := []key.Binding{
		p.keyMap.Tab,
		p.keyMap.Details,
	}
	return bindings
}

func (p *worklogPage) Help() help.KeyMap {
	var shortList []key.Binding
	var fullList [][]key.Binding

	shortList = append(shortList,
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
		),
		key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "toggle details"),
		),
		key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	)

	for _, v := range shortList {
		fullList = append(fullList, []key.Binding{v})
	}

	return core.NewSimpleHelp(shortList, fullList)
}

func (p *worklogPage) IsListFocused() bool {
	return p.focusedPane == PanelTypeList
}
