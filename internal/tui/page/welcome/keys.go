package welcome

import "github.com/charmbracelet/bubbles/v2/key"

// KeyMap defines the key bindings for the welcome page.
type KeyMap struct {
	Continue key.Binding
}

// DefaultKeyMap returns the default key mappings for the welcome page.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Continue: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "continue to worklog"),
		),
	}
}
