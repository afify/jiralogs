package tui

import "github.com/charmbracelet/lipgloss"

var (
	BannerStyle = lipgloss.NewStyle().
			Foreground(DarkBlue).
			Bold(true)

	BannerSStyle = lipgloss.NewStyle().
			Foreground(Orange).
			Bold(true)

	BannerSubtitle = lipgloss.NewStyle().
			Foreground(Yellow).
			Italic(true)

	AppStyle = lipgloss.NewStyle().
			Padding(0)

	HeaderStyle = lipgloss.NewStyle().
			Background(BgDark).
			Foreground(Cyan).
			Bold(true).
			Padding(0, 2).
			MarginBottom(1)

	ContentStyle = lipgloss.NewStyle().
			Margin(1, 2)

	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Gray).
			Padding(1, 2).
			Align(lipgloss.Center)

	ListTitleStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true).
			Padding(0, 0, 1, 0)

	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(Orange).
				Bold(true).
				PaddingLeft(1)

	InputStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(BgLight).
			Padding(0, 1)

	LabelStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			MarginBottom(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Yellow).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Cyan)

	ProgressFullStyle = lipgloss.NewStyle().
				Foreground(Green)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(Gray)

	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(Cyan).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(Gray)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	BoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Gray).
			Padding(1, 2).
			MarginBottom(1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Italic(true)

	KeyStyle = lipgloss.NewStyle().
			Foreground(Yellow).
			Bold(true)

	DescStyle = lipgloss.NewStyle().
			Foreground(Gray)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Cyan)

	BadgeCompliant = lipgloss.NewStyle().
			Background(Green).
			Foreground(White).
			Bold(true).
			Padding(0, 1)

	BadgePartial = lipgloss.NewStyle().
			Background(Yellow).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1)

	BadgeMissing = lipgloss.NewStyle().
			Background(Red).
			Foreground(White).
			Bold(true).
			Padding(0, 1)

	BadgeWeekend = lipgloss.NewStyle().
			Background(lipgloss.Color("33")).
			Foreground(White).
			Bold(true).
			Padding(0, 1)

	FrameStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(DarkBlue).
			Padding(1)

	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(Cyan)

	ProgressTextStyle = lipgloss.NewStyle().
				Foreground(Yellow).
				Italic(true)
)
