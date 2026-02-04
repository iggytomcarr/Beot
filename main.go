package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"Beot/db"
	"Beot/ui"
)

// Version info - set via ldflags at build time
var (
	Version   = "0.1"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

func main() {
	// Handle version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("Beot %s (commit: %s, built: %s)\n", Version, CommitSHA, BuildDate)
		return
	}

	// Set version for UI
	ui.Version = Version

	// Connect to MongoDB
	if err := db.Connect(); err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Disconnect()

	p := tea.NewProgram(ui.NewAppModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
