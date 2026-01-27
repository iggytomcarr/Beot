# Lesson 8: Polish & Sound

In this final lesson you'll add the finishing touches that make Beot feel complete: sound notifications, celebration screens, and visual polish.

---

## What You'll Learn

- Terminal bell notifications
- Animated completion screens
- Visual enhancements with Lip Gloss
- Final code organization
- Ideas for future enhancements

---

## Sound Notifications

### Terminal Bell

The simplest cross-platform sound is the terminal bell:

```go
fmt.Print("\a")  // ASCII bell character (BEL)
```

This works on most terminals and produces a system notification sound.

We already added this in Lesson 2, but let's make sure it's in the right place in `ui/timer.go`:

```go
case tickMsg:
    if m.running && m.remainingSeconds > 0 {
        m.remainingSeconds--
        if m.remainingSeconds <= 0 {
            m.running = false
            fmt.Print("\a")  // Terminal bell on completion
            return m, func() tea.Msg {
                return TimerCompleteMsg{
                    Completed: true,
                    Subject:   m.subject,
                    Duration:  m.totalSeconds / 60,
                    StartedAt: m.startedAt,
                }
            }
        }
        return m, tickCmd()
    }
```

---

## Exercise 8.1: Enhanced Completion Screen

Let's make the completion screen more celebratory. Update the `renderComplete` method in `ui/timer.go`:

```go
func (m TimerModel) renderComplete() string {
	// Celebration header
	celebration := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("82")).
		Background(lipgloss.Color("22")).
		Padding(0, 2).
		Render("✨ VICTORY ✨")

	// Subject completed
	subject := fmt.Sprintf("%s %s", m.subject.Icon, m.subject.Name)
	subjectStyled := SubtitleStyle.Render(subject)

	// Duration
	duration := fmt.Sprintf("%d minutes of focused work!", m.totalSeconds/60)
	durationStyled := HelpStyle.Render(duration)

	// Motivational message
	messages := []string{
		"You kept your vow. 🗡️",
		"Another beot fulfilled.",
		"Your focus is your power.",
		"The streak grows stronger.",
		"Well done, warrior.",
	}
	// Pick based on time for variety
	msgIdx := time.Now().Second() % len(messages)
	message := SuccessStyle.Render(messages[msgIdx])

	// Final quote if available
	var quoteSection string
	if len(m.quotes) > 0 {
		q := m.quotes[m.currentQuoteIdx]
		quoteStyle := lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("243"))

		quoteSection = fmt.Sprintf("\n\n%s", quoteStyle.Render(fmt.Sprintf("\"%s\"", q.Text)))
		if q.Source != "" {
			quoteSection += fmt.Sprintf("\n%s", HelpStyle.Render("— "+q.Source))
		}
	}

	// Build the box content
	content := fmt.Sprintf(
		"%s\n\n%s\n%s\n\n%s%s",
		celebration,
		subjectStyled,
		durationStyled,
		message,
		quoteSection,
	)

	// Styled box
	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("82")).
		Padding(1, 3).
		Align(lipgloss.Center)

	help := HelpStyle.Render("\nPress any key to continue")

	return "\n\n" + box.Render(content) + help + "\n"
}
```

Don't forget to add the time import:

```go
import (
	"fmt"
	"math/rand"
	"time"
	// ... other imports
)
```

---

## Exercise 8.2: Abandoned Screen

When abandoning, we should also show a meaningful screen. Add this to `ui/timer.go`:

```go
func (m TimerModel) renderAbandoned() string {
	// Somber header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Render("💀 Session Abandoned")

	subject := fmt.Sprintf("%s %s", m.subject.Icon, m.subject.Name)

	// How far they got
	elapsed := m.totalSeconds - m.remainingSeconds
	elapsedMins := elapsed / 60
	elapsedSecs := elapsed % 60
	progress := fmt.Sprintf("You made it %d:%02d before stopping.", elapsedMins, elapsedSecs)

	// Accountability message
	messages := []string{
		"This will be remembered.",
		"The vow was broken.",
		"Tomorrow is another chance.",
		"Even warriors need rest sometimes.",
	}
	msgIdx := time.Now().Second() % len(messages)
	message := HelpStyle.Render(messages[msgIdx])

	content := fmt.Sprintf(
		"%s\n\n%s\n%s\n\n%s",
		header,
		SubtitleStyle.Render(subject),
		HelpStyle.Render(progress),
		message,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 3).
		Align(lipgloss.Center)

	help := HelpStyle.Render("\nPress any key to continue")

	return "\n\n" + box.Render(content) + help + "\n"
}
```

