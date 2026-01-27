# Lesson 6: Quotes System

In this lesson you'll add rotating motivational quotes that display during focus sessions, keeping you inspired throughout the pomodoro.

---

## What You'll Learn

- Running multiple tick commands simultaneously
- Rotating content on a timer
- Loading random data from MongoDB
- Managing quote display state

---

## The Quotes Feature

During a 25-minute focus session, we'll:
1. Display a random quote
2. Rotate to a new quote every 30 seconds
3. Show the quote source (if available)

This keeps the timer interesting and provides motivation.

---

## Exercise 6.1: Load Quotes for Timer

First, let's load quotes when starting a timer. Update `ui/app.go`:

Add a new message type and update the flow:

```go
// Add to the message types at the bottom
type quotesLoadedMsg struct {
	quotes []models.Quote
	err    error
}

// Add this method to AppModel
func (m AppModel) loadQuotes() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		quotes, err := m.quoteRepo.GetAll(ctx)
		if err != nil {
			return quotesLoadedMsg{err: err}
		}
		return quotesLoadedMsg{quotes: quotes}
	}
}
```

Update the SubjectSelectedMsg handler:

```go
case SubjectSelectedMsg:
	// Store the selected subject, then load quotes
	m.pendingSubject = msg.Subject  // Add this field to AppModel
	return m, m.loadQuotes()

case quotesLoadedMsg:
	quotes := msg.quotes
	if msg.err != nil || len(quotes) == 0 {
		// Fallback to empty quotes
		quotes = []models.Quote{}
	}
	m.timer = NewTimerModelWithQuotes(m.pendingSubject, 25, quotes)
	m.currentView = TimerViewState
	return m, m.timer.Init()
```

Add the `pendingSubject` field to AppModel:

```go
type AppModel struct {
	currentView    View
	menu           MenuModel
	subjectSelect  SubjectSelectModel
	timer          TimerModel
	pendingSubject models.Subject  // NEW

	// ... rest of fields
}
```

---

## Exercise 6.2: Update Timer Model for Quotes

Update `ui/timer.go` to handle quotes:

```go
package ui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"Beot/models"
)

// Message types
type tickMsg time.Time
type quoteRotateMsg time.Time

// TimerCompleteMsg is sent when the timer finishes
type TimerCompleteMsg struct {
	Completed bool
	Subject   models.Subject
	Duration  int
	StartedAt time.Time
}

// TimerModel handles the countdown
type TimerModel struct {
	subject          models.Subject
	totalSeconds     int
	remainingSeconds int
	running          bool
	progress         progress.Model
	confirming       bool
	startedAt        time.Time

	// Quotes
	quotes          []models.Quote
	currentQuoteIdx int
}

// NewTimerModel creates a timer without quotes (backward compatible)
func NewTimerModel(subject models.Subject, minutes int) TimerModel {
	return NewTimerModelWithQuotes(subject, minutes, nil)
}

// NewTimerModelWithQuotes creates a timer with quotes
func NewTimerModelWithQuotes(subject models.Subject, minutes int, quotes []models.Quote) TimerModel {
	seconds := minutes * 60
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40

	// Shuffle quotes for variety
	if len(quotes) > 0 {
		rand.Shuffle(len(quotes), func(i, j int) {
			quotes[i], quotes[j] = quotes[j], quotes[i]
		})
	}

	return TimerModel{
		subject:          subject,
		totalSeconds:     seconds,
		remainingSeconds: seconds,
		running:          false,
		progress:         prog,
		startedAt:        time.Now(),
		quotes:           quotes,
		currentQuoteIdx:  0,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func quoteRotateCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return quoteRotateMsg(t)
	})
}

func (m *TimerModel) Start() tea.Cmd {
	m.running = true
	m.startedAt = time.Now()

	// Start both the timer tick and quote rotation
	if len(m.quotes) > 1 {
		return tea.Batch(tickCmd(), quoteRotateCmd())
	}
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
					return TimerCompleteMsg{
						Completed: false,
						Subject:   m.subject,
						Duration:  m.totalSeconds / 60,
						StartedAt: m.startedAt,
					}
				}
			case "n", "esc":
				m.confirming = false
				m.running = true
				cmds := []tea.Cmd{tickCmd()}
				if len(m.quotes) > 1 {
					cmds = append(cmds, quoteRotateCmd())
				}
				return m, tea.Batch(cmds...)
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
				cmds := []tea.Cmd{tickCmd()}
				if len(m.quotes) > 1 {
					cmds = append(cmds, quoteRotateCmd())
				}
				return m, tea.Batch(cmds...)
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

	case quoteRotateMsg:
		if m.running && len(m.quotes) > 1 {
			m.currentQuoteIdx = (m.currentQuoteIdx + 1) % len(m.quotes)
			return m, quoteRotateCmd()
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

	// Subject
	subjectDisplay := SubtitleStyle.Render(fmt.Sprintf("%s %s", m.subject.Icon, m.subject.Name))

	status := "🍅 Focus Time"
	if !m.running {
		status = "⏸  Paused"
	}

	// Progress bar
	progressBar := m.progress.ViewAs(percent)

	// Quote
	quoteDisplay := m.renderQuote()

	help := HelpStyle.Render("space pause • q/esc abandon")

	return fmt.Sprintf(
		"\n  %s\n  %s\n\n  %s\n\n  %s\n\n%s\n  %s\n",
		subjectDisplay,
		HelpStyle.Render(status),
		progressBar,
		timeDisplay,
		quoteDisplay,
		help,
	)
}

func (m TimerModel) renderQuote() string {
	if len(m.quotes) == 0 {
		return ""
	}

	quote := m.quotes[m.currentQuoteIdx]

	// Style the quote
	quoteStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("243")).
		Width(50).
		Align(lipgloss.Center)

	sourceStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center)

	text := quoteStyle.Render(fmt.Sprintf("\"%s\"", quote.Text))

	if quote.Source != "" {
		source := sourceStyle.Render("— " + quote.Source)
		return fmt.Sprintf("\n%s\n%s\n", text, source)
	}

	return fmt.Sprintf("\n%s\n", text)
}

func (m TimerModel) renderConfirmation() string {
	subject := fmt.Sprintf("%s %s", m.subject.Icon, m.subject.Name)
	title := ErrorStyle.Render("Abandon " + subject + "?")
	message := "This will be logged as abandoned 💀"
	help := HelpStyle.Render("[y] yes, abandon • [n] no, continue")

	return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", title, message, help)
}

func (m TimerModel) renderComplete() string {
	subject := fmt.Sprintf("%s %s", m.subject.Icon, m.subject.Name)

	// Show a final quote on completion
	var finalQuote string
	if len(m.quotes) > 0 {
		q := m.quotes[m.currentQuoteIdx]
		finalQuote = fmt.Sprintf("\n\"%s\"", q.Text)
		if q.Source != "" {
			finalQuote += fmt.Sprintf("\n— %s", q.Source)
		}
	}

	content := fmt.Sprintf(
		"%s\n\n%s%s\n\n%s",
		SuccessStyle.Render("✅ Session Complete!"),
		subject,
		HelpStyle.Render(finalQuote),
		HelpStyle.Render("Press any key to continue"),
	)

	return "\n" + BoxStyle.Render(content) + "\n"
}
```

---

## Exercise 6.3: Add Import for models

Make sure `ui/timer.go` has the models import:

```go
import (
	"fmt"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"Beot/models"
)
```

---

## Exercise 6.4: Update Full app.go

Here's the complete updated `ui/app.go`:

