package shared

import "github.com/charmbracelet/lipgloss"

// Colors used throughout the application - Modern palette inspired by Bubble Tea & Lipgloss
var (
	// Primary colors - Vibrant and modern
	DarkBlue  = lipgloss.Color("24")      // Keep original banner blue (terminal color for ASCII art)
	LightBlue = lipgloss.Color("#73F59F") // Mint green-blue
	Orange    = lipgloss.Color("#F77F00") // Vivid orange
	Cyan      = lipgloss.Color("#14F9D5") // Electric cyan
	Green     = lipgloss.Color("#43BF6D") // Fresh green
	Yellow    = lipgloss.Color("#FDFF90") // Soft lemon
	Red       = lipgloss.Color("#FF7698") // Coral red
	Purple    = lipgloss.Color("#7D56F4") // Royal purple (from Lipgloss)
	Pink      = lipgloss.Color("#F25D94") // Bubblegum pink

	// Grays and neutrals - Refined palette
	Gray      = lipgloss.Color("#969B86") // Warm gray (from Lipgloss)
	DarkGray  = lipgloss.Color("#383838") // Charcoal
	LightGray = lipgloss.Color("#D9DCCF") // Pearl gray
	White     = lipgloss.Color("#FAFAFA") // Soft white

	// Background colors - Rich and deep
	BgDark   = lipgloss.Color("#0F0F0F") // Almost black
	BgLight  = lipgloss.Color("#1A1B3A") // Midnight blue
	BgAccent = lipgloss.Color("#16213E") // Navy accent

	// Special purpose colors
	SuccessColor = lipgloss.Color("#43BF6D") // Green for success
	ErrorColor   = lipgloss.Color("#FF5F87") // Coral for errors
	WarningColor = lipgloss.Color("#EDFF82") // Lime for warnings
	InfoColor    = lipgloss.Color("#14F9D5") // Cyan for info

	// Popup and dialog colors
	PrimaryColor = lipgloss.Color("#874BFD") // Bright purple
	DialogBg     = lipgloss.Color("#353533") // Dark dialog background
	BrightText   = lipgloss.Color("#FFF7DB") // Cream text
	SubduedText  = lipgloss.Color("#C1C6B2") // Muted sage
	DimText      = lipgloss.Color("#696969") // Dim gray
	OverlayShade = lipgloss.Color("#1A1A2E") // Dark overlay

	// Main gradient colors - Beautiful combinations
	GradientStart = lipgloss.Color("#F25D94") // Coral pink
	GradientMid   = lipgloss.Color("#EDFF82") // Lime yellow
	GradientEnd   = lipgloss.Color("#73F59F") // Mint green

	// Card gradient sets for stats page
	Card1GradientStart = lipgloss.Color("#874BFD") // Purple
	Card1GradientMid   = lipgloss.Color("#A550DF") // Light purple
	Card1GradientEnd   = lipgloss.Color("#6124DF") // Deep purple

	Card2GradientStart = lipgloss.Color("#43BF6D") // Green
	Card2GradientMid   = lipgloss.Color("#52B788") // Sea green
	Card2GradientEnd   = lipgloss.Color("#2D6A4F") // Forest green

	Card3GradientStart = lipgloss.Color("#14F9D5") // Cyan
	Card3GradientMid   = lipgloss.Color("#00B4D8") // Sky blue
	Card3GradientEnd   = lipgloss.Color("#0077B6") // Ocean blue

	Card4GradientStart = lipgloss.Color("#F77F00") // Orange
	Card4GradientMid   = lipgloss.Color("#FCBF49") // Gold
	Card4GradientEnd   = lipgloss.Color("#EAE2B7") // Cream

	// Additional gradient sets
	WarmGradientStart = lipgloss.Color("#FF006E") // Hot pink
	WarmGradientMid   = lipgloss.Color("#FB5607") // Orange red
	WarmGradientEnd   = lipgloss.Color("#FFBE0B") // Golden

	CoolGradientStart = lipgloss.Color("#643AFF") // Indigo
	CoolGradientMid   = lipgloss.Color("#5A189A") // Deep purple
	CoolGradientEnd   = lipgloss.Color("#9D4EDD") // Violet

	// Button gradients (for compatibility)
	ButtonGradientStart = lipgloss.Color("#667EEA") // Indigo gradient start
	ButtonGradientEnd   = lipgloss.Color("#764BA2") // Purple gradient end

	// Status gradients (for compatibility)
	SuccessGradientStart = lipgloss.Color("#43BF6D") // Green gradient start
	SuccessGradientEnd   = lipgloss.Color("#2D6A4F") // Darker green end

	ErrorGradientStart = lipgloss.Color("#FF5F87") // Pink-red gradient start
	ErrorGradientEnd   = lipgloss.Color("#E63946") // Red gradient end

	WarningGradientStart = lipgloss.Color("#FDFF90") // Amber gradient start
	WarningGradientEnd   = lipgloss.Color("#F77F00") // Orange gradient end

	InfoGradientStart = lipgloss.Color("#14F9D5") // Cyan gradient start
	InfoGradientEnd   = lipgloss.Color("#0077B6") // Darker cyan end

	Reset = ""
)
