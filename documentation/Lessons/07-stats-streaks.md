# Lesson 7: Stats & Streaks

In this lesson you'll build a comprehensive statistics screen showing today's sessions, overall progress, and streak information.

---

## What You'll Learn

- Aggregating data from MongoDB
- Building complex UI layouts with Lip Gloss
- Streak calculation logic
- Displaying session history

---

## The Stats Screen

We'll show:
1. Current streak with visual indicator
2. Today's sessions as icons
3. Total completed vs abandoned
4. Completion rate percentage
5. Recent session history

---

## Exercise 7.1: Enhanced Session Repository

First, let's add more query methods to `db/sessions.go`:

```go
// Add these methods to SessionRepo

// GetTodayStats returns today's session counts
func (r *SessionRepo) GetTodayStats(ctx context.Context) (completed, abandoned int, err error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	completed64, err := r.db.Collection(SessionsCollection).CountDocuments(ctx, bson.M{
		"started_at": bson.M{"$gte": startOfDay},
		"status":     models.StatusCompleted,
	})
	if err != nil {
		return 0, 0, err
	}

	abandoned64, err := r.db.Collection(SessionsCollection).CountDocuments(ctx, bson.M{
		"started_at": bson.M{"$gte": startOfDay},
		"status":     models.StatusAbandoned,
	})
	if err != nil {
		return 0, 0, err
	}

	return int(completed64), int(abandoned64), nil
}

// GetSubjectStats returns session counts grouped by subject
func (r *SessionRepo) GetSubjectStats(ctx context.Context) ([]SubjectStat, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"status": models.StatusCompleted}},
		bson.M{"$group": bson.M{
			"_id":   "$subject_name",
			"icon":  bson.M{"$first": "$subject_icon"},
			"count": bson.M{"$sum": 1},
		}},
		bson.M{"$sort": bson.M{"count": -1}},
	}

	cursor, err := r.db.Collection(SessionsCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []SubjectStat
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// SubjectStat holds aggregated subject data
type SubjectStat struct {
	Name  string `bson:"_id"`
	Icon  string `bson:"icon"`
	Count int    `bson:"count"`
}

// GetTotalMinutes returns total completed focus time
func (r *SessionRepo) GetTotalMinutes(ctx context.Context) (int, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"status": models.StatusCompleted}},
		bson.M{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$duration"},
		}},
	}

	cursor, err := r.db.Collection(SessionsCollection).Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total int `bson:"total"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	return results[0].Total, nil
}
```

---

## Exercise 7.2: Streak Calculator

Create `internal/streak/calculator.go`:

```go
package streak

import "time"

// Calculate returns the current streak based on days with completed sessions
func Calculate(days []time.Time) int {
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
			// Found a gap - streak broken
			break
		}
		// day.After(expected) means duplicate day, skip it
	}

	return streak
}

// LongestStreak calculates the longest streak from a list of dates
func LongestStreak(days []time.Time) int {
	if len(days) == 0 {
		return 0
	}

	longest := 0
	current := 1

	for i := 1; i < len(days); i++ {
		prev := days[i-1].Truncate(24 * time.Hour)
		curr := days[i].Truncate(24 * time.Hour)

		// Check if consecutive (curr is one day before prev since sorted desc)
		expectedPrev := curr.AddDate(0, 0, 1)

		if prev.Equal(expectedPrev) {
			current++
		} else if !prev.Equal(curr) {
			// Not consecutive and not same day
			if current > longest {
				longest = current
			}
			current = 1
		}
		// Same day - don't increment or reset
	}

	if current > longest {
		longest = current
	}

	return longest
}

// StreakInfo contains detailed streak information
type StreakInfo struct {
	Current int
	Longest int
	Total   int // Total days with sessions
}

// GetStreakInfo calculates comprehensive streak data
func GetStreakInfo(days []time.Time) StreakInfo {
	return StreakInfo{
		Current: Calculate(days),
		Longest: LongestStreak(days),
		Total:   len(days),
	}
}
```

Create the directory:

```bash
mkdir -p internal/streak
```

---

## Exercise 7.3: Stats Model

Create `ui/stats.go`:

```go
package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"Beot/db"
	"Beot/internal/streak"
	"Beot/models"
)

