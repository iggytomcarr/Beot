# Lesson 5: Subjects & Sessions

In this lesson you'll add subject selection before starting a timer and save completed/abandoned sessions to the database.

---

## What You'll Learn

- Passing data between views
- Injecting database repositories into the UI
- Creating a subject selection screen
- Saving session data on completion
- Displaying dynamic data from the database

---

## The Data Flow

Here's what we're building:

```
Menu → Select Subject → Timer (with subject) → Save Session → Menu
```

We need to:
1. Load subjects from database
2. Let user select one
3. Pass it to the timer
4. Save the session when timer completes

---

## Exercise 5.1: Subject Selection View

Create `ui/subject_select.go`:

```go
package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"Beot/models"
)

// SubjectSelectModel handles subject selection
type SubjectSelectModel struct {
	subjects []models.Subject
	cursor   int
}

// NewSubjectSelectModel creates a subject selector
func NewSubjectSelectModel(subjects []models.Subject) SubjectSelectModel {
	return SubjectSelectModel{
		subjects: subjects,
		cursor:   0,
	}
}

func (m SubjectSelectModel) Init() tea.Cmd {
	return nil
}

func (m SubjectSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.subjects)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.subjects) > 0 {
				return m, func() tea.Msg {
					return SubjectSelectedMsg{Subject: m.subjects[m.cursor]}
				}
			}
		case "esc", "q":
			return m, func() tea.Msg {
				return BackToMenuMsg{}
			}
		}
	}
	return m, nil
}

func (m SubjectSelectModel) View() string {
	title := TitleStyle.Render("What are you focusing on?")

	if len(m.subjects) == 0 {
		return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n",
			title,
			ErrorStyle.Render("No subjects found!"),
			HelpStyle.Render("esc back"),
		)
	}

	var items string
	for i, s := range m.subjects {
		cursor := "  "
		style := NormalStyle

		if m.cursor == i {
			cursor = "▸ "
			style = SelectedStyle
		}

		// Apply the subject's colour
		subjectStyle := style.Copy().Foreground(lipgloss.Color(s.Color))
		items += fmt.Sprintf("%s%s %s\n", cursor, s.Icon, subjectStyle.Render(s.Name))
	}

	help := HelpStyle.Render("↑/↓ navigate • enter select • esc back")

	return fmt.Sprintf("\n  %s\n\n%s\n  %s\n", title, items, help)
}

// SubjectSelectedMsg is sent when a subject is chosen
type SubjectSelectedMsg struct {
	Subject models.Subject
}

// BackToMenuMsg is sent when user wants to go back
type BackToMenuMsg struct{}
```

---

## Exercise 5.2: Update Timer to Hold Subject

Update `ui/timer.go` to include the subject:

```go
package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"Beot/models"
)

type tickMsg time.Time

// TimerCompleteMsg is sent when the timer finishes
type TimerCompleteMsg struct {
	Completed bool
	Subject   models.Subject
	Duration  int       // minutes
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
}

// NewTimerModel creates a timer for the given subject and duration
func NewTimerModel(subject models.Subject, minutes int) TimerModel {
	seconds := minutes * 60
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40

	return TimerModel{
		subject:          subject,
		totalSeconds:     seconds,
		remainingSeconds: seconds,
		running:          false,
		progress:         prog,
		startedAt:        time.Now(),
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *TimerModel) Start() tea.Cmd {
	m.running = true
	m.startedAt = time.Now()
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

	// Show subject being worked on
	subjectDisplay := SubtitleStyle.Render(fmt.Sprintf("%s %s", m.subject.Icon, m.subject.Name))

	status := "🍅 Focus Time"
	if !m.running {
		status = "⏸  Paused"
	}

	progressBar := m.progress.ViewAs(percent)
	help := HelpStyle.Render("space pause • q/esc abandon")

	return fmt.Sprintf(
		"\n  %s\n  %s\n\n  %s\n\n  %s\n\n  %s\n",
		subjectDisplay,
		HelpStyle.Render(status),
		progressBar,
		timeDisplay,
		help,
	)
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
	content := fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		SuccessStyle.Render("✅ Session Complete!"),
		subject,
		HelpStyle.Render("Press any key to continue"),
	)

	return "\n" + BoxStyle.Render(content) + "\n"
}
```

---

## Exercise 5.3: Update App Container

Update `ui/app.go` to handle the full flow and database:

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

// AppModel is the main application container
type AppModel struct {
	currentView   View
	menu          MenuModel
	subjectSelect SubjectSelectModel
	timer         TimerModel

	// Database repositories
	subjectRepo *db.SubjectRepo
	sessionRepo *db.SessionRepo
	quoteRepo   *db.QuoteRepo

	// Cached data
	subjects []models.Subject
	streak   int
}

// NewAppModel creates the application with database access
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
	// Load initial data
	return tea.Batch(
		m.loadSubjects(),
		m.loadStreak(),
	)
}

// loadSubjects fetches subjects from the database
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

// loadStreak calculates the current streak
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

// calculateStreak counts consecutive days with sessions
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
			// Gap in streak
			break
		}
	}

	return streak
}

// saveSession persists a session to the database
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

	// Data loaded messages
	case subjectsLoadedMsg:
		if msg.err == nil {
			m.subjects = msg.subjects
		}
		return m, nil

	case streakLoadedMsg:
		m.streak = msg.streak
		m.menu.SetStreak(m.streak)
		return m, nil

	case sessionSavedMsg:
		// Session saved, reload streak
		return m, m.loadStreak()

	// Navigation messages
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
		m.timer = NewTimerModel(msg.Subject, 25)
		m.currentView = TimerViewState
		return m, m.timer.Init()

	case BackToMenuMsg:
		m.currentView = MenuViewState
		return m, nil

	case TimerCompleteMsg:
		m.currentView = MenuViewState
		// Save the session
		return m, m.saveSession(msg.Completed, msg.Subject, msg.Duration, msg.StartedAt)
	}

	// Route to current view
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

