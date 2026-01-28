package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define styles as package-level variables
var (
	boxStyle             = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("136")).Padding(1, 2) // Dark gold, like mead
	titleStyle           = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("178")).MarginBottom(1)                           // Gold, like Anglo-Saxon metalwork
	positiveCounterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("108")).Bold(true)                                           // Muted green, like the English countryside
	negativeCounterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("131")).Bold(true)                                           // Earthy rust red
	helpStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("101")).MarginTop(1)                                         // Weathered stone
)

// Model holds ALL application state
type model struct {
	counter  int
	showHelp bool
}

// Init is called once when the program starts
// Return nil if you don't need to run any initial commands
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Handle keyboard input here
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "h":
			m.showHelp = !m.showHelp
		case "up", "k":
			m.counter++
		case "down", "j":
			m.counter--
		case "d":
			m.counter = m.counter + m.counter
		case "r":
			m.counter = 0
		}

	}
	return m, nil
}

func (m model) View() string {

	title := titleStyle.Render("Bēot")

	var counter string

	if m.counter >= 0 {
		counter = positiveCounterStyle.Render(fmt.Sprintf("%d", m.counter))
	} else {
		counter = negativeCounterStyle.Render(fmt.Sprintf("%d", m.counter))
	}

	var help string

	if m.showHelp {
		help = helpStyle.Render("↑/k increase • ↓/j decrease • r reset • q quit • d double • h toggle help")
	} else {
		help = ""
	}

	content := fmt.Sprintf("\n  %s\n\n Count: %s\n\n  %s\n", title, counter, help)

	return "\n" + boxStyle.Render(content) + "\n"

}

func main() {

	// Create and run the program
	p := tea.NewProgram(model{counter: 0, showHelp: true}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error :%v", err)
		os.Exit(1)
	}

}