```go
package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"Beot/db"
	"Beot/models"
)

type View int

const (
	MenuViewState View = iota
	SubjectSelectViewState
	TimerViewState
	StatsViewState
)

type AppModel struct {
	currentView    View
	menu           MenuModel
	subjectSelect  SubjectSelectModel
	timer          TimerModel
	pendingSubject models.Subject

	subjectRepo *db.SubjectRepo
	sessionRepo *db.SessionRepo
	quoteRepo   *db.QuoteRepo

	subjects []models.Subject
	streak   int
}

func NewAppModel(subjectRepo *db.SubjectRepo, sessionRepo *db.SessionRepo, quoteRepo *db.QuoteRepo) AppModel {
	return AppModel{
		currentView: MenuViewState,
		menu:        NewMenuModel(),
		subjectRepo: subjectRepo,
		sessionRepo: sessionRepo,
		quoteRepo:   quoteRepo,
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadSubjects(),
		m.loadStreak(),
	)
}

func (m AppModel) loadSubjects() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		subjects, err := m.subjectRepo.GetAll(ctx)
		if err != nil {
			return subjectsLoadedMsg{err: err}
		}
		return subjectsLoadedMsg{subjects: subjects}
	}
}

func (m AppModel) loadStreak() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		days, err := m.sessionRepo.GetDaysWithCompletedSessions(ctx, 365)
		if err != nil {
			return streakLoadedMsg{streak: 0}
		}

		streak := calculateStreak(days)
		return streakLoadedMsg{streak: streak}
	}
}

func (m AppModel) loadQuotes() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		quotes, err := m.quoteRepo.GetAll(ctx)
		if err != nil {
			return quotesLoadedMsg{err: err}
		}
		return quotesLoadedMsg{quotes: quotes}
	}
}

func calculateStreak(days []time.Time) int {
	if len(days) == 0 {
		return 0
	}

	today := time.Now().Truncate(24 * time.Hour)
	streak := 0
	expected := today

	for _, day := range days {
		day = day.Truncate(24 * time.Hour)

		if day.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else if day.Before(expected) {
			break
		}
	}

	return streak
}

func (m AppModel) saveSession(completed bool, subject models.Subject, duration int, startedAt time.Time) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status := models.StatusCompleted
		if !completed {
			status = models.StatusAbandoned
		}

		session := &models.Session{
			SubjectID:   subject.ID,
			SubjectName: subject.Name,
			SubjectIcon: subject.Icon,
			Duration:    duration,
			Status:      status,
			StartedAt:   startedAt,
			CompletedAt: time.Now(),
		}

		err := m.sessionRepo.Create(ctx, session)
		return sessionSavedMsg{err: err}
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case subjectsLoadedMsg:
		if msg.err == nil {
			m.subjects = msg.subjects
		}
		return m, nil

	case streakLoadedMsg:
		m.streak = msg.streak
		m.menu.SetStreak(m.streak)
		return m, nil

	case quotesLoadedMsg:
		quotes := msg.quotes
		if msg.err != nil || len(quotes) == 0 {
			quotes = []models.Quote{}
		}
		m.timer = NewTimerModelWithQuotes(m.pendingSubject, 25, quotes)
		m.currentView = TimerViewState
		return m, m.timer.Init()

	case sessionSavedMsg:
		return m, m.loadStreak()

	case MenuSelectionMsg:
		switch MenuChoice(msg) {
		case StartSession:
			m.subjectSelect = NewSubjectSelectModel(m.subjects)
			m.currentView = SubjectSelectViewState
			return m, nil
		case ViewStats:
			m.currentView = StatsViewState
			return m, nil
		case QuitApp:
			return m, tea.Quit
		}
		return m, nil

	case SubjectSelectedMsg:
		m.pendingSubject = msg.Subject
		return m, m.loadQuotes()

	case BackToMenuMsg:
		m.currentView = MenuViewState
		return m, nil

	case TimerCompleteMsg:
		m.currentView = MenuViewState
		return m, m.saveSession(msg.Completed, msg.Subject, msg.Duration, msg.StartedAt)
	}

	switch m.currentView {
	case MenuViewState:
		newMenu, cmd := m.menu.Update(msg)
		m.menu = newMenu.(MenuModel)
		return m, cmd

	case SubjectSelectViewState:
		newSelect, cmd := m.subjectSelect.Update(msg)
		m.subjectSelect = newSelect.(SubjectSelectModel)
		return m, cmd

	case TimerViewState:
		newTimer, cmd := m.timer.Update(msg)
		m.timer = newTimer.(TimerModel)
		return m, cmd

	case StatsViewState:
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
	case SubjectSelectViewState:
		return m.subjectSelect.View()
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

	streakText := fmt.Sprintf("Current Streak: %d days", m.streak)
	if m.streak == 0 {
		streakText = "No streak yet - complete a session today!"
	}

	help := HelpStyle.Render("esc/q back to menu")

	return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", title, streakText, help)
}

type subjectsLoadedMsg struct {
	subjects []models.Subject
	err      error
}

type streakLoadedMsg struct {
	streak int
}

type quotesLoadedMsg struct {
	quotes []models.Quote
	err    error
}

type sessionSavedMsg struct {
	err error
}
```