// Internal messages
type subjectsLoadedMsg struct {
	subjects []models.Subject
	err      error
}

type streakLoadedMsg struct {
	streak int
}

type sessionSavedMsg struct {
	err error
}
```

---

## Exercise 5.4: Update main.go

Update `main.go` to pass repositories:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"Beot/db"
	"Beot/ui"
)

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.Connect("mongodb://localhost:27017", "beot")
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		fmt.Println("Make sure MongoDB is running:")
		fmt.Println("  docker run -d -p 27017:27017 --name beot-mongo mongo:latest")
		os.Exit(1)
	}
	defer database.Disconnect()

	// Create repositories
	quoteRepo := db.NewQuoteRepo(database)
	subjectRepo := db.NewSubjectRepo(database)
	sessionRepo := db.NewSessionRepo(database)

	// Seed default data
	if err := quoteRepo.SeedDefaults(ctx); err != nil {
		fmt.Printf("Warning: Failed to seed quotes: %v\n", err)
	}

	if err := subjectRepo.SeedDefaults(ctx); err != nil {
		fmt.Printf("Warning: Failed to seed subjects: %v\n", err)
	}

	// Create app with database access
	app := ui.NewAppModel(subjectRepo, sessionRepo, quoteRepo)

	// Run the TUI
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
```

---

## Exercise 5.5: Test the Flow

Run the app and test:

```bash
go run main.go
```

1. Select "Start Focus Session"
2. Choose a subject (e.g., GoLang)
3. Watch the timer (set to 25 minutes, or temporarily change to 1 for testing)
4. Either complete or abandon
5. Check the database:

```bash
docker exec -it beot-mongo mongosh
```

```javascript
use beot
db.sessions.find().pretty()
```

You should see your session with status "completed" or "abandoned".

---

## Understanding Async Commands

Commands that talk to the database are async:

```go
func (m AppModel) loadSubjects() tea.Cmd {
    return func() tea.Msg {
        // This runs in a goroutine
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        subjects, err := m.subjectRepo.GetAll(ctx)
        if err != nil {
            return subjectsLoadedMsg{err: err}
        }
        return subjectsLoadedMsg{subjects: subjects}
    }
}
```

The pattern:
1. Return a `func() tea.Msg`
2. Inside, do your async work
3. Return a message with the result
4. Handle that message in Update()

This keeps the UI responsive while database operations happen.

---

## Checkpoint Tasks

Before moving to Lesson 6, make sure you can:

- [ ] Subject selection screen shows all subjects
- [ ] Subjects display with correct icons and colours
- [ ] Selecting a subject starts the timer
- [ ] Timer shows the selected subject
- [ ] Completing saves a "completed" session
- [ ] Abandoning saves an "abandoned" session
- [ ] Sessions appear in MongoDB
- [ ] Streak updates after completing a session
- [ ] **Challenge:** Add keyboard shortcuts (1-6) for subject selection
- [ ] **Challenge:** Show total sessions count on the stats screen
- [ ] **Challenge:** Add an "Add Subject" option to the subject select screen

---

## Common Gotchas

### "Subjects don't load"

Make sure Init() returns the loadSubjects command:

```go
func (m AppModel) Init() tea.Cmd {
    return tea.Batch(
        m.loadSubjects(),
        m.loadStreak(),
    )
}
```

### "Session not saving"

Check that TimerCompleteMsg includes all required data:

```go
return TimerCompleteMsg{
    Completed: true,
    Subject:   m.subject,
    Duration:  m.totalSeconds / 60,
    StartedAt: m.startedAt,
}
```

### "Subject colours not showing"

Make sure you're using lipgloss.Color with the hex string:

```go
subjectStyle := style.Copy().Foreground(lipgloss.Color(s.Color))
```

### "Streak always 0"

Check that:
1. Sessions are being saved with status "completed"
2. GetDaysWithCompletedSessions filters by status
3. The date comparison handles timezones correctly

### "Type assertion panic"

When updating views, make sure the types match:

```go
newSelect, cmd := m.subjectSelect.Update(msg)
m.subjectSelect = newSelect.(SubjectSelectModel)  // Must be SubjectSelectModel
```

---

## What's Next

In **Lesson 6**, we'll add motivational quotes:
- Rotating quotes during the timer
- Random quote selection
- Quote management

Sessions are being tracked. Time to add some inspiration!

---

## Quick Reference

```go
// Async command pattern
func (m AppModel) doSomething() tea.Cmd {
    return func() tea.Msg {
        // Async work here
        result, err := someOperation()
        return resultMsg{data: result, err: err}
    }
}

// Handle in Update:
case resultMsg:
    if msg.err != nil {
        // Handle error
    }
    m.data = msg.data
    return m, nil

// Batch multiple commands
return m, tea.Batch(cmd1, cmd2, cmd3)

// Pass data through message
type SubjectSelectedMsg struct {
    Subject models.Subject
}

return m, func() tea.Msg {
    return SubjectSelectedMsg{Subject: selected}
}
```

---

*"Some of the greatest innovations have come from people who only succeeded because they were too dumb to know that what they were doing was impossible."*

Sessions save. Subjects selected. The vow is being kept. 📝
