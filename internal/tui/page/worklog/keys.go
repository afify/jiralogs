package worklog

import (
	"github.com/charmbracelet/bubbles/v2/key"
)

type KeyMap struct {
	NewWorklog    key.Binding
	LogTime       key.Binding
	Cancel        key.Binding
	Tab           key.Binding
	Details       key.Binding
	Refresh       key.Binding
	FilterPeriod  key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		NewWorklog: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "new worklog"),
		),
		LogTime: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "log time"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "change focus"),
		),
		Details: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "toggle details"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "F5"),
			key.WithHelp("r/F5", "refresh"),
		),
		FilterPeriod: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter period"),
		),
	}
}
