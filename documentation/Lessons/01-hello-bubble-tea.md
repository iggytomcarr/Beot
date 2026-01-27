# Lesson 1: Hello Bubble Tea

In this lesson you'll build your first Bubble Tea application and understand the core pattern that everything else builds on.

---

## What You'll Learn

- The Elm Architecture (Model-Update-View)
- How Bubble Tea handles keyboard input
- Basic Lip Gloss styling
- Running and testing your first TUI app

---

## The Elm Architecture

Bubble Tea uses a pattern called the Elm Architecture. Understand this and everything else clicks:

```
┌─────────────────────────────────────────────────────────┐
│                                                         │
│   Model ─────► View() ─────► Terminal Output            │
│     ▲                                                   │
│     │                                                   │
│     └─────── Update(msg) ◄──── User Input / Events      │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Three parts:**

1. **Model** - A struct that holds ALL your application state
2. **Update** - A function that receives messages (keypresses, timer ticks, etc.) and returns new state
3. **View** - A function that takes the model and returns a string to display

That's it. Data flows one direction. The view is always a pure function of the current state.

---

## Exercise 1.1: The Simplest Program

Create your project structure:

```bash
cd C:\Learning\Beot
```

Create `main.go`:

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Model holds ALL application state
type model struct {
	message string
}

// Init is called once when the program starts
// Return nil if you don't need to run any initial commands
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and returns updated model + optional command
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Handle keyboard input
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI as a string
func (m model) View() string {
	return fmt.Sprintf("\n  %s\n\n  Press 'q' to quit.\n", m.message)
}

func main() {
	// Create initial model
	initialModel := model{
		message: "Hello, Bubble Tea! 🧋",
	}

	// Create and run the program
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
```

**Run it:**

```bash
go run main.go
```

You should see:

```
  Hello, Bubble Tea! 🧋

  Press 'q' to quit.
```

Press `q` to exit.

---

## Understanding the Code

### The Model

```go
type model struct {
	message string
}
```

This is just a Go struct. It can have any fields you want. Everything your app needs to "remember" goes here.

### Init()

```go
func (m model) Init() tea.Cmd {
	return nil
}
```

Called once at startup. Return `nil` if you don't need to do anything. Later we'll use this to start timers or load data.

### Update()

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}
```

This is where things happen. Every keypress, timer tick, or event becomes a "message" that arrives here.

- `msg.(type)` is Go's type switch - it checks what kind of message we got
- `tea.KeyMsg` means someone pressed a key
- `msg.String()` gives us the key as a string ("q", "enter", "up", etc.)
- `tea.Quit` is a special command that exits the program

### View()

```go
func (m model) View() string {
	return fmt.Sprintf("\n  %s\n\n  Press 'q' to quit.\n", m.message)
}
```

This just returns a string. That string is what gets displayed. Every time Update() returns, View() is called again automatically.

---

## Exercise 1.2: Adding Interactivity

Let's make something you can interact with. Update your `main.go`:

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	counter int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.counter++
		case "down", "j":
			m.counter--
		case "r":
			m.counter = 0
		}
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf(
		"\n  Counter: %d\n\n  ↑/k increase • ↓/j decrease • r reset • q quit\n",
		m.counter,
	)
}

func main() {
	p := tea.NewProgram(model{counter: 0})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
```

**Run it and try:**
- Press `↑` or `k` to increase
- Press `↓` or `j` to decrease
- Press `r` to reset
- Press `q` to quit

**Key insight:** We never directly update the display. We just modify the model in Update(), and View() automatically shows the new state.

---

## Exercise 1.3: Adding Lip Gloss Styling

Now let's make it look good. Lip Gloss is Charm's styling library.

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define styles as package-level variables
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")). // Pink
			MarginBottom(1)

	counterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")). // Cyan
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")) // Gray
)

type model struct {
	counter int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.counter++
		case "down", "j":
			m.counter--
		case "r":
			m.counter = 0
		}
	}
	return m, nil
}

