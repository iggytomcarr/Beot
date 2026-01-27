# Lesson 3: Multiple Views & Navigation

In this lesson you'll build a proper application structure with a menu, multiple screens, and navigation between them.

---

## What You'll Learn

- The view switching pattern
- Creating an app container model
- Routing messages to child views
- Shared styles across the application
- Custom message types for navigation

---

## The View Switching Pattern

So far we've had one model, one view. Real apps have multiple screens. The pattern is simple:

```go
type View int

const (
    MenuView View = iota
    TimerView
    StatsView
)

type AppModel struct {
    currentView View
    menu        MenuModel
    timer       TimerModel
}
```

The `AppModel` is a container. It holds child models and routes messages to the active one.

---

## Project Structure

Time to organise. Create this structure:

```
beot/
├── main.go
└── ui/
    ├── app.go      # Main container
    ├── menu.go     # Menu view
    ├── timer.go    # Timer view (from Lesson 2)
    └── styles.go   # Shared styles
```

```bash
mkdir ui
```

---

## Exercise 3.1: Shared Styles

Create `ui/styles.go`:

```go
package ui

import "github.com/charmbracelet/lipgloss"

// Colours
var (
	Primary   = lipgloss.Color("205") // Pink
	Secondary = lipgloss.Color("86")  // Cyan
	Success   = lipgloss.Color("82")  // Green
	Warning   = lipgloss.Color("214") // Orange
	Danger    = lipgloss.Color("196") // Red
	Muted     = lipgloss.Color("241") // Gray
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
			Foreground(lipgloss.Color("252"))

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger).
			Bold(true)

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
)
```

Now all views use consistent colours and styles.

---

## Exercise 3.2: The Menu Model

Create `ui/menu.go`:

```go
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
	QuitApp
)

// MenuModel handles the main menu
type MenuModel struct {
	choices []string
	cursor  int
	streak  int // We'll populate this later from the database
}

// NewMenuModel creates a new menu
func NewMenuModel() MenuModel {
	return MenuModel{
		choices: []string{
			"🍅 Start Focus Session",
			"📊 View Statistics",
			"💬 Manage Quotes",
			"👋 Quit",
		},
		cursor: 0,
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
	// Title
	title := TitleStyle.Render("🍅 Beot")

	// Streak display
	streakText := HelpStyle.Render("Start a session to begin your streak!")
	if m.streak > 0 {
		streakText = SuccessStyle.Render(fmt.Sprintf("🔥 %d day streak", m.streak))
	}

	// Menu items
	var items string
	for i, choice := range m.choices {
		cursor := "  "
		style := NormalStyle

		if m.cursor == i {
			cursor = "▸ "
			style = SelectedStyle
		}

		items += fmt.Sprintf("%s%s\n", cursor, style.Render(choice))
	}

	// Help
	help := HelpStyle.Render("↑/↓ navigate • enter select • q quit")

	return fmt.Sprintf("\n  %s\n  %s\n\n%s\n  %s\n", title, streakText, items, help)
}

// MenuSelectionMsg is sent when a menu item is selected
type MenuSelectionMsg int
```

**Key concepts:**

1. **Custom message type:** `MenuSelectionMsg` tells the parent what was selected
2. **Command returning a message:** `func() tea.Msg { return MenuSelectionMsg(m.cursor) }`
3. **Cursor-based selection:** Track which item is highlighted

---

## Exercise 3.3: Timer Model (Refactored)

Move your timer to `ui/timer.go` and add messages for communication:

