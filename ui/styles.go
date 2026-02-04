package ui

import "github.com/charmbracelet/lipgloss"

var (
	Primary   = lipgloss.Color("#E6DCC7") // Parchment
	Secondary = lipgloss.Color("#A9A393") // Ash
	Muted     = lipgloss.Color("#7C776C") // Muted/Helper
	Gold      = lipgloss.Color("#DAA520") // Anglo-Saxon Gold
	Success   = lipgloss.Color("82")      // Green
	Warning   = lipgloss.Color("214")     // Orange
	Danger    = lipgloss.Color("196")     // Red
)

// Text styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	NormalStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger).
			Bold(true)

	StreakStyle = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true)

	VersionStyle = lipgloss.NewStyle().
			Foreground(Muted)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)
)

// Layout styles
var (
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	CenteredStyle = lipgloss.NewStyle().
			Align(lipgloss.Center)

	TimerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	StatusStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	// IconStyle ensures all icons take up the same width
	IconStyle = lipgloss.NewStyle().Width(3)
)

// QuoteStyle for displaying motivational quotes
var QuoteStyle = lipgloss.NewStyle().
	Foreground(Secondary).
	Italic(true).
	Width(70).
	MarginLeft(4)

// OldEnglishStyle for Old English text - golden/amber color
var OldEnglishStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#DAA520")). // Gold
	Italic(true).
	Width(70).
	MarginLeft(4)

// ModernEnglishStyle for modern translation
var ModernEnglishStyle = lipgloss.NewStyle().
	Foreground(Secondary).
	Width(70).
	MarginLeft(4)

// RenderHeader renders just the Bēot title
func RenderHeader() string {
	return TitleStyle.Render("Bēot")
}

// RenderQuote renders a quote with optional source
func RenderQuote(text, source string) string {
	quote := QuoteStyle.Render("\"" + text + "\"")
	if source != "" {
		quote += "\n    " + HelpStyle.Render("— "+source)
	}
	return quote
}

// RenderPoem renders a poem with Old English and Modern English side by side
func RenderPoem(oldEnglish, modernEnglish, source, lineRef string) string {
	oe := OldEnglishStyle.Render(oldEnglish)
	me := ModernEnglishStyle.Render(modernEnglish)

	attribution := source
	if lineRef != "" {
		attribution += ", " + lineRef
	}

	return oe + "\n\n" + me + "\n    " + HelpStyle.Render("— "+attribution)
}
