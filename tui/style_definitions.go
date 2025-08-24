package tui

import (
	"LogS/shared"
	"github.com/charmbracelet/lipgloss"
)

var (
	BannerStyle = lipgloss.NewStyle().
			Foreground(shared.DarkBlue).
			Bold(true)

	BannerSStyle = lipgloss.NewStyle().
			Foreground(shared.Orange).
			Bold(true)

	BannerSubtitle = lipgloss.NewStyle().
			Foreground(shared.Yellow).
			Italic(true)

	AppStyle = lipgloss.NewStyle().
			Padding(0)

	HeaderStyle = lipgloss.NewStyle().
			Background(shared.BgDark).
			Foreground(shared.Cyan).
			Bold(true).
			Padding(0, 2).
			MarginBottom(1)

	ContentStyle = lipgloss.NewStyle().
			Margin(1, 2)

	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(shared.Gray).
			Padding(1, 2).
			Align(lipgloss.Center)

	ListTitleStyle = lipgloss.NewStyle().
			Foreground(shared.Cyan).
			Bold(true).
			Padding(0, 0, 1, 0)

	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(shared.Orange).
				Bold(true).
				PaddingLeft(1)

	InputStyle = lipgloss.NewStyle().
			Foreground(shared.White).
			Background(shared.BgLight).
			Padding(0, 1)

	LabelStyle = lipgloss.NewStyle().
			Foreground(shared.Cyan).
			MarginBottom(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(shared.Green).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(shared.Red).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(shared.Yellow).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(shared.Cyan)

	ProgressFullStyle = lipgloss.NewStyle().
				Foreground(shared.Green)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(shared.Gray)

	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(shared.Cyan).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(shared.Gray)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(shared.Gray).
			Padding(1, 2).
			MarginBottom(1)

	HelpStyle = lipgloss.NewStyle().
			Background(shared.BgLight).
			Padding(0, 1)

	KeyStyle = lipgloss.NewStyle().
			Foreground(shared.Yellow).
			Bold(true)

	DescStyle = lipgloss.NewStyle().
			Foreground(shared.Gray)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(shared.Cyan)

	BadgeCompliant = lipgloss.NewStyle().
			Background(shared.Green).
			Foreground(shared.White).
			Bold(true).
			Padding(0, 1)

	BadgePartial = lipgloss.NewStyle().
			Background(shared.Yellow).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	BadgeMissing = lipgloss.NewStyle().
			Background(shared.Red).
			Foreground(shared.White).
			Bold(true).
			Padding(0, 1)

	BadgeWeekend = lipgloss.NewStyle().
			Background(lipgloss.Color("#0087FF")).
			Foreground(shared.White).
			Bold(true).
			Padding(0, 1)

	FrameStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(shared.Purple).
			Background(lipgloss.Color("#0F0F0F")). // Clean dark background like Crush
			Padding(1)

	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(shared.Cyan)

	ProgressTextStyle = lipgloss.NewStyle().
				Foreground(shared.Yellow).
				Italic(true)
)
