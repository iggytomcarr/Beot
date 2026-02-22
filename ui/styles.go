package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

// RenderHeader renders just the Bēot title (compact, for timer etc.)
func RenderHeader() string {
	return TitleStyle.Render("Bēot")
}

// bannerLines contains the ASCII art for bēot (lowercase) with macron above ē
var bannerLines = []string{
	"           ▄▄▄▄",
	" ██                           ██",
	" █████▄    ▄██▄     ▄██▄    ██████",
	" ██  ██   ██  ██   ██  ██     ██",
	" ██  ██   ██████   ██  ██     ██",
	" ██  ██   ██       ██  ██     ██",
	" █████▀    ▀██▀     ▀██▀     ▀██",
}

type rgb struct{ r, g, b uint8 }

var bannerGradient = []rgb{
	{0x7E, 0xB8, 0xDA}, // Steel blue
	{0x9B, 0x7E, 0xC8}, // Amethyst
	{0xDA, 0xA5, 0x20}, // Anglo-Saxon gold
}

func lerpRGB(a, b rgb, t float64) rgb {
	return rgb{
		r: uint8(float64(a.r) + t*(float64(b.r)-float64(a.r))),
		g: uint8(float64(a.g) + t*(float64(b.g)-float64(a.g))),
		b: uint8(float64(a.b) + t*(float64(b.b)-float64(a.b))),
	}
}

func gradientAt(pos, total int) lipgloss.Color {
	if total <= 1 {
		return lipgloss.Color("#daa520")
	}
	t := float64(pos) / float64(total-1)

	segs := len(bannerGradient) - 1
	seg := int(t * float64(segs))
	if seg >= segs {
		seg = segs - 1
	}
	lt := t*float64(segs) - float64(seg)

	c := lerpRGB(bannerGradient[seg], bannerGradient[seg+1], lt)
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", c.r, c.g, c.b))
}

// RenderBanner renders the large ASCII art BĒOT title with gradient
func RenderBanner() string {
	maxW := 0
	for _, line := range bannerLines {
		if w := len([]rune(line)); w > maxW {
			maxW = w
		}
	}

	var b strings.Builder
	for i, line := range bannerLines {
		for j, ch := range []rune(line) {
			if ch == ' ' {
				b.WriteRune(' ')
			} else {
				style := lipgloss.NewStyle().Foreground(gradientAt(j, maxW)).Bold(true)
				b.WriteString(style.Render(string(ch)))
			}
		}
		if i < len(bannerLines)-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
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
