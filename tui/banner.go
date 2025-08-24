// Package tui renders a LOGS wordmark in a stylized way.
package tui

import (
	"LogS/shared"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/lipgloss"
)

// letterform represents a letterform. It can be stretched horizontally by
// a given amount via the boolean argument.
type letterform func(bool) string

const diag = `╱`

// Opts are the options for rendering the LOGS title art.
type Opts struct {
	FieldColor   lipgloss.Color // diagonal lines
	TitleColorA  lipgloss.Color // left gradient ramp point
	TitleColorB  lipgloss.Color // right gradient ramp point
	CharmColor   lipgloss.Color // Charm™ text color
	VersionColor lipgloss.Color // Version text color
	Width        int            // width of the rendered logo, used for truncation
}

// RenderBanner renders the LOGS logo. Set the argument to true to render the narrow
// version, intended for use in a sidebar.
//
// The compact argument determines whether it renders compact for the sidebar
// or wider for the main pane.
func RenderBanner() string {
	// Use exact Crush colors
	o := Opts{
		FieldColor:   shared.Purple,
		TitleColorA:  shared.Pink,
		TitleColorB:  shared.Cyan,
		CharmColor:   shared.Purple,
		VersionColor: shared.Gray,
		Width:        0, // No truncation by default
	}

	return Render("v1.0.0", false, o)
}

// Render renders the LOGS logo with options.
func Render(version string, compact bool, o Opts) string {
	const charm = " Charm™"

	fg := func(c lipgloss.Color, s string) string {
		return lipgloss.NewStyle().Foreground(c).Render(s)
	}

	// Title.
	const spacing = 1
	letterforms := []letterform{
		letterL,
		letterO,
		letterG,
		letterSStylized,
	}
	stretchIndex := -1 // -1 means no stretching.
	if !compact {
		stretchIndex = rand.IntN(len(letterforms))
	}

	logs := renderWord(spacing, stretchIndex, letterforms...)
	logsWidth := lipgloss.Width(logs)
	b := new(strings.Builder)
	for r := range strings.SplitSeq(logs, "\n") {
		fmt.Fprintln(b, applyForegroundGrad(r, o.TitleColorA, o.TitleColorB))
	}
	logs = b.String()

	// Charm and version.
	metaRowGap := 1
	maxVersionWidth := logsWidth - lipgloss.Width(charm) - metaRowGap
	version = truncate(version, maxVersionWidth, "…") // truncate version if too long.
	gap := max(0, logsWidth-lipgloss.Width(charm)-lipgloss.Width(version))
	metaRow := fg(o.CharmColor, charm) + strings.Repeat(" ", gap) + fg(o.VersionColor, version)

	// Join the meta row and big LOGS title.
	logs = strings.TrimSpace(metaRow + "\n" + logs)

	// Narrow version.
	if compact {
		field := fg(o.FieldColor, strings.Repeat(diag, logsWidth))
		return strings.Join([]string{field, field, logs, field, ""}, "\n")
	}

	fieldHeight := lipgloss.Height(logs)

	// Left field.
	const leftWidth = 6
	leftFieldRow := fg(o.FieldColor, strings.Repeat(diag, leftWidth))
	leftField := new(strings.Builder)
	for range fieldHeight {
		fmt.Fprintln(leftField, leftFieldRow)
	}

	// Right field.
	rightWidth := max(15, o.Width-logsWidth-leftWidth-2) // 2 for the gap.
	const stepDownAt = 0
	rightField := new(strings.Builder)
	for i := range fieldHeight {
		width := rightWidth
		if i >= stepDownAt {
			width = rightWidth - (i - stepDownAt)
		}
		fmt.Fprint(rightField, fg(o.FieldColor, strings.Repeat(diag, width)), "\n")
	}

	// Return the wide version.
	const hGap = " "
	logo := lipgloss.JoinHorizontal(lipgloss.Top, leftField.String(), hGap, logs, hGap, rightField.String())
	if o.Width > 0 {
		// Truncate the logo to the specified width.
		lines := strings.Split(logo, "\n")
		for i, line := range lines {
			lines[i] = truncate(line, o.Width, "")
		}
		logo = strings.Join(lines, "\n")
	}
	return logo
}

