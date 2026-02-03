# Changelog

## [Unreleased]

### Added
- **Windows Installer** - Inno Setup script with PATH integration
  - Installs to Program Files
  - Adds Beot to Windows PATH
  - Creates Start Menu shortcuts
  - Clean uninstall with PATH removal
- **Old English Poems** - Passages from The Wanderer and Beowulf
  - `poems` MongoDB collection with Old English and Modern English text
  - 14 curated passages with line references
  - Gold-styled Old English text with translation below
  - Menu toggle to switch between Quotes and Poems display mode
- **Release Automation**
  - GoReleaser config for cross-platform builds (Windows, Linux, macOS)
  - GitHub Actions workflow for automated releases
  - Version info embedded via ldflags (`--version` flag)
- **Environment Configuration**
  - `.env` file support via godotenv
  - `BEOT_MONGODB_URI` environment variable (required)
  - `.env.example` template

### Changed
- MongoDB credentials moved from hardcoded to environment variable

### Security
- Removed hardcoded database credentials from source code

## [0.2.0] - 2026-02-03

### Added
- **Pomodoro Timer** - 25-minute focus sessions
  - Progress bar with Anglo-Saxon gold gradient
  - Pause/resume with spacebar
  - Session completion and abandonment tracking
  - Terminal bell on completion
- **Subject Selection** - Choose focus area before starting
  - GoLang, React, Music, Reading, Writing
  - Custom icons for each subject
- **Quotes System** - Rotating motivational quotes
  - Displayed during timer sessions
  - Rotate every 3 minutes
  - Manage quotes view (add/delete)
- **Statistics View**
  - Completed and abandoned session counts
  - Total focus time
  - Current and longest streaks
  - Sessions by subject breakdown
- **MongoDB Integration**
  - `sessions`, `quotes`, `subjects` collections
  - Seed command for initial data
- **Multiple Views & Navigation**
  - Menu, Subject Select, Timer, Stats, Quotes views
  - Elm Architecture (Model-Update-View) via Bubble Tea

## [0.1.0] - Initial Release

### Added
- Counter app as initial Bubble Tea learning exercise
- Anglo-Saxon themed colour palette (gold, sage green, rust, weathered stone)
- Togglable help text with `h` key
- Alternate screen mode (`tea.WithAltScreen`) for clean terminal experience
- Keyboard controls: increment, decrement, double, reset counter
- Lip Gloss styling with rounded border box layout

### Fixed
- Fixed import using single quotes instead of double quotes
- Fixed colon instead of semicolon in error handling (`p.Run()`)
- Fixed counter variable declared as `int` instead of `string` in View
