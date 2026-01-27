# Beot: A Pomodoro CLI Tool in Go

A comprehensive guide to building a terminal-based Pomodoro timer with Bubble Tea, MongoDB, and accountability features.

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Prerequisites & Setup](#2-prerequisites--setup)
3. [Phase 1: Hello Bubble Tea](#3-phase-1-hello-bubble-tea)
4. [Phase 2: Building the Timer](#4-phase-2-building-the-timer)
5. [Phase 3: Multiple Views & Navigation](#5-phase-3-multiple-views--navigation)
6. [Phase 4: MongoDB Integration](#6-phase-4-mongodb-integration)
7. [Phase 5: Subjects & Sessions](#7-phase-5-subjects--sessions)
8. [Phase 6: Quotes System](#8-phase-6-quotes-system)
9. [Phase 7: Stats & Streaks](#9-phase-7-stats--streaks)
10. [Phase 8: Polish & Sound](#10-phase-8-polish--sound)
11. [Challenges & Extensions](#11-challenges--extensions)
12. [Resources](#12-resources)

---

## 1. Project Overview

### What You're Building

A terminal Pomodoro timer that:
- Runs 25-minute focus sessions tied to subjects (GoLang, Music, etc.)
- Displays rotating motivational quotes during sessions
- Tracks completed AND abandoned sessions (accountability!)
- Shows streaks and statistics
- Persists everything to MongoDB

### What You'll Learn

- **Go fundamentals**: structs, interfaces, error handling, packages
- **Bubble Tea**: The Elm architecture (Model-Update-View), commands, messages
- **Lip Gloss**: Terminal styling
- **MongoDB Go Driver**: CRUD operations, connections, queries
- **Project structure**: Organising a real Go application

### The Elm Architecture (Core Concept)

Bubble Tea uses the Elm architecture. Understand this and everything else clicks:

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

1. **Model**: Your application state (a struct)
2. **Update**: Receives messages, returns new state + optional commands
3. **View**: Renders the model as a string for the terminal

State flows one direction. Messages trigger changes. The view is just a function of state.

---

## 2. Prerequisites & Setup

### Install Go

```bash
go version  # Should be 1.21+
```

### Install MongoDB

**Option A: Local (macOS)**
```bash
brew tap mongodb/brew
brew install mongodb-community
brew services start mongodb-community
```

**Option B: Docker**
```bash
docker run -d -p 27017:27017 --name beot-mongo mongo:latest
```

### Create Your Project

```bash
mkdir beot && cd beot
go mod init beot

go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles/progress
go get github.com/charmbracelet/bubbles/textinput
go get go.mongodb.org/mongo-driver/mongo
```

### Project Structure

```
beot/
├── main.go
├── ui/
│   ├── app.go
│   ├── menu.go
│   ├── timer.go
│   ├── subject_select.go
│   ├── stats.go
│   ├── quotes.go
│   └── styles.go
├── db/
│   ├── mongo.go
│   ├── sessions.go
│   ├── quotes.go
│   └── subjects.go
├── models/
│   ├── session.go
│   ├── subject.go
│   └── quote.go
└── internal/
    └── streak/
        └── calculator.go
```

---

## 3. Phase 1: Hello Bubble Tea

**Goal**: Understand the basic Bubble Tea pattern.

### Exercise 1.1: The Simplest Program

Create `main.go`:

```go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    message string
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
        }
    }
    return m, nil
}

func (m model) View() string {
    return fmt.Sprintf("\n  %s\n\n  Press 'q' to quit.\n", m.message)
}

func main() {
    p := tea.NewProgram(model{message: "Hello, Bubble Tea! 🧋"})
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}
```

**Run it:** `go run main.go`

**🎯 Learning points:**
- `model` struct holds ALL your state
- `Init()` runs once at startup
- `Update()` receives messages, returns new state
- `View()` returns a string to display

### Exercise 1.2: Adding Interactivity

```go
type model struct {
    counter int
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
        }
    }
    return m, nil
}

func (m model) View() string {
    return fmt.Sprintf("\n  Counter: %d\n\n  ↑/k up • ↓/j down • q quit\n", m.counter)
}
```

### Exercise 1.3: Adding Lip Gloss Styling

```go
import "github.com/charmbracelet/lipgloss"

var (
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("205"))
    
    helpStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("241"))
)

func (m model) View() string {
    title := titleStyle.Render("✨ My Counter")
    help := helpStyle.Render("↑/k up • ↓/j down • q quit")
    return fmt.Sprintf("\n  %s\n\n  Counter: %d\n\n  %s\n", title, m.counter, help)
}
```

### 📝 Checkpoint Tasks

- [ ] Run the counter and verify keys work
- [ ] Add a key to reset counter to 0
- [ ] Change the colours
- [ ] Add a border using `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())`

---

## 4. Phase 2: Building the Timer

**Goal**: Create a countdown timer with tick-based updates.

### Understanding Commands

Things happen via **messages**. **Commands** produce messages:

```go
// tea.Tick sends a message after duration
tea.Tick(time.Second, func(t time.Time) tea.Msg {
    return tickMsg(t)
})
```

### Exercise 2.1: Basic Countdown

Create `ui/timer.go`:

```go
package ui

import (
    "fmt"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type TimerModel struct {
    totalSeconds     int
    remainingSeconds int
    running          bool
}

func NewTimerModel(minutes int) TimerModel {
    seconds := minutes * 60
    return TimerModel{
        totalSeconds:     seconds,
        remainingSeconds: seconds,
        running:          false,
    }
}

func (m *TimerModel) Start() tea.Cmd {
    m.running = true
    return tickCmd()
}

func tickCmd() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func (m TimerModel) Init() tea.Cmd {
    return m.Start()
}

func (m TimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" || msg.String() == "ctrl+c" {
            return m, tea.Quit
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
    }
    return m, nil
}

func (m TimerModel) View() string {
    minutes := m.remainingSeconds / 60
    seconds := m.remainingSeconds % 60
    timeDisplay := fmt.Sprintf("%02d:%02d", minutes, seconds)
    
    status := "🍅 Focus Time"
    if !m.running && m.remainingSeconds <= 0 {
        status = "✅ Complete!"
    }
    
    return fmt.Sprintf("\n  %s\n\n  %s\n\n  Press 'q' to quit\n", status, timeDisplay)
}
```

### Exercise 2.2: Adding Progress Bar

```go
import "github.com/charmbracelet/bubbles/progress"

type TimerModel struct {
    totalSeconds     int
    remainingSeconds int
    running          bool
    progress         progress.Model
}

func NewTimerModel(minutes int) TimerModel {
    seconds := minutes * 60
    prog := progress.New(progress.WithDefaultGradient())
    prog.Width = 40
    
    return TimerModel{
        totalSeconds:     seconds,
        remainingSeconds: seconds,
        progress:         prog,
    }
}

func (m TimerModel) View() string {
    elapsed := m.totalSeconds - m.remainingSeconds
    percent := float64(elapsed) / float64(m.totalSeconds)
    
    minutes := m.remainingSeconds / 60
    seconds := m.remainingSeconds % 60
    timeDisplay := fmt.Sprintf("%02d:%02d", minutes, seconds)
    
    return fmt.Sprintf("\n  🍅 Focus\n\n  %s  %s\n\n  'q' to quit\n",
        m.progress.ViewAs(percent),
        timeDisplay,
    )
}
```

### Exercise 2.3: Abandon Confirmation

```go
type TimerModel struct {
    // ... existing fields
    confirming bool
}

func (m TimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.confirming {
            switch msg.String() {
            case "y":
                return m, tea.Quit  // Later: save as abandoned
            case "n", "esc":
                m.confirming = false
                return m, tickCmd()
            }
            return m, nil
        }
        
        switch msg.String() {
        case "q":
            m.confirming = true
            m.running = false
            return m, nil
        case "ctrl+c":
            return m, tea.Quit
        }
    // ... tick handling
    }
    return m, nil
}

func (m TimerModel) View() string {
    if m.confirming {
        return "\n  Give up? This will be logged as abandoned 💀\n\n  [y] yes  [n] no\n"
    }
    // ... normal view
}
```

### 📝 Checkpoint Tasks

- [ ] Timer counts down
- [ ] Progress bar fills
- [ ] 'q' shows confirmation
- [ ] 'n' resumes timer

---

## 5. Phase 3: Multiple Views & Navigation

**Goal**: Create a menu system.

### The View Switching Pattern

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

### Exercise 3.1: Styles File

Create `ui/styles.go`:

```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
    Primary   = lipgloss.Color("205")
    Secondary = lipgloss.Color("86")
    Success   = lipgloss.Color("82")
    Danger    = lipgloss.Color("196")
    Muted     = lipgloss.Color("241")
    
    TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(Primary)
    HelpStyle     = lipgloss.NewStyle().Foreground(Muted)
    SelectedStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true)
    NormalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
    SuccessStyle  = lipgloss.NewStyle().Foreground(Success).Bold(true)
    ErrorStyle    = lipgloss.NewStyle().Foreground(Danger).Bold(true)
    
    BoxStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(Primary).
        Padding(1, 2)
)
```

### Exercise 3.2: Menu Model

Create `ui/menu.go`:

```go
package ui

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
)

type MenuChoice int

const (
    StartSession MenuChoice = iota
    ViewStats
    ManageQuotes
    QuitApp
)

type MenuModel struct {
    choices []string
    cursor  int
    streak  int
}

func NewMenuModel() MenuModel {
    return MenuModel{
        choices: []string{
            "🍅 Start Focus Session",
            "📊 View Statistics",
            "💬 Manage Quotes",
            "👋 Quit",
        },
    }
}

func (m *MenuModel) SetStreak(s int) { m.streak = s }

func (m MenuModel) Init() tea.Cmd { return nil }

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            if m.cursor > 0 { m.cursor-- }
        case "down", "j":
            if m.cursor < len(m.choices)-1 { m.cursor++ }
        case "enter", " ":
            return m, func() tea.Msg { return MenuSelectionMsg(m.cursor) }
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m MenuModel) View() string {
    title := TitleStyle.Render("🍅 Beot")
    
    streak := HelpStyle.Render("Start a session to begin!")
    if m.streak > 0 {
        streak = SuccessStyle.Render(fmt.Sprintf("🔥 %d day streak", m.streak))
    }
    
    var items string
    for i, choice := range m.choices {
        cursor, style := "  ", NormalStyle
        if m.cursor == i {
            cursor, style = "▸ ", SelectedStyle
        }
        items += fmt.Sprintf("%s%s\n", cursor, style.Render(choice))
    }
    
    help := HelpStyle.Render("↑/↓ navigate • enter select")
    return fmt.Sprintf("\n  %s\n  %s\n\n%s\n  %s\n", title, streak, items, help)
}

type MenuSelectionMsg int
```

### Exercise 3.3: App Container

Create `ui/app.go`:

```go
package ui

import tea "github.com/charmbracelet/bubbletea"

type View int

const (
    MenuViewState View = iota
    SubjectSelectViewState
    TimerViewState
    StatsViewState
)

type AppModel struct {
    currentView View
    menu        MenuModel
    timer       TimerModel
}

func NewAppModel() AppModel {
    return AppModel{
        currentView: MenuViewState,
        menu:        NewMenuModel(),
    }
}

func (m AppModel) Init() tea.Cmd { return nil }

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "esc" && m.currentView != MenuViewState {
            m.currentView = MenuViewState
            return m, nil
        }
    
    case MenuSelectionMsg:
        switch MenuChoice(msg) {
        case StartSession:
            m.timer = NewTimerModel(25)
            m.currentView = TimerViewState
            return m, m.timer.Init()
        case ViewStats:
            m.currentView = StatsViewState
        case QuitApp:
            return m, tea.Quit
        }
        return m, nil
    
    case TimerCompleteMsg:
        m.currentView = MenuViewState
        return m, nil
    }
    
    // Route to current view
    switch m.currentView {
    case MenuViewState:
        newMenu, cmd := m.menu.Update(msg)
        m.menu = newMenu.(MenuModel)
        return m, cmd
    case TimerViewState:
        newTimer, cmd := m.timer.Update(msg)
        m.timer = newTimer.(TimerModel)
        return m, cmd
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
        return "\n  Stats (TODO)\n\n  esc to go back\n"
    default:
        return "Unknown"
    }
}

type TimerCompleteMsg struct{}
type TimerAbandonMsg struct{}
```

Update `main.go`:

```go
package main

import (
    "fmt"
    "os"
    tea "github.com/charmbracelet/bubbletea"
    "beot/ui"
)

func main() {
    p := tea.NewProgram(ui.NewAppModel(), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}
```

### 📝 Checkpoint Tasks

- [ ] Navigate menu with arrows
- [ ] Start timer from menu
- [ ] Return to menu on completion
- [ ] esc returns from other views

---

## 6. Phase 4: MongoDB Integration

**Goal**: Connect to MongoDB and create CRUD operations.

### Exercise 4.1: Database Connection

Create `db/mongo.go`:

```go
package db

import (
    "context"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
    client   *mongo.Client
    database *mongo.Database
}

const (
    QuotesCollection   = "quotes"
    SessionsCollection = "sessions"
    SubjectsCollection = "subjects"
)

func Connect(uri, dbName string) (*DB, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        return nil, fmt.Errorf("connect failed: %w", err)
    }
    
    if err = client.Ping(ctx, nil); err != nil {
        return nil, fmt.Errorf("ping failed: %w", err)
    }
    
    return &DB{client: client, database: client.Database(dbName)}, nil
}

func (db *DB) Disconnect() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return db.client.Disconnect(ctx)
}

func (db *DB) Collection(name string) *mongo.Collection {
    return db.database.Collection(name)
}
```

### Exercise 4.2: Models

Create `models/quote.go`:

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Quote struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Text      string             `bson:"text"`
    Source    string             `bson:"source,omitempty"`
    CreatedAt time.Time          `bson:"created_at"`
}
```

Create `models/subject.go`:

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Subject struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Name      string             `bson:"name"`
    Icon      string             `bson:"icon"`
    Color     string             `bson:"color"`
    CreatedAt time.Time          `bson:"created_at"`
}

var DefaultSubjects = []Subject{
    {Name: "GoLang", Icon: "🐹", Color: "#00ADD8"},
    {Name: "Godot", Icon: "🎮", Color: "#478CBF"},
    {Name: "React", Icon: "⚛️", Color: "#61DAFB"},
    {Name: "Music", Icon: "🎹", Color: "#9B59B6"},
    {Name: "Reading", Icon: "📚", Color: "#F39C12"},
    {Name: "General", Icon: "⭐", Color: "#ECF0F1"},
}
```

Create `models/session.go`:

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type SessionStatus string

const (
    StatusCompleted SessionStatus = "completed"
    StatusAbandoned SessionStatus = "abandoned"
)

type Session struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    SubjectID   primitive.ObjectID `bson:"subject_id"`
    SubjectName string             `bson:"subject_name"`
    SubjectIcon string             `bson:"subject_icon"`
    Duration    int                `bson:"duration"`
    Status      SessionStatus      `bson:"status"`
    StartedAt   time.Time          `bson:"started_at"`
    CompletedAt time.Time          `bson:"completed_at"`
}
```

### Exercise 4.3: Quote Repository

Create `db/quotes.go`:

```go
package db

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "beot/models"
)

type QuoteRepo struct {
    db *DB
}

func NewQuoteRepo(db *DB) *QuoteRepo {
    return &QuoteRepo{db: db}
}

func (r *QuoteRepo) Create(ctx context.Context, text, source string) (*models.Quote, error) {
    quote := models.Quote{
        Text:      text,
        Source:    source,
        CreatedAt: time.Now(),
    }
    
    result, err := r.db.Collection(QuotesCollection).InsertOne(ctx, quote)
    if err != nil {
        return nil, err
    }
    
    quote.ID = result.InsertedID.(primitive.ObjectID)
    return &quote, nil
}

func (r *QuoteRepo) GetAll(ctx context.Context) ([]models.Quote, error) {
    cursor, err := r.db.Collection(QuotesCollection).Find(ctx, bson.M{})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var quotes []models.Quote
    return quotes, cursor.All(ctx, &quotes)
}

func (r *QuoteRepo) GetRandom(ctx context.Context) (*models.Quote, error) {
    pipeline := bson.A{bson.M{"$sample": bson.M{"size": 1}}}
    
    cursor, err := r.db.Collection(QuotesCollection).Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var quotes []models.Quote
    if err := cursor.All(ctx, &quotes); err != nil {
        return nil, err
    }
    
    if len(quotes) == 0 {
        return nil, nil
    }
    return &quotes[0], nil
}

func (r *QuoteRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
    _, err := r.db.Collection(QuotesCollection).DeleteOne(ctx, bson.M{"_id": id})
    return err
}

func (r *QuoteRepo) SeedDefaults(ctx context.Context) error {
    count, _ := r.db.Collection(QuotesCollection).CountDocuments(ctx, bson.M{})
    if count > 0 {
        return nil
    }
    
    quotes := []interface{}{
        models.Quote{Text: "Some of the greatest innovations have come from people who only succeeded because they were too dumb to know that what they were doing was impossible.", CreatedAt: time.Now()},
        models.Quote{Text: "Game design is decision making, and decisions must be made with confidence.", CreatedAt: time.Now()},
        models.Quote{Text: "If you aren't dropping, you aren't learning. And if you aren't learning, you aren't a juggler.", Source: "Juggler's saying", CreatedAt: time.Now()},
        models.Quote{Text: "A computer is a creative amplifier.", CreatedAt: time.Now()},
    }
    
    _, err := r.db.Collection(QuotesCollection).InsertMany(ctx, quotes)
    return err
}
```

### 📝 Checkpoint Tasks

- [ ] Connect to MongoDB on app start
- [ ] Seed default quotes
- [ ] Verify in mongosh: `db.quotes.find()`
- [ ] Create SubjectRepo following same pattern

---

## 7. Phase 5: Subjects & Sessions

**Goal**: Subject selection and session saving.

### Exercise 5.1: Subject Select UI

Create `ui/subject_select.go`:

```go
package ui

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "beot/models"
)

type SubjectSelectModel struct {
    subjects []models.Subject
    cursor   int
}

func NewSubjectSelectModel(subjects []models.Subject) SubjectSelectModel {
    return SubjectSelectModel{subjects: subjects}
}

func (m SubjectSelectModel) Init() tea.Cmd { return nil }

func (m SubjectSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            if m.cursor > 0 { m.cursor-- }
        case "down", "j":
            if m.cursor < len(m.subjects) { m.cursor++ }
        case "enter", " ":
            if m.cursor == len(m.subjects) {
                return m, nil // Add new (TODO)
            }
            return m, func() tea.Msg {
                return SubjectSelectedMsg{Subject: m.subjects[m.cursor]}
            }
        case "esc":
            return m, func() tea.Msg { return BackToMenuMsg{} }
        }
    }
    return m, nil
}

func (m SubjectSelectModel) View() string {
    title := TitleStyle.Render("What are you focusing on?")
    
    var items string
    for i, s := range m.subjects {
        cursor, style := "  ", NormalStyle
        if m.cursor == i {
            cursor, style = "▸ ", SelectedStyle
        }
        subStyle := style.Copy().Foreground(lipgloss.Color(s.Color))
        items += fmt.Sprintf("%s%s %s\n", cursor, s.Icon, subStyle.Render(s.Name))
    }
    
    // Add new option
    addCursor, addStyle := "  ", HelpStyle
    if m.cursor == len(m.subjects) {
        addCursor, addStyle = "▸ ", SelectedStyle
    }
    items += fmt.Sprintf("\n%s%s\n", addCursor, addStyle.Render("+ Add new"))
    
    help := HelpStyle.Render("↑/↓ navigate • enter select • esc back")
    return fmt.Sprintf("\n  %s\n\n%s\n  %s\n", title, items, help)
}

type SubjectSelectedMsg struct{ Subject models.Subject }
type BackToMenuMsg struct{}
```

### Exercise 5.2: Session Repository

Create `db/sessions.go`:

```go
package db

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo/options"
    "beot/models"
)

type SessionRepo struct {
    db *DB
}

func NewSessionRepo(db *DB) *SessionRepo {
    return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, session *models.Session) error {
    result, err := r.db.Collection(SessionsCollection).InsertOne(ctx, session)
    if err != nil {
        return err
    }
    session.ID = result.InsertedID.(primitive.ObjectID)
    return nil
}

func (r *SessionRepo) GetToday(ctx context.Context) ([]models.Session, error) {
    now := time.Now()
    startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
    
    cursor, err := r.db.Collection(SessionsCollection).Find(ctx,
        bson.M{"started_at": bson.M{"$gte": startOfDay}},
        options.Find().SetSort(bson.M{"started_at": 1}),
    )
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var sessions []models.Session
    return sessions, cursor.All(ctx, &sessions)
}

func (r *SessionRepo) GetDaysWithSessions(ctx context.Context, limit int) ([]time.Time, error) {
    pipeline := bson.A{
        bson.M{"$match": bson.M{"status": models.StatusCompleted}},
        bson.M{"$sort": bson.M{"started_at": -1}},
        bson.M{"$group": bson.M{
            "_id": bson.M{
                "year":  bson.M{"$year": "$started_at"},
                "month": bson.M{"$month": "$started_at"},
                "day":   bson.M{"$dayOfMonth": "$started_at"},
            },
        }},
        bson.M{"$sort": bson.M{"_id.year": -1, "_id.month": -1, "_id.day": -1}},
        bson.M{"$limit": limit},
    }
    
    cursor, err := r.db.Collection(SessionsCollection).Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var results []struct {
        ID struct {
            Year, Month, Day int `bson:"year,month,day"`
        } `bson:"_id"`
    }
    if err := cursor.All(ctx, &results); err != nil {
        return nil, err
    }
    
    days := make([]time.Time, len(results))
    for i, r := range results {
        days[i] = time.Date(r.ID.Year, time.Month(r.ID.Month), r.ID.Day, 0, 0, 0, 0, time.UTC)
    }
    return days, nil
}
```

### 📝 Checkpoint Tasks

- [ ] Select subject before starting timer
- [ ] See subject icon on timer
- [ ] Complete session, verify in DB
- [ ] Abandon session, verify status is "abandoned"

---

## 8. Phase 6: Quotes System

**Goal**: Rotating quotes during timer.

### Quote Rotation

```go
// In timer.go

type quoteRotateMsg struct{}

func rotateQuoteCmd() tea.Cmd {
    return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
        return quoteRotateMsg{}
    })
}

// Add to TimerModel:
type TimerModel struct {
    // ... existing
    quotes          []models.Quote
    currentQuoteIdx int
}

// In Update:
case quoteRotateMsg:
    if m.running && len(m.quotes) > 1 {
        m.currentQuoteIdx = (m.currentQuoteIdx + 1) % len(m.quotes)
        return m, rotateQuoteCmd()
    }

// Start both commands:
func (m *TimerModel) Start() tea.Cmd {
    m.running = true
    return tea.Batch(tickCmd(), rotateQuoteCmd())
}
```

---

## 9. Phase 7: Stats & Streaks

**Goal**: Display statistics and streak.

### Streak Calculator

Create `internal/streak/calculator.go`:

```go
package streak

import "time"

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
            break
        }
    }
    return streak
}
```

### Stats View

```go
// Show today's sessions as icons:
func renderToday(sessions []models.Session) string {
    if len(sessions) == 0 {
        return "Today: No sessions yet"
    }
    
    var icons string
    for _, s := range sessions {
        if s.Status == models.StatusCompleted {
            icons += s.SubjectIcon + " "
        } else {
            icons += "💀 "
        }
    }
    return fmt.Sprintf("Today: %s", icons)
}
```

---

## 10. Phase 8: Polish & Sound

### Terminal Bell

```go
// When timer completes:
if m.remainingSeconds <= 0 {
    m.running = false
    fmt.Print("\a")  // Terminal bell
    return m, func() tea.Msg { return TimerCompleteMsg{} }
}
```

### Completion Screen

```go
func (m TimerModel) renderComplete() string {
    box := BoxStyle.Copy().BorderForeground(Success)
    content := fmt.Sprintf(
        "%s\n\n%s %s\n\n%s",
        SuccessStyle.Render("✅ Session Complete!"),
        m.subject.Icon, m.subject.Name,
        HelpStyle.Render("Press any key"),
    )
    return fmt.Sprintf("\n\n%s\n", box.Render(content))
}
```

---

## 11. Challenges & Extensions

Once core works, try these:

1. **Break Timer** - 5min short, 15min long every 4 sessions
2. **Text Input** - Add custom subjects with `textinput` bubble
3. **Config File** - YAML/JSON for settings
4. **Weekly Goal** - Track progress toward target
5. **Export** - CSV/JSON export of history
6. **Themes** - Multiple colour schemes
7. **Session Notes** - Add notes on completion

---

## 12. Resources

### Documentation
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Bubbles (Components)](https://github.com/charmbracelet/bubbles)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)

### Quick Reference

**Bubble Tea Lifecycle:**
```
Init() → Update(msg) → View() → (repeat)
```

**Common Commands:**
```go
tea.Quit
tea.Batch(cmd1, cmd2)
tea.Tick(duration, func)
```

**Lip Gloss:**
```go
style := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("205")).
    Border(lipgloss.RoundedBorder()).
    Padding(1, 2)

output := style.Render("Hello!")
```

**MongoDB:**
```go
collection.InsertOne(ctx, doc)
collection.Find(ctx, filter)
collection.UpdateOne(ctx, filter, update)
collection.DeleteOne(ctx, filter)
collection.Aggregate(ctx, pipeline)
```

---

## Final Notes

Work through phases in order. Each builds on the last. Most importantly: **make mistakes**. Go error messages help, and debugging is where learning happens.

Good luck building your Pomodoro timer! 🍅

---

*"If you aren't dropping, you aren't learning."*
