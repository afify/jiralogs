package core

import (
	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
)

type KeyMapHelp interface {
	Help() help.KeyMap
}

type simpleHelp struct {
	shortList []key.Binding
	fullList  [][]key.Binding
}

func NewSimpleHelp(shortList []key.Binding, fullList [][]key.Binding) help.KeyMap {
	return &simpleHelp{
		shortList: shortList,
		fullList:  fullList,
	}
}

// FullHelp implements help.KeyMap.
func (s *simpleHelp) FullHelp() [][]key.Binding {
	return s.fullList
}

// ShortHelp implements help.KeyMap.
func (s *simpleHelp) ShortHelp() []key.Binding {
	return s.shortList
}