```go
package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// Timer messages
type tickMsg time.Time

// TimerCompleteMsg is sent when the timer finishes
type TimerCompleteMsg struct {
	Completed bool // true = completed, false = abandoned
}

// TimerModel handles the countdown
type TimerModel struct {
	totalSeconds     int
	remainingSeconds int
	running          bool
	progress         progress.Model
	confirming       bool
}

// NewTimerModel creates a timer for the given minutes
func NewTimerModel(minutes int) TimerModel {
	seconds := minutes * 60
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40

	return TimerModel{
		totalSeconds:     seconds,
		remainingSeconds: seconds,
		running:          false,
		progress:         prog,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Start begins the timer
func (m *TimerModel) Start() tea.Cmd {
	m.running = true
	return tickCmd()
}

func (m TimerModel) Init() tea.Cmd {
	return m.Start()
}

func (m TimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.confirming {
			switch msg.String() {
			case "y":
				return m, func() tea.Msg {
					return TimerCompleteMsg{Completed: false} // Abandoned
				}
			case "n", "esc":
				m.confirming = false
				m.running = true
				return m, tickCmd()
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc":
			m.confirming = true
			m.running = false
			return m, nil
		case " ":
			m.running = !m.running
			if m.running {
				return m, tickCmd()
			}
			return m, nil
		}

	case tickMsg:
		if m.running && m.remainingSeconds > 0 {
			m.remainingSeconds--
			if m.remainingSeconds <= 0 {
				m.running = false
				fmt.Print("\a") // Terminal bell
				return m, func() tea.Msg {
					return TimerCompleteMsg{Completed: true}
				}
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

func (m TimerModel) View() string {
	if m.confirming {
		return m.renderConfirmation()
	}

	if m.remainingSeconds <= 0 {
		return m.renderComplete()
	}

	return m.renderTimer()
}

func (m TimerModel) renderTimer() string {
	elapsed := m.totalSeconds - m.remainingSeconds
	percent := float64(elapsed) / float64(m.totalSeconds)

	minutes := m.remainingSeconds / 60
	seconds := m.remainingSeconds % 60
	timeDisplay := TitleStyle.Render(fmt.Sprintf("%02d:%02d", minutes, seconds))

	status := SubtitleStyle.Render("🍅 Focus Time")
	if !m.running {
		status = WarningStyle.Render("⏸  Paused")
	}

	progressBar := m.progress.ViewAs(percent)
	percentText := HelpStyle.Render(fmt.Sprintf("(%d%%)", int(percent*100)))
	help := HelpStyle.Render("space pause • q/esc abandon")

	return fmt.Sprintf(
		"\n  %s\n\n  %s\n\n  %s  %s\n\n  %s\n",
		status,
		progressBar,
		timeDisplay,
		percentText,
		help,
	)
}

func (m TimerModel) renderConfirmation() string {
	title := ErrorStyle.Render("Give up?")
	message := "This will be logged as abandoned 💀"
	help := HelpStyle.Render("[y] yes, abandon • [n] no, continue")

	return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", title, message, help)
}

func (m TimerModel) renderComplete() string {
	content := fmt.Sprintf(
		"%s\n\n%s",
		SuccessStyle.Render("✅ Session Complete!"),
		HelpStyle.Render("Press any key to continue"),
	)

	return "\n" + BoxStyle.Render(content) + "\n"
}
```

**Changes from Lesson 2:**

1. Uses shared styles from `styles.go`
2. Sends `TimerCompleteMsg` instead of quitting directly
3. Split View() into smaller render functions
4. Exported types start with capital letters (Go convention for packages)

---

## Exercise 3.4: The App Container

Create `ui/app.go`:

```go
package ui

import tea "github.com/charmbracelet/bubbletea"

// View represents which screen is active
type View int

const (
	MenuViewState View = iota
	TimerViewState
	StatsViewState
)

// AppModel is the main application container
type AppModel struct {
	currentView View
	menu        MenuModel
	timer       TimerModel
}

// NewAppModel creates the application
func NewAppModel() AppModel {
	return AppModel{
		currentView: MenuViewState,
		menu:        NewMenuModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle messages that affect navigation
	switch msg := msg.(type) {

	case MenuSelectionMsg:
		switch MenuChoice(msg) {
		case StartSession:
			m.timer = NewTimerModel(25) // 25-minute pomodoro
			m.currentView = TimerViewState
			return m, m.timer.Init()
		case ViewStats:
			m.currentView = StatsViewState
			return m, nil
		case QuitApp:
			return m, tea.Quit
		}
		return m, nil

	case TimerCompleteMsg:
		// Timer finished (completed or abandoned)
		// Later: save session to database here
		m.currentView = MenuViewState
		return m, nil
	}

	// Route messages to the active view
	switch m.currentView {
	case MenuViewState:
		newMenu, cmd := m.menu.Update(msg)
		m.menu = newMenu.(MenuModel)
		return m, cmd

	case TimerViewState:
		newTimer, cmd := m.timer.Update(msg)
		m.timer = newTimer.(TimerModel)
		return m, cmd

	case StatsViewState:
		// Handle stats view input
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" || keyMsg.String() == "q" {
				m.currentView = MenuViewState
				return m, nil
			}
		}
	}

	return m, nil
}

func (m AppModel) View() string {
	switch m.currentView {
	case MenuViewState:
		return m.menu.View()
	case TimerViewState:
		return m.timer.View()
	case StatsViewState:
		return m.renderStats()
	default:
		return "Unknown view"
	}
}

func (m AppModel) renderStats() string {
	title := TitleStyle.Render("📊 Statistics")
	placeholder := NormalStyle.Render("Coming soon...")
	help := HelpStyle.Render("esc/q back to menu")

	return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", title, placeholder, help)
}
```

