# Lesson 2: Building the Timer

In this lesson you'll build a countdown timer using Bubble Tea's command system. This is where the Elm Architecture really shines.

---

## What You'll Learn

- How commands and messages work together
- Using `tea.Tick` for time-based updates
- The progress bar component from Bubbles
- Handling timer completion

---

## Understanding Commands

In Lesson 1, we only returned `nil` or `tea.Quit` from Update(). But Bubble Tea can do much more through **commands**.

A command is a function that does something (usually async) and then sends a message back:

```
Update() returns Command
    ↓
Command runs (e.g., waits 1 second)
    ↓
Command sends Message
    ↓
Update() receives Message
    ↓
(cycle repeats)
```

For our timer, we need `tea.Tick` - it waits for a duration, then sends a message:

```go
tea.Tick(time.Second, func(t time.Time) tea.Msg {
    return tickMsg(t)
})
```

---

## Exercise 2.1: Basic Countdown

Let's start with a simple countdown. Create `main.go`:

```go
package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Custom message type for our tick
type tickMsg time.Time

// Styles
var (
	timerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

type model struct {
	totalSeconds     int
	remainingSeconds int
	running          bool
}

// newModel creates a timer for the given minutes
func newModel(minutes int) model {
	seconds := minutes * 60
	return model{
		totalSeconds:     seconds,
		remainingSeconds: seconds,
		running:          false,
	}
}

// tickCmd returns a command that sends a tick after 1 second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	// Start the timer immediately
	m.running = true
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ": // Spacebar to pause/resume
			m.running = !m.running
			if m.running {
				return m, tickCmd()
			}
			return m, nil
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
	}

	return m, nil
}

func (m model) View() string {
	// Format time as MM:SS
	minutes := m.remainingSeconds / 60
	seconds := m.remainingSeconds % 60
	timeDisplay := timerStyle.Render(fmt.Sprintf("%02d:%02d", minutes, seconds))

	// Status
	status := statusStyle.Render("🍅 Focus Time")
	if !m.running && m.remainingSeconds > 0 {
		status = statusStyle.Render("⏸  Paused")
	} else if m.remainingSeconds <= 0 {
		status = statusStyle.Render("✅ Complete!")
	}

	// Help
	help := helpStyle.Render("space pause • r reset • q quit")

	return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", status, timeDisplay, help)
}

func main() {
	// Start with 1 minute for testing (change to 25 for real use)
	p := tea.NewProgram(newModel(1))
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

**Try:**
- Watch it count down
- Press `space` to pause/resume
- Press `r` to reset
- Let it complete

---

## Understanding the Tick Cycle

Here's what happens:

1. `Init()` returns `tickCmd()` - schedules first tick
2. After 1 second, `tickMsg` arrives at `Update()`
3. We decrement `remainingSeconds`
4. If not zero, return another `tickCmd()` - schedules next tick
5. If zero, return `nil` - no more ticks

**Key insight:** We don't have a "running loop". Each tick schedules the next one. Stop returning `tickCmd()` and the timer stops.

```go
case tickMsg:
    if m.running && m.remainingSeconds > 0 {
        m.remainingSeconds--           // Decrement
        if m.remainingSeconds <= 0 {
            m.running = false
            return m, nil              // Stop ticking
        }
        return m, tickCmd()            // Keep ticking
    }
```

---

## Exercise 2.2: Adding a Progress Bar

The `bubbles` package has a built-in progress bar. Let's use it.

```go
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