// StatsModel handles the statistics view
type StatsModel struct {
	sessionRepo *db.SessionRepo

	// Data
	streak        streak.StreakInfo
	todaySessions []models.Session
	todayStats    struct{ completed, abandoned int }
	totalStats    struct{ completed, abandoned int64 }
	subjectStats  []db.SubjectStat
	totalMinutes  int

	// State
	loading bool
	err     error
}

// NewStatsModel creates a new stats view
func NewStatsModel(sessionRepo *db.SessionRepo) StatsModel {
	return StatsModel{
		sessionRepo: sessionRepo,
		loading:     true,
	}
}

func (m StatsModel) Init() tea.Cmd {
	return m.loadStats()
}

func (m StatsModel) loadStats() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var data statsLoadedMsg

		// Load today's sessions
		sessions, err := m.sessionRepo.GetToday(ctx)
		if err != nil {
			data.err = err
			return data
		}
		data.todaySessions = sessions

		// Load today's stats
		completed, abandoned, err := m.sessionRepo.GetTodayStats(ctx)
		if err != nil {
			data.err = err
			return data
		}
		data.todayCompleted = completed
		data.todayAbandoned = abandoned

		// Load total stats
		totalCompleted, totalAbandoned, err := m.sessionRepo.GetStats(ctx)
		if err != nil {
			data.err = err
			return data
		}
		data.totalCompleted = totalCompleted
		data.totalAbandoned = totalAbandoned

		// Load streak data
		days, err := m.sessionRepo.GetDaysWithCompletedSessions(ctx, 365)
		if err != nil {
			data.err = err
			return data
		}
		data.streakInfo = streak.GetStreakInfo(days)

		// Load subject stats
		subjectStats, err := m.sessionRepo.GetSubjectStats(ctx)
		if err != nil {
			data.err = err
			return data
		}
		data.subjectStats = subjectStats

		// Load total minutes
		totalMinutes, err := m.sessionRepo.GetTotalMinutes(ctx)
		if err != nil {
			data.err = err
			return data
		}
		data.totalMinutes = totalMinutes

		return data
	}
}

func (m StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case "r":
			// Refresh stats
			m.loading = true
			return m, m.loadStats()
		}

	case statsLoadedMsg:
		m.loading = false
		m.err = msg.err

		if msg.err == nil {
			m.todaySessions = msg.todaySessions
			m.todayStats.completed = msg.todayCompleted
			m.todayStats.abandoned = msg.todayAbandoned
			m.totalStats.completed = msg.totalCompleted
			m.totalStats.abandoned = msg.totalAbandoned
			m.streak = msg.streakInfo
			m.subjectStats = msg.subjectStats
			m.totalMinutes = msg.totalMinutes
		}
	}

	return m, nil
}

func (m StatsModel) View() string {
	if m.loading {
		return "\n  Loading statistics...\n"
	}

	if m.err != nil {
		return fmt.Sprintf("\n  %s\n\n  %s\n",
			ErrorStyle.Render("Error loading stats"),
			HelpStyle.Render("esc back • r retry"),
		)
	}

	var sections []string

	// Title
	sections = append(sections, TitleStyle.Render("📊 Statistics"))

	// Streak section
	sections = append(sections, m.renderStreak())

	// Today section
	sections = append(sections, m.renderToday())

	// Overall section
	sections = append(sections, m.renderOverall())

	// Subject breakdown
	if len(m.subjectStats) > 0 {
		sections = append(sections, m.renderSubjects())
	}

	// Help
	sections = append(sections, HelpStyle.Render("esc/q back • r refresh"))

	return "\n  " + strings.Join(sections, "\n\n  ") + "\n"
}

func (m StatsModel) renderStreak() string {
	var streakDisplay string

	if m.streak.Current > 0 {
		// Show fire emojis based on streak length
		fires := "🔥"
		if m.streak.Current >= 7 {
			fires = "🔥🔥"
		}
		if m.streak.Current >= 30 {
			fires = "🔥🔥🔥"
		}

		streakDisplay = SuccessStyle.Render(fmt.Sprintf("%s %d day streak!", fires, m.streak.Current))
	} else {
		streakDisplay = HelpStyle.Render("No current streak - complete a session today!")
	}

	details := HelpStyle.Render(fmt.Sprintf(
		"Longest: %d days • Total active days: %d",
		m.streak.Longest,
		m.streak.Total,
	))

	return fmt.Sprintf("%s\n  %s", streakDisplay, details)
}