**Don't forget the import:**

```go
import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)
```

---

## Exercise 3.5: Update main.go

Update `main.go`:

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"Beot/ui"
)

func main() {
	p := tea.NewProgram(ui.NewAppModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
```

**Note:** `tea.WithAltScreen()` uses the alternate terminal buffer - the app takes over the whole screen and restores it when done.

**Run it:**

```bash
go run main.go
```

---

## Understanding Message Routing

The key insight is how messages flow:

```
User presses 'enter' on menu
    ↓
MenuModel.Update() receives KeyMsg
    ↓
MenuModel returns MenuSelectionMsg command
    ↓
AppModel.Update() receives MenuSelectionMsg
    ↓
AppModel changes currentView, creates TimerModel
    ↓
AppModel returns timer's Init() command
    ↓
Timer starts ticking
```

Each model only handles messages it understands. The parent routes appropriately.

---

## Type Assertions

You'll see this pattern often:

```go
newMenu, cmd := m.menu.Update(msg)
m.menu = newMenu.(MenuModel)  // Type assertion
```

`Update()` returns `tea.Model` (an interface), but we need `MenuModel`. The `.(MenuModel)` asserts the type.

**If you're worried about panics:**

```go
if newMenu, ok := newMenu.(MenuModel); ok {
    m.menu = newMenu
}
```

But in practice, if your code is correct, the assertion won't fail.

---

## Checkpoint Tasks

Before moving to Lesson 4, make sure you can:

- [ ] Navigate the menu with arrow keys
- [ ] Select "Start Focus Session" to begin timer
- [ ] Timer counts down and shows progress
- [ ] Abandoning returns to menu
- [ ] Completing returns to menu
- [ ] "View Statistics" shows placeholder
- [ ] ESC from stats returns to menu
- [ ] "Quit" exits the app
- [ ] **Challenge:** Add keyboard shortcut numbers (1, 2, 3, 4) to select menu items directly
- [ ] **Challenge:** Add a "Settings" menu option that shows a placeholder view
- [ ] **Challenge:** Pass the timer duration as a parameter when selecting "Start Session"

---

## Common Gotchas

### "My timer doesn't start"

Make sure you return the timer's Init() command:

```go
case StartSession:
    m.timer = NewTimerModel(25)
    m.currentView = TimerViewState
    return m, m.timer.Init()  // Don't forget this!
```

### "Messages go to the wrong view"

Check your routing in AppModel.Update():

```go
switch m.currentView {
case MenuViewState:
    // Only send to menu when menu is active
case TimerViewState:
    // Only send to timer when timer is active
}
```

### "Import errors"

Make sure your module name matches. In `go.mod`:

```
module Beot
```

Then import as:

```go
import "Beot/ui"
```

### "Styles not found"

Styles must be exported (capital letter) to use across packages:

```go
var TitleStyle = ...  // Exported - can use in other files
var titleStyle = ...  // Not exported - only in styles.go
```

---

## What's Next

In **Lesson 4**, we'll add MongoDB:
- Database connection
- Repositories for CRUD operations
- Saving sessions
- Loading quotes

The structure is in place. Time to persist data!

---

## Quick Reference

```go
// View switching pattern
type View int

const (
    ViewA View = iota
    ViewB
)

type AppModel struct {
    currentView View
    viewA       ViewAModel
    viewB       ViewBModel
}

// Route in Update:
switch m.currentView {
case ViewA:
    newA, cmd := m.viewA.Update(msg)
    m.viewA = newA.(ViewAModel)
    return m, cmd
case ViewB:
    // ...
}

// Route in View:
switch m.currentView {
case ViewA:
    return m.viewA.View()
case ViewB:
    return m.viewB.View()
}

// Custom message for navigation
type NavigateMsg struct{ Target View }

// Sending navigation message
return m, func() tea.Msg {
    return NavigateMsg{Target: ViewB}
}

// Alternate screen (full app takeover)
tea.NewProgram(model, tea.WithAltScreen())
```

---

*"Game design is decision making, and decisions must be made with confidence."*

You've made the architecture decision. The views flow. Navigate onwards. 🗺️
