package stats

import "github.com/charmbracelet/bubbles/v2/key"

// KeyMap defines the key bindings for the stats page.
type KeyMap struct {
	LogTime     key.Binding
	Refresh     key.Binding
	ListTickets key.Binding
	Calendar    key.Binding
	Quit        key.Binding
}

// DefaultKeyMap returns the default key mappings for the stats page.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		LogTime: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "log time"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh data"),
		),
		ListTickets: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "list tickets"),
		),
		Calendar: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "calendar view"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c/q", "quit"),
		),
	}
}
