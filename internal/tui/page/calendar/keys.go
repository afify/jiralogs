package calendar

import (
	"github.com/charmbracelet/bubbles/v2/key"
)

type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	PrevYear   key.Binding
	NextYear   key.Binding
	Today      key.Binding
	Back       key.Binding
	Quit       key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("→/l", "right"),
		),
		PrevYear: key.NewBinding(
			key.WithKeys("[", "shift+left"),
			key.WithHelp("[", "prev year"),
		),
		NextYear: key.NewBinding(
			key.WithKeys("]", "shift+right"),
			key.WithHelp("]", "next year"),
		),
		Today: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "go to today"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "b"),
			key.WithHelp("esc/b", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}