Update the view logic to use it:

```go
func (m TimerModel) View() string {
	if m.confirming {
		return m.renderConfirmation()
	}

	// Check if timer was abandoned (completed=false flag would be set)
	// Actually, we navigate away on abandon, so this screen shows briefly
	// We need to handle this differently...

	if m.remainingSeconds <= 0 {
		return m.renderComplete()
	}

	return m.renderTimer()
}
```

Actually, since we immediately send a message and navigate away on abandon, we need to show this screen in the app container. Let's add a "showing result" state.

---

## Exercise 8.3: Result Screen in App

Update `ui/app.go` to show completion/abandon result before returning to menu:

```go
type View int

const (
	MenuViewState View = iota
	SubjectSelectViewState
	TimerViewState
	ResultViewState  // NEW
	StatsViewState
)

// Add to AppModel
type AppModel struct {
	// ... existing fields
	lastResult *TimerCompleteMsg  // NEW
}

// Update TimerCompleteMsg handler
case TimerCompleteMsg:
	m.lastResult = &msg
	m.currentView = ResultViewState
	// Save session in background
	return m, m.saveSession(msg.Completed, msg.Subject, msg.Duration, msg.StartedAt)

// Add result view handling
case ResultViewState:
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Any key returns to menu
		m.currentView = MenuViewState
		m.lastResult = nil
		return m, nil
	}

// Add to View()
case ResultViewState:
	return m.renderResult()
```

Add the render method:

```go
func (m AppModel) renderResult() string {
	if m.lastResult == nil {
		return "Error: No result to display"
	}

	result := m.lastResult

	if result.Completed {
		return m.renderCompletionResult(result)
	}
	return m.renderAbandonedResult(result)
}

func (m AppModel) renderCompletionResult(result *TimerCompleteMsg) string {
	celebration := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("82")).
		Background(lipgloss.Color("22")).
		Padding(0, 2).
		Render("✨ VICTORY ✨")

	subject := fmt.Sprintf("%s %s", result.Subject.Icon, result.Subject.Name)

	duration := fmt.Sprintf("%d minutes of focused work!", result.Duration)

	messages := []string{
		"You kept your vow. 🗡️",
		"Another beot fulfilled.",
		"Your focus is your power.",
		"The streak grows stronger.",
		"Well done, warrior.",
	}
	msgIdx := time.Now().Second() % len(messages)
	message := SuccessStyle.Render(messages[msgIdx])

	content := fmt.Sprintf(
		"%s\n\n%s\n%s\n\n%s",
		celebration,
		SubtitleStyle.Render(subject),
		HelpStyle.Render(duration),
		message,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("82")).
		Padding(1, 3).
		Align(lipgloss.Center)

	help := HelpStyle.Render("\n\nPress any key to continue")

	return "\n\n" + box.Render(content) + help + "\n"
}

func (m AppModel) renderAbandonedResult(result *TimerCompleteMsg) string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Render("💀 Session Abandoned")

	subject := fmt.Sprintf("%s %s", result.Subject.Icon, result.Subject.Name)

	messages := []string{
		"This will be remembered.",
		"The vow was broken.",
		"Tomorrow is another chance.",
		"Even warriors need rest sometimes.",
	}
	msgIdx := time.Now().Second() % len(messages)
	message := HelpStyle.Render(messages[msgIdx])

	content := fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		header,
		SubtitleStyle.Render(subject),
		message,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 3).
		Align(lipgloss.Center)

	help := HelpStyle.Render("\n\nPress any key to continue")

	return "\n\n" + box.Render(content) + help + "\n"
}
```

---

## Exercise 8.4: Visual Polish for Menu

Update `ui/menu.go` with better styling:

```go
func (m MenuModel) View() string {
	// App title with box
	titleBox := lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		Padding(0, 1).
		Render("🍅 BEOT")

	subtitle := HelpStyle.Render("A binding vow to focus")

	// Streak display
	var streakText string
	if m.streak > 0 {
		fires := "🔥"
		if m.streak >= 7 {
			fires = "🔥🔥"
		}
		if m.streak >= 30 {
			fires = "🔥🔥🔥"
		}
		streakText = SuccessStyle.Render(fmt.Sprintf("%s %d day streak", fires, m.streak))
	} else {
		streakText = HelpStyle.Render("Start a session to begin your streak!")
	}

	// Menu items with better spacing
	var items string
	for i, choice := range m.choices {
		cursor := "   "
		style := NormalStyle

		if m.cursor == i {
			cursor = " ▸ "
			style = SelectedStyle
		}

		items += fmt.Sprintf("%s%s\n", cursor, style.Render(choice))
	}

	// Help
	help := HelpStyle.Render("↑/↓ navigate • enter select • q quit")

	return fmt.Sprintf(
		"\n  %s\n  %s\n\n  %s\n\n%s\n  %s\n",
		titleBox,
		subtitle,
		streakText,
		items,
		help,
	)
}
```

---

## Exercise 8.5: Warning Bell Before End

Add a warning sound when 1 minute remains. Update `ui/timer.go`:

```go
case tickMsg:
    if m.running && m.remainingSeconds > 0 {
        m.remainingSeconds--

        // Warning at 1 minute remaining
        if m.remainingSeconds == 60 {
            fmt.Print("\a")  // Warning bell
        }

        if m.remainingSeconds <= 0 {
            m.running = false
            // Multiple bells for completion
            fmt.Print("\a\a\a")
            return m, func() tea.Msg {
                return TimerCompleteMsg{
                    Completed: true,
                    Subject:   m.subject,
                    Duration:  m.totalSeconds / 60,
                    StartedAt: m.startedAt,
                }
            }
        }
        return m, tickCmd()
    }
```

---

## Exercise 8.6: Color-Coded Progress

Make the timer change colour as time runs out. Update `renderTimer` in `ui/timer.go`:

```go
func (m TimerModel) renderTimer() string {
	elapsed := m.totalSeconds - m.remainingSeconds
	percent := float64(elapsed) / float64(m.totalSeconds)
	remainingPercent := 1 - percent

	// Color based on remaining time
	var timeColor lipgloss.Color
	var statusEmoji string
	switch {
	case remainingPercent > 0.5:
		timeColor = lipgloss.Color("82")  // Green
		statusEmoji = "🍅"
	case remainingPercent > 0.25:
		timeColor = lipgloss.Color("214") // Orange
		statusEmoji = "🍅"
	default:
		timeColor = lipgloss.Color("196") // Red
		statusEmoji = "⏰"
	}

	minutes := m.remainingSeconds / 60
	seconds := m.remainingSeconds % 60

	timeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(timeColor)

	timeDisplay := timeStyle.Render(fmt.Sprintf("%02d:%02d", minutes, seconds))

	// Subject
	subjectDisplay := SubtitleStyle.Render(fmt.Sprintf("%s %s", m.subject.Icon, m.subject.Name))

	status := fmt.Sprintf("%s Focus Time", statusEmoji)
	if !m.running {
		status = "⏸  Paused"
	}

	// Progress bar
	progressBar := m.progress.ViewAs(percent)

	// Quote
	quoteDisplay := m.renderQuote()

	help := HelpStyle.Render("space pause • q/esc abandon")

	return fmt.Sprintf(
		"\n  %s\n  %s\n\n  %s\n\n  %s\n%s\n  %s\n",
		subjectDisplay,
		HelpStyle.Render(status),
		progressBar,
		timeDisplay,
		quoteDisplay,
		help,
	)
}
```

---

## Final Testing Checklist

Run through the complete app flow:

```bash
go run main.go
```

1. **Menu**
   - [ ] Title and subtitle display
   - [ ] Streak shows (or "Start a session..." message)
   - [ ] Navigation works
   - [ ] Selection works

2. **Subject Selection**
   - [ ] All subjects display with icons
   - [ ] Colours show correctly
   - [ ] Selection starts timer

3. **Timer**
   - [ ] Countdown works
   - [ ] Progress bar fills
   - [ ] Quote displays and rotates
   - [ ] Pause/resume works
   - [ ] Colour changes as time decreases
   - [ ] Warning bell at 1 minute (if testing with longer timer)
   - [ ] Completion bell sounds

