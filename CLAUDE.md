# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Beot (Anglo-Saxon word meaning "a binding vow") is a terminal-based Pomodoro timer built with Go. It uses Bubble Tea for the TUI framework, Lip Gloss for styling, and MongoDB for persistence.

**Core Features:**
- 25-minute focus sessions tied to subjects (GoLang, Music, React, etc.)
- Tracks both completed AND abandoned sessions for accountability
- Rotating motivational quotes during sessions
- Streaks and statistics

## Development Commands

```bash
# Run the application
go run main.go

# Build binary
go build -o beot

# Manage dependencies
go mod tidy

# Start MongoDB (Docker)
docker run -d -p 27017:27017 --name beot-mongo mongo:latest
```

## Architecture

The application follows the Elm Architecture (Model-Update-View) via Bubble Tea:

```
Model ──► View() ──► Terminal Output
  ▲
  └─── Update(msg) ◄── User Input / Events
```

**Planned Package Structure:**
- `ui/` - Bubble Tea models and views (app.go, menu.go, timer.go, subject_select.go, stats.go, styles.go)
- `db/` - MongoDB repositories (mongo.go, sessions.go, quotes.go, subjects.go)
- `models/` - Data structures (session.go, subject.go, quote.go)
- `internal/streak/` - Streak calculation logic

**Key Patterns:**
- Each UI view is a Bubble Tea model implementing `Init()`, `Update()`, `View()`
- `AppModel` acts as the root container, routing messages to child views
- Commands produce messages; `tea.Tick` for time-based events
- State flows unidirectionally

## Key Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - Pre-built components (progress, textinput)
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `go.mongodb.org/mongo-driver/mongo` - Database driver

## Implementation Reference

The `documentation/BEOT_LEARNING_GUIDE.md` contains a comprehensive 8-phase tutorial with complete code examples:
1. Hello Bubble Tea - Basic framework
2. Building the Timer - Countdown logic with tick commands
3. Multiple Views & Navigation - Menu system and view switching
4. MongoDB Integration - Database connection and repositories
5. Subjects & Sessions - Subject selection and session persistence
6. Quotes System - Rotating quotes with `tea.Tick`
7. Stats & Streaks - Statistics view and streak calculation
8. Polish & Sound - Terminal bells and completion screens

## MongoDB Collections

- `quotes` - Motivational quotes (text, source, created_at)
- `sessions` - Pomodoro sessions (subject_id, duration, status, started_at, completed_at)
- `subjects` - Focus subjects (name, icon, color)

Session status values: `completed`, `abandoned`