func (m StatsModel) renderToday() string {
	title := SubtitleStyle.Render("Today")

	if len(m.todaySessions) == 0 {
		return fmt.Sprintf("%s\n  %s", title, HelpStyle.Render("No sessions yet today"))
	}

	// Show session icons
	var icons string
	for _, s := range m.todaySessions {
		if s.Status == models.StatusCompleted {
			icons += s.SubjectIcon + " "
		} else {
			icons += "💀 "
		}
	}

	summary := fmt.Sprintf("%d completed", m.todayStats.completed)
	if m.todayStats.abandoned > 0 {
		summary += fmt.Sprintf(", %d abandoned", m.todayStats.abandoned)
	}

	return fmt.Sprintf("%s\n  %s\n  %s", title, icons, HelpStyle.Render(summary))
}

func (m StatsModel) renderOverall() string {
	title := SubtitleStyle.Render("All Time")

	total := m.totalStats.completed + m.totalStats.abandoned
	if total == 0 {
		return fmt.Sprintf("%s\n  %s", title, HelpStyle.Render("No sessions recorded"))
	}

	rate := float64(m.totalStats.completed) / float64(total) * 100

	// Completion rate bar
	barWidth := 20
	filledWidth := int(rate / 100 * float64(barWidth))
	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

	rateStyle := SuccessStyle
	if rate < 70 {
		rateStyle = WarningStyle
	}
	if rate < 50 {
		rateStyle = ErrorStyle
	}

	rateDisplay := rateStyle.Render(fmt.Sprintf("%s %.0f%%", bar, rate))

	// Total time
	hours := m.totalMinutes / 60
	mins := m.totalMinutes % 60
	timeDisplay := fmt.Sprintf("Total focus time: %dh %dm", hours, mins)

	sessionsDisplay := fmt.Sprintf(
		"Sessions: %d completed, %d abandoned",
		m.totalStats.completed,
		m.totalStats.abandoned,
	)

	return fmt.Sprintf("%s\n  %s\n  %s\n  %s",
		title,
		rateDisplay,
		HelpStyle.Render(sessionsDisplay),
		HelpStyle.Render(timeDisplay),
	)
}

func (m StatsModel) renderSubjects() string {
	title := SubtitleStyle.Render("By Subject")

	var lines []string
	for _, stat := range m.subjectStats {
		line := fmt.Sprintf("%s %s: %d sessions", stat.Icon, stat.Name, stat.Count)
		lines = append(lines, line)
	}

	return fmt.Sprintf("%s\n  %s", title, HelpStyle.Render(strings.Join(lines, "\n  ")))
}

// statsLoadedMsg contains all loaded stats data
type statsLoadedMsg struct {
	todaySessions  []models.Session
	todayCompleted int
	todayAbandoned int
	totalCompleted int64
	totalAbandoned int64
	streakInfo     streak.StreakInfo
	subjectStats   []db.SubjectStat
	totalMinutes   int
	err            error
}
```

---

## Exercise 7.4: Integrate Stats into App

Update `ui/app.go` to use the new StatsModel:

```go
// Add to imports
import "Beot/internal/streak"

// Update AppModel struct
type AppModel struct {
	currentView    View
	menu           MenuModel
	subjectSelect  SubjectSelectModel
	timer          TimerModel
	stats          StatsModel  // NEW
	pendingSubject models.Subject

	subjectRepo *db.SubjectRepo
	sessionRepo *db.SessionRepo
	quoteRepo   *db.QuoteRepo

	subjects []models.Subject
	streak   int
}