func (m model) View() string {
	title := titleStyle.Render("⚡ Counter App")
	counter := counterStyle.Render(fmt.Sprintf("%d", m.counter))
	help := helpStyle.Render("↑/k up • ↓/j down • r reset • q quit")

	return fmt.Sprintf("\n  %s\n\n  Count: %s\n\n  %s\n", title, counter, help)
}

func main() {
	p := tea.NewProgram(model{counter: 0})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
```

**Run it** - you should see colours now!

---

## Lip Gloss Basics

Styles are created with `lipgloss.NewStyle()` and chained:

```go
style := lipgloss.NewStyle().
	Bold(true).
	Italic(true).
	Foreground(lipgloss.Color("205")).  // Text colour
	Background(lipgloss.Color("0")).    // Background colour
	Padding(1, 2).                       // Vertical, Horizontal
	Margin(1).                           // All sides
	Border(lipgloss.RoundedBorder()).   // Box border
	Width(40)                            // Fixed width
```

Apply with `.Render()`:

```go
output := style.Render("Hello!")
```

**Colours:** You can use:
- ANSI numbers: `lipgloss.Color("205")` (0-255)
- Hex codes: `lipgloss.Color("#FF5733")`
- Named: `lipgloss.Color("red")`

---

## Exercise 1.4: Add a Border

Let's wrap the whole thing in a nice box:

```go
var (
	// ... existing styles ...

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 2)
)

func (m model) View() string {
	title := titleStyle.Render("⚡ Counter App")
	counter := counterStyle.Render(fmt.Sprintf("%d", m.counter))
	help := helpStyle.Render("↑/k up • ↓/j down • r reset • q quit")

	content := fmt.Sprintf("%s\n\nCount: %s\n\n%s", title, counter, help)

	return "\n" + boxStyle.Render(content) + "\n"
}
```

---

## Checkpoint Tasks

Before moving to Lesson 2, make sure you can:

- [ ] Run the app and see styled output
- [ ] Increase/decrease the counter with keys
- [ ] Reset with 'r'
- [ ] Quit with 'q'
- [ ] **Challenge:** Add a key that doubles the counter
- [ ] **Challenge:** Add a key that shows/hides the help text (you'll need a `showHelp bool` in your model)
- [ ] **Challenge:** Change the colours to your preference
- [ ] **Challenge:** Make the counter turn red when negative

---

## Common Gotchas

### "Why isn't my state updating?"

Make sure you're returning the modified model:
```go
// Wrong - modifies but doesn't return
m.counter++
return m, nil  // This works because we modified m

// If you had a pointer receiver, be careful:
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // With pointers, modifications persist automatically
}
```

### "My styles aren't showing"

- Make sure your terminal supports colours (most do)
- Windows Terminal, VS Code terminal, and PowerShell all work
- Old cmd.exe might not show colours properly

### "The screen flickers"

Bubble Tea handles this for you, but if you see issues:
```go
p := tea.NewProgram(model{}, tea.WithAltScreen())
```
`WithAltScreen()` uses the alternate terminal buffer - cleaner for full-screen apps.

---

## What's Next

In **Lesson 2**, we'll build the actual timer with:
- Tick-based updates (every second)
- Progress bar component
- Completion detection

You now understand the core pattern. Everything else is just:
1. More fields in the model
2. More message types in Update()
3. More complex rendering in View()

---

## Quick Reference

```go
// Basic structure
type model struct { /* state */ }
func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m model) View() string { return "output" }

// Key messages
case tea.KeyMsg:
    switch msg.String() {
    case "q":        // letter
    case "ctrl+c":   // combo
    case "enter":    // special key
    case "up":       // arrow
    }

// Quit the program
return m, tea.Quit

// Lip Gloss
style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
output := style.Render("text")
```

---

*"A computer is a creative amplifier."*

You've made your first beot - a small one, but you kept it. On to the next. 🗡️
