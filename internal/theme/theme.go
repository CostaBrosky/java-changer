package theme

import "github.com/charmbracelet/lipgloss"

// JV Theme - Custom color palette inspired by Java
var (
	// Primary colors - Java-inspired orange/red
	Primary   = lipgloss.Color("#f89820") // Java orange
	Secondary = lipgloss.Color("#5382a1") // Java blue
	Accent    = lipgloss.Color("#e76f00") // Dark orange

	// Semantic colors
	Success = lipgloss.Color("#00d26a") // Green
	Error   = lipgloss.Color("#ff3b30") // Red
	Warning = lipgloss.Color("#ffcc00") // Yellow
	Info    = lipgloss.Color("#5ac8fa") // Light blue

	// UI colors
	Text      = lipgloss.Color("#ffffff") // White
	TextFaint = lipgloss.Color("#8e8e93") // Gray
	Border    = lipgloss.Color("#5382a1") // Java blue

	// Specific shades
	Highlight = lipgloss.Color("#ff6b35") // Bright orange
	Muted     = lipgloss.Color("#636366") // Dark gray
)

// Styles - Pre-configured styles for common use cases
var (
	// Title styles
	Title = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		Underline(true)

	Subtitle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	// Message styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Info)

	// Text styles
	Bold = lipgloss.NewStyle().
		Bold(true)

	Faint = lipgloss.NewStyle().
		Foreground(TextFaint).
		Faint(true)

	Code = lipgloss.NewStyle().
		Foreground(Highlight)

	// Interactive element styles
	CurrentStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	ValueStyle = lipgloss.NewStyle().
			Foreground(Text)

	PathStyle = lipgloss.NewStyle().
			Foreground(Info)

	// Box styles
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(1, 2)

	SuccessBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Success).
			Padding(1, 3).
			Align(lipgloss.Center)

	ErrorBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Error).
			Padding(1, 2)

	WarningBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Warning).
			Padding(1, 2)

	InfoBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Info).
		Padding(1, 2)

	// Special box with double border for titles
	TitleBox = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(Primary).
			Padding(1, 2).
			Align(lipgloss.Center)

	// Table styles
	TableStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(Border)

	TableHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Padding(0, 1)

	TableCell = lipgloss.NewStyle().
			Padding(0, 1)

	// Command/step styles
	StepStyle = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	CommandStyle = lipgloss.NewStyle().
			Foreground(Success)

	// Banner style (for ASCII art)
	Banner = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true)
)

// Helper functions for common patterns

// SuccessMessage returns a formatted success message
func SuccessMessage(msg string) string {
	return SuccessStyle.Render("✓ " + msg)
}

// ErrorMessage returns a formatted error message
func ErrorMessage(msg string) string {
	return ErrorStyle.Render("✗ " + msg)
}

// WarningMessage returns a formatted warning message
func WarningMessage(msg string) string {
	return WarningStyle.Render("⚠ " + msg)
}

// InfoMessage returns a formatted info message
func InfoMessage(msg string) string {
	return InfoStyle.Render("ℹ " + msg)
}

// Highlight returns text with highlight color
func HighlightText(text string) string {
	return lipgloss.NewStyle().Foreground(Highlight).Render(text)
}
