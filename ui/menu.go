package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// MenuChoice represents the menu options
type MenuChoice int

const (
	StartSession MenuChoice = iota
	ViewStats
	ManageQuotes
	ToggleDisplayMode
	QuitApp
)

type menuItem struct {
	icon string
	text string
}

// Version is set from main.go
var Version = "dev"

// MenuModel handles the main menu
type MenuModel struct {
	choices     []menuItem
	cursor      int
	streak      int         // We'll populate this later from the database
	displayMode DisplayMode // Current display mode for timer
}

// NewMenuModel creates a new menu
func NewMenuModel() MenuModel {
	return MenuModel{
		choices: []menuItem{
			{icon: "🎯", text: "Start Focus Session"},
			{icon: "📜", text: "View Statistics"},
			{icon: "💬", text: "Manage Quotes"},
			{icon: "📖", text: "Display: Quotes"},
			{icon: "🚪", text: "Quit"},
		},
		cursor:      0,
		displayMode: DisplayModeQuotes,
	}
}

// GetDisplayMode returns the current display mode
func (m MenuModel) GetDisplayMode() DisplayMode {
	return m.displayMode
}

// updateDisplayModeText updates the menu item text for display mode
func (m *MenuModel) updateDisplayModeText() {
	if m.displayMode == DisplayModePoems {
		m.choices[3] = menuItem{icon: "📖", text: "Display: Old English Poems"}
	} else {
		m.choices[3] = menuItem{icon: "💬", text: "Display: Quotes"}
	}
}

// SetStreak updates the streak display
func (m *MenuModel) SetStreak(s int) {
	m.streak = s
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			// Handle display mode toggle locally
			if MenuChoice(m.cursor) == ToggleDisplayMode {
				if m.displayMode == DisplayModeQuotes {
					m.displayMode = DisplayModePoems
				} else {
					m.displayMode = DisplayModeQuotes
				}
				m.updateDisplayModeText()
				return m, nil
			}
			// Send a message about what was selected
			return m, func() tea.Msg {
				return MenuSelectionMsg(m.cursor)
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	// Title banner and version
	title := RenderBanner()
	version := VersionStyle.Render("v" + Version)

	// Menu items
	var items string
	for i, choice := range m.choices {
		cursor := "  "
		style := NormalStyle

		if m.cursor == i {
			cursor = "▸ "
			style = SelectedStyle
		}

		icon := IconStyle.Render(choice.icon)
		items += fmt.Sprintf("%s%s%s\n", cursor, icon, style.Render(choice.text))
	}

	// Streak display (moved to bottom)
	streakText := HelpStyle.Render("Start a session to begin your streak!")
	if m.streak > 0 {
		streakText = StreakStyle.Render(fmt.Sprintf("⚡ %d day streak", m.streak))
	}

	// Help
	help := HelpStyle.Render("↑/↓ navigate • enter select • q quit")

	return fmt.Sprintf("\n%s\n  %s\n\n%s\n  %s\n\n  %s\n", title, version, items, streakText, help)
}

// MenuSelectionMsg is sent when a menu item is selected
type MenuSelectionMsg int
