package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

// Define styles as package-level variables
var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#C9A84C")).MarginBottom(1)
	goldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#C9A84C")) // Gold lattice
	rubyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B2500")) // Garnet inlay
	blueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E3A5F")) // Millefiori glass
	timerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#C9A84C")).MarginBottom(1)
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B4513"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7A6B5D")).MarginTop(1)
	vowStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B4513")).Italic(true)
)

func renderHeader() string {
	g := goldStyle.Render
	r := rubyStyle.Render
	b := blueStyle.Render

	// Gold border with garnet diamond lattice and blue millefiori inlays
	// Each inner row is exactly 29 visible characters wide (matching the border)
	// All inner content is exactly 28 chars wide
	top := g("╔════════════════════════════╗")
	r1 := g("║") + r("/") + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + g("║")
	r2 := g("║") + r(`\/\/\/\/\/\/\/\/\/\/\/\/\/\/`) + g("║")
	r3 := g("║") + r(`/\/\/\/\/\/\/\/\/\/\/\/\/\/\`) + g("║")
	r4 := g("║") + r(`\`) + b("o") + r(`/\`) + b("o") + r(`/\`) + b("o") + r(`/\`) + b("o") + r(`/\`) + b("o") + r(`/\`) + b("o") + r(`/\`) + b("o") + r(`/\`) + b("o") + r(`/\`) + b("o") + r(`/\`) + g("║")
	r5 := g("║") + r(`/\/\/\/\/\/\/\/\/\/\/\/\/\/\`) + g("║")
	r6 := g("║") + r(`\/\/\/\/\/\/\/\/\/\/\/\/\/\/`) + g("║")
	r7 := g("║") + r("/") + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + b("o") + r(`\/`) + g("║")
	bot := g("╚════════════════════════════╝")

	clasp := top + "\n" + r1 + "\n" + r2 + "\n" + r3 + "\n" + r4 + "\n" + r5 + "\n" + r6 + "\n" + r7 + "\n" + bot

	title := titleStyle.Render("Bēot")

	vow := vowStyle.Render("Hark! Hear me, hearth-kin and war-fellows!\n" +
		"The fire is high, the mead is spent, and the night has teeth.\n" +
		"\n" +
		"Now is no hour for quiet men.\n" +
		"Now is the time to rise, to stand tall beneath the roof-beams,\n" +
		"to let word and will be one.\n" +
		"\n" +
		"I speak not of what I have done —\n" +
		"I speak of what I shall do.\n" +
		"By blade and breath, by bone and blood,\n" +
		"I bind my honour to my word.\n" +
		"\n" +
		"Let the gods bear witness.\n" +
		"Let the ancestors listen from the dark.\n" +
		"Let wyrd itself turn its face toward me.\n" +
		"\n" +
		"This is my Bēot.\n" +
		"If I fail it, let my name be broken.\n" +
		"If I keep it, let it live after me.")

	right := title + "\n\n" + vow

	return lipgloss.JoinHorizontal(lipgloss.Top, clasp, "   ", right)
}

// Model holds ALL application state
type model struct {
	totalSeconds     int
	remainingSeconds int
	running          bool
	progress         progress.Model
}

func newModel(minutes int) model {
	seconds := minutes * 60

	prog := progress.New(progress.WithGradient("#4A3728", "#C9A84C"))
	prog.Width = 80

	return model{totalSeconds: seconds, remainingSeconds: seconds, running: false, progress: prog}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})

}

// Init is called once when the program starts
// Return nil if you don't need to run any initial commands
func (m model) Init() tea.Cmd {
	m.running = true
	return tea.Batch(tea.ClearScreen, tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Handle keyboard input here
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ": // Spacebar to pause/resume
			m.running = !m.running
			if m.running {
				return m, tickCmd()
			}
		case "r": // Reset
			m.remainingSeconds = m.totalSeconds
			m.running = true
			return m, tickCmd()
		}
	case tickMsg:
		if m.running && m.remainingSeconds > 0 {
			m.remainingSeconds--
			if m.remainingSeconds <= 0 {
				m.running = false
				// Timer complete!
				return m, nil
			}
			return m, tickCmd()
		}
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	// Calculate progress (0.0 to 1.0)
	elasped := m.totalSeconds - m.remainingSeconds
	percent := float64(elasped) / float64(m.totalSeconds)

	// Format time as MM:SS
	minutes := m.remainingSeconds / 60
	seconds := m.remainingSeconds % 60
	timeDisplay := timerStyle.Render(fmt.Sprintf("%02d:%02d", minutes, seconds))

	// Status
	status := statusStyle.Render("Focus Time")
	if !m.running && m.remainingSeconds > 0 {
		status = statusStyle.Render("Paused")
	} else if m.remainingSeconds <= 0 {
		status = statusStyle.Render("Complete!")
	}

	progressBar := m.progress.ViewAs(percent)

	// Help
	help := helpStyle.Render("Spacebar to pause/resume • r reset • q quit • h toggle help")

	header := renderHeader()

	return fmt.Sprintf(
		"\n%s\n\n %s\n\n %s\n\n %s  %s\n\n %s\n",
		header,
		status,
		progressBar,
		timeDisplay,
		helpStyle.Render(fmt.Sprintf("(%d%% complete)", int(percent*100))),
		help,
	)

}

func main() {

	// Create and run the program
	p := tea.NewProgram(newModel(1))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error :%v", err)
		os.Exit(1)
	}

}