// renderWord renders letterforms to fork a word. stretchIndex is the index of
// the letter to stretch, or -1 if no letter should be stretched.
func renderWord(spacing int, stretchIndex int, letterforms ...letterform) string {
	if spacing < 0 {
		spacing = 0
	}

	renderedLetterforms := make([]string, len(letterforms))

	// pick one letter randomly to stretch
	for i, letter := range letterforms {
		renderedLetterforms[i] = letter(i == stretchIndex)
	}

	if spacing > 0 {
		// Add spaces between the letters and render.
		renderedLetterforms = intersperse(renderedLetterforms, strings.Repeat(" ", spacing))
	}
	return strings.TrimSpace(
		lipgloss.JoinHorizontal(lipgloss.Top, renderedLetterforms...),
	)
}

// letterL renders the letter L in a stylized way.
func letterL(stretch bool) string {
	// Here's what we're making:
	//
	// █
	// █
	// ▀▀▀▀

	left := heredoc.Doc(`
		█
		█
		▀`)
	bottom := heredoc.Doc(`


		▀`)
	return joinLetterform(
		left,
		stretchLetterformPart(bottom, letterformProps{
			stretch:    stretch,
			width:      4,
			minStretch: 7,
			maxStretch: 12,
		}),
	)
}

// letterO renders the letter O in a stylized way.
func letterO(stretch bool) string {
	// Here's what we're making:
	//
	// ▄▀▀▀▄
	// █   █
	// ▀▀▀▀

	left := heredoc.Doc(`
		▄
		█
		▀`)
	middle := heredoc.Doc(`
		▀

		▀`)
	right := heredoc.Doc(`
		▄
		█
		▀`)
	return joinLetterform(
		left,
		stretchLetterformPart(middle, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
		right,
	)
}

// letterG renders the letter G in a stylized way.
func letterG(stretch bool) string {
	// Here's what we're making:
	//
	// ▄▀▀▀
	// █ ▀▀█
	// ▀▀▀▀

	left := heredoc.Doc(`
		▄
		█
		▀`)
	middle := heredoc.Doc(`
		▀
		 
		▀`)
	right := heredoc.Doc(`
		
		█
		▀`)
	return joinLetterform(
		left,
		stretchLetterformPart(middle, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
		right,
	)
}

// letterSStylized renders the letter S in a stylized way.
func letterSStylized(stretch bool) string {
	// Here's what we're making:
	//
	// ▄▀▀▀▀▀
	// ▀▀▀▀▀█
	// ▀▀▀▀▀

	left := heredoc.Doc(`
		▄
		▀
		▀`)
	center := heredoc.Doc(`
		▀
		▀
		▀`)
	right := heredoc.Doc(`
		▀
		█
		`)
	return joinLetterform(
		left,
		stretchLetterformPart(center, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
		right,
	)
}

func joinLetterform(letters ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, letters...)
}

// letterformProps defines letterform stretching properties.
type letterformProps struct {
	width      int
	minStretch int
	maxStretch int
	stretch    bool
}

// stretchLetterformPart is a helper function for letter stretching.
func stretchLetterformPart(s string, p letterformProps) string {
	if p.maxStretch < p.minStretch {
		p.minStretch, p.maxStretch = p.maxStretch, p.minStretch
	}
	n := p.width
	if p.stretch {
		n = rand.IntN(p.maxStretch-p.minStretch) + p.minStretch
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = s
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// Helper functions
func applyForegroundGrad(s string, colorA, colorB lipgloss.Color) string {
	// Simple gradient application
	styleA := lipgloss.NewStyle().Foreground(colorA)
	styleB := lipgloss.NewStyle().Foreground(colorB)

	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}

	result := ""
	for i, r := range runes {
		// Interpolate between colors
		if i < len(runes)/2 {
			result += styleA.Render(string(r))
		} else {
			result += styleB.Render(string(r))
		}
	}
	return result
}

func truncate(s string, width int, ellipsis string) string {
	if lipgloss.Width(s) <= width {
		return s
	}
	return s[:width-len(ellipsis)] + ellipsis
}

func intersperse(items []string, separator string) []string {
	if len(items) <= 1 {
		return items
	}

	result := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		result = append(result, item)
		if i < len(items)-1 {
			result = append(result, separator)
		}
	}
	return result
}

// Status rendering (keeping original functionality)
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
