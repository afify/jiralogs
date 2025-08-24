package styles

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

func NewCharmtoneTheme() *Theme {
	t := &Theme{
		Name:   "charmtone",
		IsDark: true,

		Primary:   charmtone.Charple,
		Secondary: charmtone.Dolly,
		Tertiary:  charmtone.Bok,
		Accent:    charmtone.Zest,

		// Backgrounds
		BgBase:        charmtone.Pepper,
		BgBaseLighter: charmtone.BBQ,
		BgSubtle:      charmtone.Charcoal,
		BgOverlay:     charmtone.Iron,

		// Foregrounds
		FgBase:      charmtone.Ash,
		FgMuted:     charmtone.Squid,
		FgHalfMuted: charmtone.Smoke,
		FgSubtle:    charmtone.Oyster,
		FgSelected:  charmtone.Salt,

		// Borders
		Border:      charmtone.Charcoal,
		BorderFocus: charmtone.Charple,

		// Status
		Success: charmtone.Guac,
		Error:   charmtone.Sriracha,
		Warning: charmtone.Zest,
		Info:    charmtone.Malibu,

		// Colors
		White: charmtone.Butter,

		BlueLight: charmtone.Sardine,
		Blue:      charmtone.Malibu,

		Yellow: charmtone.Mustard,
		Citron: charmtone.Citron,

		Green:      charmtone.Julep,
		GreenDark:  charmtone.Guac,
		GreenLight: charmtone.Bok,

		Red:      charmtone.Coral,
		RedDark:  charmtone.Sriracha,
		RedLight: charmtone.Salmon,
		Cherry:   charmtone.Cherry,
	}

	// Text selection.
	t.TextSelection = lipgloss.NewStyle().Foreground(t.FgSelected).Background(t.Primary)

	// LSP and MCP status.
	t.ItemOfflineIcon = lipgloss.NewStyle().Foreground(t.FgMuted).SetString("●")
	t.ItemBusyIcon = t.ItemOfflineIcon.Foreground(t.Warning)
	t.ItemErrorIcon = t.ItemOfflineIcon.Foreground(t.Error)
	t.ItemOnlineIcon = t.ItemOfflineIcon.Foreground(t.Success)

	t.YoloIconFocused = lipgloss.NewStyle().Foreground(t.FgSubtle).Background(t.Warning).Bold(true).SetString(" ! ")
	t.YoloIconBlurred = t.YoloIconFocused.Foreground(t.BgBase).Background(t.FgMuted)
	t.YoloDotsFocused = lipgloss.NewStyle().Foreground(t.Accent).SetString(":::")
	t.YoloDotsBlurred = t.YoloDotsFocused.Foreground(t.FgMuted)

	return t
}