---

## Testing Quote Rotation

For testing, you can temporarily reduce the rotation interval:

```go
func quoteRotateCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {  // 5 seconds for testing
		return quoteRotateMsg(t)
	})
}
```

Remember to change it back to 30 seconds for production.

---

## Adding More Quotes

You can add quotes directly in MongoDB:

```javascript
use beot

db.quotes.insertOne({
    text: "The only way to do great work is to love what you do.",
    source: "Steve Jobs",
    created_at: new Date()
})

db.quotes.insertMany([
    {
        text: "Focus on being productive instead of busy.",
        source: "Tim Ferriss",
        created_at: new Date()
    },
    {
        text: "Your time is limited, don't waste it living someone else's life.",
        source: "Steve Jobs",
        created_at: new Date()
    }
])
```

---

## Checkpoint Tasks

Before moving to Lesson 7, make sure you can:

- [ ] Quotes display during the timer
- [ ] Quotes rotate every 30 seconds (or your test interval)
- [ ] Quote source shows when available
- [ ] Quotes are shuffled on each session
- [ ] Timer still works when no quotes exist
- [ ] Pausing stops quote rotation
- [ ] Resuming restarts quote rotation
- [ ] **Challenge:** Add a "next quote" key to manually advance
- [ ] **Challenge:** Show quote count indicator (e.g., "3/7")
- [ ] **Challenge:** Add a "favourite quote" feature that weights display

---

## Common Gotchas

### "Quotes never rotate"

Make sure you return `quoteRotateCmd()` after each rotation:

```go
case quoteRotateMsg:
    if m.running && len(m.quotes) > 1 {
        m.currentQuoteIdx = (m.currentQuoteIdx + 1) % len(m.quotes)
        return m, quoteRotateCmd()  // Schedule next rotation!
    }
```

### "Same quote always shows"

Check that you're shuffling:

```go
rand.Shuffle(len(quotes), func(i, j int) {
    quotes[i], quotes[j] = quotes[j], quotes[i]
})
```

### "Quote rotation continues when paused"

Only return the rotation command when running:

```go
if m.running && len(m.quotes) > 1 {
    return m, quoteRotateCmd()
}
```

### "Quotes don't load"

Check the flow:
1. SubjectSelectedMsg stores pendingSubject
2. SubjectSelectedMsg returns loadQuotes()
3. quotesLoadedMsg creates timer with quotes
4. quotesLoadedMsg returns timer.Init()

### "Layout breaks with long quotes"

Set a max width on the quote style:

```go
quoteStyle := lipgloss.NewStyle().
    Width(50).  // Constrain width
    Align(lipgloss.Center)
```

---

## What's Next

In **Lesson 7**, we'll build the statistics screen:
- Today's sessions with icons
- Completion rate
- Streak calculation details
- Session history

Quotes inspire. Stats motivate. Let's build them!

---

## Quick Reference

```go
// Multiple simultaneous ticks
func (m *TimerModel) Start() tea.Cmd {
    return tea.Batch(tickCmd(), quoteRotateCmd())
}

// Rotating through a slice
m.currentQuoteIdx = (m.currentQuoteIdx + 1) % len(m.quotes)

// Conditional command batch
cmds := []tea.Cmd{tickCmd()}
if len(m.quotes) > 1 {
    cmds = append(cmds, quoteRotateCmd())
}
return m, tea.Batch(cmds...)

// Shuffle slice
rand.Shuffle(len(items), func(i, j int) {
    items[i], items[j] = items[j], items[i]
})

// Text wrapping with lipgloss
style := lipgloss.NewStyle().Width(50)
wrapped := style.Render(longText)
```

---

*"If you aren't dropping, you aren't learning. And if you aren't learning, you aren't a juggler."*

The quotes rotate. Wisdom cycles. Keep focusing. 💬
