# Bēot 

**Bēot** (Anglo-Saxon: "a binding vow") is a terminal-based Pomodoro timer built with Go. It uses Bubble Tea for the TUI framework, Lip Gloss for styling, and MongoDB for persistence.

The core idea: every focus session is a vow. You either honour it or abandon it. Both outcomes are tracked for accountability.


**The fire is high,**
**the mead is spent,**
**and the night has teeth.**

**Now is no hour for quiet men.**
**Now is the time to rise,**
**to stand tall beneath the roof-beams,**
**to let word and will be one.**

**I speak not of what I have done —**
**I speak of what I shall do.**

**By blade and breath,**
**by bone and blood,**
**I bind my honour to my word.**

**Let the gods bear witness.**
**Let the ancestors listen from the dark.**
**Let wyrd itself turn its face toward me.**

**This is my Bēot.**

**If I fail it,**
**let my name be broken.**

**If I keep it,**
**let it live after me.**


## Features

- 25-minute focus sessions tied to subjects (GoLang, Music, React, etc.)
- Tracks both completed and abandoned sessions
- Rotating motivational quotes during sessions
- Streaks and statistics
- Anglo-Saxon themed terminal UI

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework (Elm Architecture)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - Pre-built components (progress bars, text inputs)
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) - Data persistence

## Getting Started

### Prerequisites

- Go 1.21+
- MongoDB (or Docker)

### Run

```bash
go run main.go
```

### Build

```bash
go build -o beot
```

### Start MongoDB

```bash
docker run -d -p 27017:27017 --name beot-mongo mongo:latest
```

## Architecture

The application follows the Elm Architecture via Bubble Tea:

```
Model --> View() --> Terminal Output
  ^
  └── Update(msg) <-- User Input / Events
```

### Planned Structure

```
beot/
├── main.go
├── ui/          # Bubble Tea models and views
├── db/          # MongoDB repositories
├── models/      # Data structures
└── internal/
    └── streak/  # Streak calculation logic
```

### MongoDB Collections

| Collection | Purpose |
|------------|---------|
| `quotes` | Motivational quotes |
| `sessions` | Pomodoro sessions (status: completed/abandoned) |
| `subjects` | Focus subjects (name, icon, colour) |

## Controls

| Key | Action |
|-----|--------|
| `↑` / `k` | Increment counter |
| `↓` / `j` | Decrement counter |
| `d` | Double counter |
| `r` | Reset counter |
| `h` | Toggle help |
| `q` / `Ctrl+C` | Quit |

## Licence

This project is for personal learning purposes.