// Update the MenuSelectionMsg handler
case MenuSelectionMsg:
    switch MenuChoice(msg) {
    case StartSession:
        m.subjectSelect = NewSubjectSelectModel(m.subjects)
        m.currentView = SubjectSelectViewState
        return m, nil
    case ViewStats:
        m.stats = NewStatsModel(m.sessionRepo)
        m.currentView = StatsViewState
        return m, m.stats.Init()  // Load stats data
    case QuitApp:
        return m, tea.Quit
    }
    return m, nil

// Update the view routing in Update()
case StatsViewState:
    newStats, cmd := m.stats.Update(msg)
    m.stats = newStats.(StatsModel)
    return m, cmd

// Update View()
case StatsViewState:
    return m.stats.View()
```

---

## Exercise 7.5: Fix Import in app.go

Make sure the streak import is correct:

```go
import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"Beot/db"
	"Beot/models"
	// Note: We don't need to import streak in app.go anymore
	// since StatsModel handles it internally
)
```

Actually, we can remove the streak calculation from app.go since stats.go handles it. Just keep the simple version for the menu streak display:

```go
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
```

---

## Testing the Stats

Run the app and complete a few sessions:

```bash
go run main.go
```

1. Complete some sessions with different subjects
2. Abandon at least one session
3. Go to "View Statistics"

You should see:
- Your current streak
- Today's session icons
- Completion rate with a progress bar
- Total focus time
- Breakdown by subject

---

## Checkpoint Tasks

Before moving to Lesson 8, make sure you can:

- [ ] Stats screen loads and displays data
- [ ] Current streak shows correctly
- [ ] Today's sessions show as icons
- [ ] Abandoned sessions show as 💀
- [ ] Completion rate bar displays
- [ ] Subject breakdown shows counts
- [ ] 'r' refreshes the stats
- [ ] ESC returns to menu
- [ ] **Challenge:** Add a "This Week" section
- [ ] **Challenge:** Show best day (most sessions completed)
- [ ] **Challenge:** Add streak milestones (7 days, 30 days, etc.)

---

## Common Gotchas

### "Stats don't load"

Make sure you return `m.stats.Init()` when switching to stats view:

```go
case ViewStats:
    m.stats = NewStatsModel(m.sessionRepo)
    m.currentView = StatsViewState
    return m, m.stats.Init()  // This loads the data!
```

### "Streak is always 0"

Check that:
1. Sessions are saved with "completed" status
2. GetDaysWithCompletedSessions filters correctly
3. Date truncation handles timezone

### "Type assertion fails for StatsModel"

Make sure StatsModel.Update returns the correct type:

```go
func (m StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ...
    return m, nil  // Returns StatsModel
}
```

### "Subject stats empty"

Verify the aggregation pipeline matches your data:

```javascript
// Test in mongosh
use beot
db.sessions.aggregate([
    { $match: { status: "completed" } },
    { $group: { _id: "$subject_name", count: { $sum: 1 } } }
])
```

### "Completion rate shows NaN"

Handle the zero-total case:

```go
total := m.totalStats.completed + m.totalStats.abandoned
if total == 0 {
    return "No sessions recorded"
}
rate := float64(m.totalStats.completed) / float64(total) * 100
```

---

## What's Next

In **Lesson 8**, we'll add final polish:
- Terminal bell on completion
- Completion celebration screen
- Sound effects (optional)
- Final touches

The data tells the story. Let's make it celebratory!

---

## Quick Reference

```go
// MongoDB aggregation pipeline
pipeline := bson.A{
    bson.M{"$match": bson.M{"status": "completed"}},
    bson.M{"$group": bson.M{
        "_id":   "$field",
        "count": bson.M{"$sum": 1},
    }},
    bson.M{"$sort": bson.M{"count": -1}},
}

// Streak calculation
today := time.Now().Truncate(24 * time.Hour)
expected := today
for _, day := range days {
    if day.Truncate(24*time.Hour).Equal(expected) {
        streak++
        expected = expected.AddDate(0, 0, -1)
    } else {
        break
    }
}

// Progress bar with characters
barWidth := 20
filled := int(percent / 100 * float64(barWidth))
bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

// Conditional styling
style := SuccessStyle
if value < threshold {
    style = WarningStyle
}
```

---

*"Game design is decision making, and decisions must be made with confidence."*

The stats reveal the truth. The streak holds. Keep going. 📊
