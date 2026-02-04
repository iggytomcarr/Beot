# Bēot

**Bēot** (Anglo-Saxon: "a binding vow") is a terminal-based Pomodoro timer built with Go. It uses Bubble Tea for the TUI framework, Lip Gloss for styling, and MongoDB for persistence.

What is a Bēot?

In Anglo-Saxon culture, a bēot was a public, binding vow, spoken aloud in the mead hall before witnesses.
A warrior did not boast about past deeds — he declared what he would do next, and bound his honour to the outcome.

A bēot was not aspirational. It was terminal.

You either:

fulfilled the vow, gaining honour and lasting reputation, or

failed it, and lived with the loss of name, standing, and trust.

There was no quiet abandonment. Once spoken, a bēot existed — remembered by the hall, the lord, and fate itself (wyrd).

Bēot (this application) treats every focus session the same way.

When you start a session, you are making a vow:

For the next 25 minutes, I will hold to this work.

Completing a session is keeping your word.
Abandoning it is breaking your vow.

Both outcomes are recorded — not to punish, but to make the promise real.
Like the original ritual, the power of a bēot lies in its witness and memory.

This tool does not ask for perfection.
It asks for honesty, presence, and the courage to speak your intent — and then live with the result.

c þæt þonne forhicge,
swā mē Higelāc sīe mīn mundbora,
wiþ þā grimman gryre-gæst Grendel
gefeohtan fēa sīðe,
hand-geswingum, swā hē hēa rīce
for ealra þāra eorla rīcsode.

Nō ic mē mid sweorde oððe mid swylcum woruld-wǣpnum
wiþ þone gāst wīðfeohte,
ac ic hine hand-ġeswingum,
fēond on frēce, fēorh benēote.

— Beowulf, approx. lines 433–441

# So
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

## Future Features

### My Wyrd

**My Wyrd** is a planned web-based companion to Bēot — a shareable, visual rendering of your focus journey.

- Generate a unique graphic showing your sessions, streaks, and progress
- Share with friends to show your commitment and accomplishments
- Visualise your wyrd (fate) as shaped by your vows kept and broken
- Built with JavaScript, designed to be embedded or shared on social media

*"Wyrd" in Anglo-Saxon culture refers to the concept of fate or destiny — the web of events that shapes one's life. My Wyrd will show the tapestry of your focus sessions over time.*

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework (Elm Architecture)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - Pre-built components (progress bars, text inputs)
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) - Data persistence

## Installation

### Windows Installer (Recommended)

Download the latest `BeotSetup-x.x.x.exe` from [Releases](https://github.com/yourusername/beot/releases).

The installer will:
- Install `beot.exe` to Program Files
- Add Beot to your PATH (so you can run `beot` from any terminal)
- Create Start Menu shortcuts

After installation, open any terminal and run:
```powershell
beot
```

### Configuration

Beot requires a MongoDB connection. Create a `.env` file from the template:

```bash
cp .env.example .env
```

Then edit `.env` with your MongoDB connection string:

```bash
# For local MongoDB:
BEOT_MONGODB_URI=mongodb://localhost:27017

# For MongoDB Atlas:
BEOT_MONGODB_URI=mongodb+srv://<username>:<password>@<cluster>.mongodb.net/?retryWrites=true&w=majority
```

The `.env` file is gitignored and will not be committed.

### From Source

#### Prerequisites

- Go 1.21+
- MongoDB (local or cloud)

#### Run

```bash
go run main.go
```

#### Build

```bash
go build -o beot
```

#### Start MongoDB (Local)

```bash
docker run -d -p 27017:27017 --name beot-mongo mongo:latest
```

## Creating a Release

### Prerequisites

1. Convert the icon: `assets/beot.svg` → `assets/beot.ico` (see `assets/README.md`)
2. Install [Inno Setup](https://jrsoftware.org/isinfo.php) (for local builds)

### Automated Release (GitHub Actions)

Tag and push to trigger the release workflow:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This will:
1. Build Windows binary with version info embedded
2. Run Inno Setup to create the installer
3. Build binaries for Linux/macOS
4. Create a GitHub Release with all artifacts and checksums

### Manual Local Build

```bash
# Build with version info
go build -ldflags "-X main.Version=1.0.0 -X main.CommitSHA=$(git rev-parse --short HEAD) -X main.BuildDate=$(date -u +%Y-%m-%d)" -o dist/beot.exe

# Build installer (requires Inno Setup)
iscc installer/beot.iss
```

### Release Files

| File | Description |
|------|-------------|
| `BeotSetup-x.x.x.exe` | Windows installer with PATH integration |
| `beot_x.x.x_windows_amd64.zip` | Standalone Windows binary |
| `beot_x.x.x_linux_amd64.tar.gz` | Linux binary |
| `beot_x.x.x_darwin_amd64.tar.gz` | macOS binary |
| `checksums.txt` | SHA256 checksums for verification |

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

## Licence

This project is for personal learning purposes.
