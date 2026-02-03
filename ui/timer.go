package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"Beot/db"
)

// Timer messages
type tickMsg time.Time
type quoteTickMsg time.Time

// TimerCompleteMsg is sent when the timer finishes
type TimerCompleteMsg struct {
	Completed   bool   // true = completed, false = abandoned
	SubjectID   string // Subject ID for saving
	SubjectName string // Subject name for display
	Duration    int    // Duration in minutes
	StartedAt   time.Time
}

// DisplayMode determines what content is shown during the timer
type DisplayMode int

const (
	DisplayModeQuotes DisplayMode = iota
	DisplayModePoems
)

// TimerModel handles the countdown
type TimerModel struct {
	totalSeconds     int
	remainingSeconds int
	running          bool
	progress         progress.Model
	confirming       bool
	currentQuote     string
	currentSource    string
	// Poem fields for dual-language display
	currentOldEnglish    string
	currentModernEnglish string
	currentPoemSource    string
	currentPoemLineRef   string
	displayMode          DisplayMode
	subjectID            string
	subjectName          string
	startedAt            time.Time
}

// NewTimerModel creates a timer for the given minutes
func NewTimerModel(minutes int, subjectID, subjectName string) TimerModel {
	return NewTimerModelWithMode(minutes, subjectID, subjectName, DisplayModeQuotes)
}

// NewTimerModelWithMode creates a timer with specified display mode
func NewTimerModelWithMode(minutes int, subjectID, subjectName string, mode DisplayMode) TimerModel {
	seconds := minutes * 60
	prog := progress.New(progress.WithGradient("#4A3728", "#C9A84C"))
	prog.Width = 80

	m := TimerModel{
		totalSeconds:     seconds,
		remainingSeconds: seconds,
		running:          true,
		progress:         prog,
		displayMode:      mode,
		subjectID:        subjectID,
		subjectName:      subjectName,
		startedAt:        time.Now(),
	}

	// Load initial content based on mode
	if mode == DisplayModePoems {
		m.loadRandomPoem()
	} else {
		m.loadRandomQuote()
	}

	return m
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func quoteTickCmd() tea.Cmd {
	return tea.Tick(3*time.Minute, func(t time.Time) tea.Msg {
		return quoteTickMsg(t)
	})
}

func (m *TimerModel) loadRandomQuote() {
	quote, err := db.GetRandomQuoteForSubject(m.subjectName)
	if err != nil || quote == nil {
		m.currentQuote = "Focus on your task."
		m.currentSource = ""
		return
	}
	m.currentQuote = quote.Text
	m.currentSource = quote.Source
}

func (m *TimerModel) loadRandomPoem() {
	poem, err := db.GetRandomPoem()
	if err != nil || poem == nil {
		// Fallback to a default passage
		m.currentOldEnglish = "Wyrd oft nereð\nunfǽgne eorl, þonne his ellen déah"
		m.currentModernEnglish = "Fate often saves\nan undoomed man, when his courage holds"
		m.currentPoemSource = "Beowulf"
		m.currentPoemLineRef = "lines 572-573"
		return
	}
	m.currentOldEnglish = poem.OldEnglish
	m.currentModernEnglish = poem.ModernEnglish
	m.currentPoemSource = poem.Source
	m.currentPoemLineRef = poem.LineRef
}

func (m *TimerModel) loadRandomContent() {
	if m.displayMode == DisplayModePoems {
		m.loadRandomPoem()
	} else {
		m.loadRandomQuote()
	}
}

func (m TimerModel) Init() tea.Cmd {
	return tea.Batch(tickCmd(), quoteTickCmd())
}

func (m TimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// If timer is complete, any key returns to menu
		if m.remainingSeconds <= 0 {
			return m, func() tea.Msg {
				return TimerCompleteMsg{
					Completed:   true,
					SubjectID:   m.subjectID,
					SubjectName: m.subjectName,
					Duration:    m.totalSeconds / 60,
					StartedAt:   m.startedAt,
				}
			}
		}

		if m.confirming {
			switch msg.String() {
			case "y":
				return m, func() tea.Msg {
					return TimerCompleteMsg{
						Completed:   false, // Abandoned
						SubjectID:   m.subjectID,
						SubjectName: m.subjectName,
						Duration:    m.totalSeconds / 60,
						StartedAt:   m.startedAt,
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
		case "q":
			m.confirming = true
			m.running = false
			return m, nil
		case " ":
			m.running = !m.running
			if m.running {
				return m, tickCmd()
			}
			return m, nil
		case "r":
			m.remainingSeconds = m.totalSeconds
			m.running = true
			return m, tickCmd()
		}

	case tickMsg:
		if m.running && m.remainingSeconds > 0 {
			m.remainingSeconds--
			if m.remainingSeconds <= 0 {
				m.running = false
				fmt.Print("\a") // Terminal bell
				return m, func() tea.Msg {
					return TimerCompleteMsg{
						Completed:   true,
						SubjectID:   m.subjectID,
						SubjectName: m.subjectName,
						Duration:    m.totalSeconds / 60,
						StartedAt:   m.startedAt,
					}
				}
			}
			return m, tickCmd()
		}

	case quoteTickMsg:
		if m.running {
			m.loadRandomContent()
			return m, quoteTickCmd()
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
	timeDisplay := TimerStyle.Render(fmt.Sprintf("%02d:%02d", minutes, seconds))

	status := StatusStyle.Render(fmt.Sprintf("Focus Time: %s", m.subjectName))
	if !m.running && m.remainingSeconds > 0 {
		status = StatusStyle.Render("Paused")
	} else if m.remainingSeconds <= 0 {
		status = StatusStyle.Render("Complete!")
	}

	progressBar := m.progress.ViewAs(percent)
	help := HelpStyle.Render("Spacebar to pause/resume • r reset • q quit")

	header := RenderHeader()

	// Render content based on display mode
	var content string
	if m.displayMode == DisplayModePoems {
		content = RenderPoem(m.currentOldEnglish, m.currentModernEnglish, m.currentPoemSource, m.currentPoemLineRef)
	} else {
		content = RenderQuote(m.currentQuote, m.currentSource)
	}

	return fmt.Sprintf(
		"\n  %s\n\n  %s\n\n  %s\n\n  %s\n\n  %s  %s\n\n  %s\n",
		header,
		content,
		status,
		progressBar,
		timeDisplay,
		HelpStyle.Render(fmt.Sprintf("(%d%% complete)", int(percent*100))),
		help,
	)
}

func (m TimerModel) renderConfirmation() string {
	title := ErrorStyle.Render("Give up?")
	message := "This will be logged as abandoned 💀"
	help := HelpStyle.Render("[y] yes, abandon • [n] no, continue")

	return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", title, message, help)
}

func (m TimerModel) renderComplete() string {
	title := SuccessStyle.Render("Your vow is kept.")

	message := NormalStyle.Render(fmt.Sprintf(
		"You held to your word for %d minutes.\nYour honour remains unbroken.",
		m.totalSeconds/60,
	))

	subject := StatusStyle.Render(fmt.Sprintf("Subject: %s", m.subjectName))

	content := fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s",
		title,
		message,
		subject,
		HelpStyle.Render("Press any key to continue"),
	)

	return "\n" + BoxStyle.Render(content) + "\n"
}
