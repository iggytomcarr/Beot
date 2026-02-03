package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"Beot/db"
)

// View represents which screen is active
type View int

const (
	MenuViewState View = iota
	SubjectSelectViewState
	TimerViewState
	StatsViewState
	QuotesViewState
)

// AppModel is the main application container
type AppModel struct {
	currentView   View
	menu          MenuModel
	subjectSelect SubjectSelectModel
	timer         TimerModel
	quotes        QuotesModel
	stats         *db.SessionStats
	statsErr      error
}

// NewAppModel creates the application
func NewAppModel() AppModel {
	return AppModel{
		currentView: MenuViewState,
		menu:        NewMenuModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	// Load initial streak for menu display
	return func() tea.Msg {
		stats, _ := db.GetSessionStats()
		return StatsLoadedMsg{Stats: stats}
	}
}

type StatsLoadedMsg struct {
	Stats *db.SessionStats
	Err   error
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle messages that affect navigation
	switch msg := msg.(type) {

	case StatsLoadedMsg:
		m.stats = msg.Stats
		m.statsErr = msg.Err
		if msg.Stats != nil {
			m.menu.SetStreak(msg.Stats.CurrentStreak)
		}
		return m, nil

	case MenuSelectionMsg:
		switch MenuChoice(msg) {
		case StartSession:
			m.subjectSelect = NewSubjectSelectModel()
			m.currentView = SubjectSelectViewState
			return m, m.subjectSelect.LoadSubjects()
		case ViewStats:
			m.currentView = StatsViewState
			return m, func() tea.Msg {
				stats, err := db.GetSessionStats()
				return StatsLoadedMsg{Stats: stats, Err: err}
			}
		case ManageQuotes:
			m.quotes = NewQuotesModel()
			m.currentView = QuotesViewState
			return m, m.quotes.LoadQuotes()
		case QuitApp:
			return m, tea.Quit
		}
		return m, nil

	case SubjectSelectedMsg:
		m.timer = NewTimerModelWithMode(25, msg.Subject.ID.Hex(), msg.Subject.Name, m.menu.GetDisplayMode())
		m.currentView = TimerViewState
		return m, m.timer.Init()

	case BackToMenuMsg:
		m.currentView = MenuViewState
		return m, nil

	case TimerCompleteMsg:
		// Save session to database
		status := db.StatusCompleted
		if !msg.Completed {
			status = db.StatusAbandoned
		}

		subjectID, _ := primitive.ObjectIDFromHex(msg.SubjectID)
		db.CreateSession(subjectID, msg.SubjectName, msg.Duration, status, msg.StartedAt)

		// Reload stats for streak update
		m.currentView = MenuViewState
		return m, func() tea.Msg {
			stats, _ := db.GetSessionStats()
			return StatsLoadedMsg{Stats: stats}
		}
	}

	// Route messages to the active view
	switch m.currentView {
	case MenuViewState:
		newMenu, cmd := m.menu.Update(msg)
		m.menu = newMenu.(MenuModel)
		return m, cmd

	case SubjectSelectViewState:
		newSubjectSelect, cmd := m.subjectSelect.Update(msg)
		m.subjectSelect = newSubjectSelect.(SubjectSelectModel)
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

	case QuotesViewState:
		newQuotes, cmd := m.quotes.Update(msg)
		m.quotes = newQuotes.(QuotesModel)
		return m, cmd
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
	case QuotesViewState:
		return m.quotes.View()
	default:
		return "Unknown view"
	}
}

func (m AppModel) renderStats() string {
	title := TitleStyle.Render("📜 Statistics")

	if m.statsErr != nil {
		return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n",
			title,
			ErrorStyle.Render("Error loading stats: "+m.statsErr.Error()),
			HelpStyle.Render("esc/q back to menu"),
		)
	}

	if m.stats == nil {
		return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n",
			title,
			NormalStyle.Render("Loading..."),
			HelpStyle.Render("esc/q back to menu"),
		)
	}

	s := m.stats

	// Format hours and minutes
	hours := s.TotalMinutes / 60
	minutes := s.TotalMinutes % 60

	var timeStr string
	if hours > 0 {
		timeStr = fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		timeStr = fmt.Sprintf("%dm", minutes)
	}

	// Build stats display
	statsDisplay := fmt.Sprintf(
		"%s\n\n"+
			"  %sSessions Completed:  %d\n"+
			"  %sSessions Abandoned:  %d\n"+
			"  %sTotal Focus Time:    %s\n\n"+
			"%s\n\n"+
			"  %sCurrent Streak:      %d days\n"+
			"  %sLongest Streak:      %d days",
		SelectedStyle.Render("Sessions"),
		IconStyle.Render("✓"), s.CompletedSessions,
		IconStyle.Render("💀"), s.AbandonedSessions,
		IconStyle.Render("⏱"), timeStr,
		SelectedStyle.Render("Streaks"),
		IconStyle.Render("⚡"), s.CurrentStreak,
		IconStyle.Render("🏆"), s.LongestStreak,
	)

	// Get sessions by subject
	bySubject, err := db.GetSessionsBySubject()
	if err == nil && len(bySubject) > 0 {
		statsDisplay += "\n\n" + SelectedStyle.Render("By Subject") + "\n"
		for name, count := range bySubject {
			statsDisplay += fmt.Sprintf("\n  %s: %d sessions", name, count)
		}
	}

	// My Wyrd link
	wyrdLink := "\n\n" + SelectedStyle.Render("Share Your Journey") + "\n\n" +
		"  " + IconStyle.Render("🌐") + NormalStyle.Render("My Wyrd: ") + HelpStyle.Render("coming soon...")

	help := HelpStyle.Render("esc/q back to menu")

	return fmt.Sprintf("\n  %s\n\n%s%s\n\n  %s\n", title, statsDisplay, wyrdLink, help)
}
