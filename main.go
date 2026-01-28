package main

import (
	"fmt"
	"os"
	"time"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

// Define styles as package-level variables
var (
	timerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("101")).MarginTop(1) // Weathered stone
)

// Model holds ALL application state
type model struct {
	totalSeconds     int
	remainingSeconds int
	running          bool
}

func newModel(minutes int) model {
	seconds := minutes * 60

	return model{totalSeconds: seconds, remainingSeconds: seconds, running: false}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})

}

// Init is called once when the program starts
// Return nil if you don't need to run any initial commands
func (m model) Init() tea.Cmd {

	// Start the timer
	m.running = true
	return tickCmd()

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
				return m, tickCmd()
			}
		}
	}
	return m, nil
}

func (m model) View() string {
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

	// Help
	help := helpStyle.Render("Spacebar to pause/resume • r reset • q quit • h toggle help")

	return fmt.Sprintf("\n %s\n\n  %s\n\n %s\n", status, timeDisplay, help)

}

func main() {

	// Create and run the program
	p := tea.NewProgram(newModel(1))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error :%v", err)
		os.Exit(1)
	}

}
