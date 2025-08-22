package tui

import "github.com/charmbracelet/lipgloss"

func RenderBanner() string {
	logPart := BannerStyle.Render(`
██╗      ██████╗  ██████╗ `)
	sPart := BannerSStyle.Render(`███████╗`)
	line1 := logPart + sPart

	logPart2 := BannerStyle.Render(`██║     ██╔═══██╗██╔════╝ `)
	sPart2 := BannerSStyle.Render(`██╔════╝`)
	line2 := logPart2 + sPart2

	logPart3 := BannerStyle.Render(`██║     ██║   ██║██║  ███╗`)
	sPart3 := BannerSStyle.Render(`███████╗`)
	line3 := logPart3 + sPart3

	logPart4 := BannerStyle.Render(`██║     ██║   ██║██║   ██║`)
	sPart4 := BannerSStyle.Render(`╚════██║`)
	line4 := logPart4 + sPart4

	logPart5 := BannerStyle.Render(`███████╗╚██████╔╝╚██████╔╝`)
	sPart5 := BannerSStyle.Render(`███████║`)
	line5 := logPart5 + sPart5

	logPart6 := BannerStyle.Render(`╚══════╝ ╚═════╝  ╚═════╝ `)
	sPart6 := BannerSStyle.Render(`╚══════╝`)
	line6 := logPart6 + sPart6

	subtitle := BannerSubtitle.Render("Made with LOVE ❤️")

	centeredSubtitle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(subtitle)

	return lipgloss.JoinVertical(lipgloss.Left,
		line1,
		line2,
		line3,
		line4,
		line5,
		line6,
		"",
		centeredSubtitle,
	)
}

func RenderStatus(status string) string {
	switch status {
	case "success":
		return SuccessStyle.Render("✓")
	case "error":
		return ErrorStyle.Render("✗")
	case "warning":
		return WarningStyle.Render("⚠")
	case "info":
		return InfoStyle.Render("ℹ")
	default:
		return status
	}
}