var (
	timerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

type model struct {
	totalSeconds     int
	remainingSeconds int
	running          bool
	progress         progress.Model
}

func newModel(minutes int) model {
	seconds := minutes * 60

	// Create progress bar with gradient colours
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40

	return model{
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

func (m model) Init() tea.Cmd {
	m.running = true
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ":
			m.running = !m.running
			if m.running {
				return m, tickCmd()
			}
			return m, nil
		case "r":
			m.remainingSeconds = m.totalSeconds
			m.running = true
			return m, tickCmd()
		}

	case tickMsg:
		if m.running && m.remainingSeconds > 0 {
			m.remainingSeconds--
			if m.remainingSeconds <= 0 {
				m.running = false
				return m, nil
			}
			return m, tickCmd()
		}

	// Handle progress bar's internal messages
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	// Calculate progress (0.0 to 1.0)
	elapsed := m.totalSeconds - m.remainingSeconds
	percent := float64(elapsed) / float64(m.totalSeconds)

	// Format time
	minutes := m.remainingSeconds / 60
	seconds := m.remainingSeconds % 60
	timeDisplay := timerStyle.Render(fmt.Sprintf("%02d:%02d", minutes, seconds))

	// Status
	status := statusStyle.Render("🍅 Focus Time")
	if !m.running && m.remainingSeconds > 0 {
		status = statusStyle.Render("⏸  Paused")
	} else if m.remainingSeconds <= 0 {
		status = statusStyle.Render("✅ Complete!")
	}

	// Progress bar
	progressBar := m.progress.ViewAs(percent)

	help := helpStyle.Render("space pause • r reset • q quit")

	return fmt.Sprintf(
		"\n  %s\n\n  %s\n\n  %s  %s\n\n  %s\n",
		status,
		progressBar,
		timeDisplay,
		helpStyle.Render(fmt.Sprintf("(%d%% complete)", int(percent*100))),
		help,
	)
}

func main() {
	p := tea.NewProgram(newModel(1))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
```

**New concepts:**

1. **Bubbles component:** `progress.Model` is a pre-built component
2. **Component messages:** Progress bar has its own internal messages (`progress.FrameMsg`) for smooth animations
3. **ViewAs:** `m.progress.ViewAs(percent)` renders the bar at a specific percentage

---

## Progress Bar Options

```go
// Default gradient (purple to pink)
prog := progress.New(progress.WithDefaultGradient())

// Custom gradient
prog := progress.New(progress.WithGradient("#FF0000", "#00FF00"))

// Solid colour
prog := progress.New(progress.WithSolidFill("#FF5733"))

// No percentage shown
prog := progress.New(
    progress.WithDefaultGradient(),
    progress.WithoutPercentage(),
)

// Custom width
prog.Width = 60
```

---

## Exercise 2.3: Abandon Confirmation

A core feature of Beot: when you quit early, it counts as "abandoned". Let's add a confirmation:

```go
type model struct {
	totalSeconds     int
	remainingSeconds int
	running          bool
	progress         progress.Model
	confirming       bool // NEW: are we showing quit confirmation?
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// If we're showing the confirmation dialog
		if m.confirming {
			switch msg.String() {
			case "y":
				// Yes, abandon
				return m, tea.Quit
			case "n", "esc":
				// No, continue
				m.confirming = false
				m.running = true
				return m, tickCmd()
			}
			return m, nil // Ignore other keys during confirmation
		}

		// Normal key handling
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			// Show confirmation instead of quitting
			m.confirming = true
			m.running = false // Pause while confirming
			return m, nil
		case " ":
			m.running = !m.running
			if m.running {
				return m, tickCmd()
			}
			return m, nil
		case "r":
			m.remainingSeconds = m.totalSeconds
			m.running = true
			return m, tickCmd()
		}

	case tickMsg:
		if m.running && m.remainingSeconds > 0 {
			m.remainingSeconds--
			if m.remainingSeconds <= 0 {
				m.running = false
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
	// Show confirmation dialog
	if m.confirming {
		return fmt.Sprintf(
			"\n  %s\n\n  %s\n\n  %s\n",
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("Give up?"),
			"This will be logged as abandoned 💀",
			helpStyle.Render("[y] yes, abandon • [n] no, continue"),
		)
	}

	// ... rest of normal view
}
```

**Key pattern:** The `confirming` boolean changes what Update() and View() do. Same model, different behaviour based on state.

---

## Checkpoint Tasks

Before moving to Lesson 3, make sure you can:

- [ ] Timer counts down every second
- [ ] Progress bar fills as time passes
- [ ] Space pauses and resumes
- [ ] 'r' resets the timer
- [ ] 'q' shows confirmation dialog
- [ ] 'n' or ESC returns to timer
- [ ] 'y' quits the app
- [ ] Timer shows "Complete!" when done
- [ ] **Challenge:** Add a terminal bell when complete: `fmt.Print("\a")`
- [ ] **Challenge:** Change the progress bar colour based on time remaining (green > 50%, yellow > 25%, red < 25%)
- [ ] **Challenge:** Add keyboard shortcut to add 5 minutes to the timer

---

## Common Gotchas

### "Timer keeps going when paused"

Make sure you check `m.running` before processing ticks:

```go
case tickMsg:
    if m.running && m.remainingSeconds > 0 {  // Check running first!
        // ...
    }
```

### "Timer runs twice as fast"

You're probably returning `tickCmd()` twice. Only return it once per tick:

```go
// Wrong - ticks will queue up
return m, tea.Batch(tickCmd(), tickCmd())

// Right - one tick at a time
return m, tickCmd()
```

### "Progress bar jumps instead of animating"

Make sure you handle `progress.FrameMsg`:

```go
case progress.FrameMsg:
    progressModel, cmd := m.progress.Update(msg)
    m.progress = progressModel.(progress.Model)
    return m, cmd
```

### "Weird behaviour when pausing"

When pausing, make sure you return `nil` instead of `tickCmd()`:

```go
case " ":
    m.running = !m.running
    if m.running {
        return m, tickCmd()  // Resume: start ticking
    }
    return m, nil            // Pause: stop ticking
```

---

## What's Next

In **Lesson 3**, we'll build a menu system with multiple views:
- Main menu with navigation
- Switching between menu and timer
- A shared styles file
- Message routing between views

The timer works - now we need a way to start it!

---

## Quick Reference

```go
// Tick command - runs after duration, sends message
tea.Tick(time.Second, func(t time.Time) tea.Msg {
    return myMsg(t)
})

// Progress bar
import "github.com/charmbracelet/bubbles/progress"

prog := progress.New(progress.WithDefaultGradient())
prog.Width = 40

// In View:
progressBar := m.progress.ViewAs(0.5) // 50%

// In Update - handle progress messages:
case progress.FrameMsg:
    progressModel, cmd := m.progress.Update(msg)
    m.progress = progressModel.(progress.Model)
    return m, cmd

// Batch multiple commands
tea.Batch(cmd1, cmd2, cmd3)
```

---

*"If you aren't dropping, you aren't learning. And if you aren't learning, you aren't a juggler."*

The timer ticks. The vow holds. Onwards. 🍅