4. **Completion Screen**
   - [ ] Victory message displays
   - [ ] Subject shown
   - [ ] Duration shown
   - [ ] Motivational message
   - [ ] Any key returns to menu

5. **Abandon Flow**
   - [ ] Confirmation dialog appears
   - [ ] 'y' shows abandoned screen
   - [ ] 'n' resumes timer
   - [ ] Abandoned screen displays
   - [ ] Session saved as abandoned

6. **Statistics**
   - [ ] Loads data correctly
   - [ ] Shows streak
   - [ ] Shows today's sessions
   - [ ] Shows completion rate
   - [ ] Shows subject breakdown

---

## Checkpoint Tasks

Congratulations on completing Beot! Final polish tasks:

- [ ] All sound notifications work
- [ ] Completion screen is celebratory
- [ ] Abandoned screen shows accountability
- [ ] Timer colour changes with urgency
- [ ] Menu looks polished
- [ ] **Challenge:** Add a settings screen to adjust timer duration
- [ ] **Challenge:** Add keyboard shortcut hints (1, 2, 3, 4 for menu items)
- [ ] **Challenge:** Add a "quick start" that remembers last subject
- [ ] **Challenge:** Add desktop notifications using a library like `beeep`

---

## Ideas for Future Enhancements

Now that you have a working app, here are ideas to extend it:

### Features
1. **Break Timer** - 5-minute short breaks, 15-minute long breaks every 4 sessions
2. **Custom Durations** - Let users set their preferred pomodoro length
3. **Session Notes** - Add notes when completing a session
4. **Tags** - Add tags to sessions for better categorization
5. **Export** - Export session history to CSV/JSON
6. **Themes** - Multiple colour schemes

### Technical
1. **Configuration File** - YAML/JSON config for settings
2. **Environment Variables** - MongoDB URI, default duration, etc.
3. **Tests** - Unit tests for streak calculation, integration tests for repos
4. **CI/CD** - GitHub Actions for building releases
5. **Cross-Platform Builds** - Release for Windows, macOS, Linux

### UI
1. **Responsive Layout** - Adapt to terminal size
2. **Animation** - Smooth transitions between views
3. **Help Screen** - Detailed keyboard shortcuts
4. **Session History View** - Browse past sessions

---

## What You've Built

You now have a fully functional terminal Pomodoro timer with:

- ✅ Bubble Tea TUI framework
- ✅ Multiple views with navigation
- ✅ MongoDB persistence
- ✅ Subject tracking
- ✅ Session recording (completed & abandoned)
- ✅ Streak calculation
- ✅ Motivational quotes
- ✅ Statistics dashboard
- ✅ Sound notifications
- ✅ Visual polish

**You kept your beot.** 🗡️

---

## Quick Reference - Complete API

```go
// Bubble Tea
tea.NewProgram(model, tea.WithAltScreen())
tea.Quit
tea.Batch(cmd1, cmd2)
tea.Tick(duration, func(t time.Time) tea.Msg { return msg })

// Lip Gloss
style := lipgloss.NewStyle().
    Bold(true).
    Italic(true).
    Foreground(lipgloss.Color("205")).
    Background(lipgloss.Color("0")).
    Padding(1, 2).
    Margin(1).
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("205")).
    Width(40).
    Align(lipgloss.Center)

output := style.Render("text")

// Borders
lipgloss.NormalBorder()
lipgloss.RoundedBorder()
lipgloss.DoubleBorder()
lipgloss.ThickBorder()

// MongoDB
collection.InsertOne(ctx, doc)
collection.Find(ctx, filter)
collection.FindOne(ctx, filter).Decode(&result)
collection.UpdateOne(ctx, filter, bson.M{"$set": update})
collection.DeleteOne(ctx, filter)
collection.CountDocuments(ctx, filter)
collection.Aggregate(ctx, pipeline)

// BSON
bson.M{"field": "value"}
bson.M{"field": bson.M{"$gte": value}}
bson.A{stage1, stage2, stage3}

// Sound
fmt.Print("\a")  // Terminal bell
```

---

## Resources

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lip Gloss Styling](https://github.com/charmbracelet/lipgloss)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)
- [Charm Blog](https://charm.sh/blog/) - Great tutorials

---

*"Some of the greatest innovations have come from people who only succeeded because they were too dumb to know that what they were doing was impossible."*

You've built something real. You kept your vow. Now go focus. 🍅🗡️
