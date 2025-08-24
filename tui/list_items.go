package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea/v2"
)

type menuItem struct {
	title string
	desc  string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

type menuItemDelegate struct{}

func (d menuItemDelegate) Height() int                             { return 2 }
func (d menuItemDelegate) Spacing() int                            { return 1 }
func (d menuItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d menuItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(menuItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s\n%s", i.title, DescStyle.Render(i.desc))

	if index == m.Index() {
		str = SelectedItemStyle.Render("▶ " + str)
	} else {
		str = ListItemStyle.Render(str)
	}

	_, _ = fmt.Fprint(w, str)
}

type periodItem struct {
	id    string
	title string
	desc  string
}

func (i periodItem) Title() string       { return i.title }
func (i periodItem) Description() string { return i.desc }
func (i periodItem) FilterValue() string { return i.title }

type periodItemDelegate struct{}

func (d periodItemDelegate) Height() int                             { return 2 }
func (d periodItemDelegate) Spacing() int                            { return 1 }
func (d periodItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d periodItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(periodItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s\n%s", i.title, DescStyle.Render(i.desc))

	if index == m.Index() {
		str = SelectedItemStyle.Render("▶ " + str)
	} else {
		str = ListItemStyle.Render(str)
	}

	_, _ = fmt.Fprint(w, str)
}
