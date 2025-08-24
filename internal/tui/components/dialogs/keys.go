package dialogs

import (
	"github.com/charmbracelet/bubbles/v2/key"
)

type KeyMap struct {
	Close key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Close: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
	}
}
