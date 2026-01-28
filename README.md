# Bēot 

**Bēot** (Anglo-Saxon: "a binding vow") is a terminal-based Pomodoro timer built with Go. It uses Bubble Tea for the TUI framework, Lip Gloss for styling, and MongoDB for persistence.

The core idea: every focus session is a vow. You either honour it or abandon it. Both outcomes are tracked for accountability.

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
