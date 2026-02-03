package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"Beot/db"
)

type SubjectSelectModel struct {
	subjects  []db.Subject
	cursor    int
	adding    bool
	textInput textinput.Model
	iconInput textinput.Model
	inputFocus int
	err       error
}

func NewSubjectSelectModel() SubjectSelectModel {
	ti := textinput.New()
	ti.Placeholder = "Subject name (e.g., GoLang)"
	ti.CharLimit = 50
	ti.Width = 40

	ii := textinput.New()
	ii.Placeholder = "Icon (e.g., 🔷)"
	ii.CharLimit = 4
	ii.Width = 10

	return SubjectSelectModel{
		textInput: ti,
		iconInput: ii,
	}
}

func (m *SubjectSelectModel) LoadSubjects() tea.Cmd {
	return func() tea.Msg {
		subjects, err := db.GetAllSubjects()
		if err != nil {
			return SubjectsLoadedMsg{Err: err}
		}
		return SubjectsLoadedMsg{Subjects: subjects}
	}
}

type SubjectsLoadedMsg struct {
	Subjects []db.Subject
	Err      error
}

type SubjectSelectedMsg struct {
	Subject db.Subject
}

type SubjectAddedMsg struct {
	Subject *db.Subject
	Err     error
}

func (m SubjectSelectModel) Init() tea.Cmd {
	return m.LoadSubjects()
}

func (m SubjectSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SubjectsLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.subjects = msg.Subjects
		}
		return m, nil

	case SubjectAddedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.subjects = append(m.subjects, *msg.Subject)
			m.adding = false
			m.textInput.Reset()
			m.iconInput.Reset()
		}
		return m, nil

	case tea.KeyMsg:
		if m.adding {
			return m.handleAddingInput(msg)
		}

		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.subjects)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.subjects) > 0 && m.cursor < len(m.subjects) {
				return m, func() tea.Msg {
					return SubjectSelectedMsg{Subject: m.subjects[m.cursor]}
				}
			}
		case "a":
			m.adding = true
			m.textInput.Focus()
			m.inputFocus = 0
			return m, textinput.Blink
		}
	}

	return m, nil
}

func (m SubjectSelectModel) handleAddingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.adding = false
		m.textInput.Reset()
		m.iconInput.Reset()
		return m, nil
	case "tab":
		if m.inputFocus == 0 {
			m.inputFocus = 1
			m.textInput.Blur()
			m.iconInput.Focus()
		} else {
			m.inputFocus = 0
			m.iconInput.Blur()
			m.textInput.Focus()
		}
		return m, nil
	case "enter":
		if m.inputFocus == 0 {
			m.inputFocus = 1
			m.textInput.Blur()
			m.iconInput.Focus()
			return m, nil
		}
		// Submit the subject
		name := m.textInput.Value()
		if name == "" {
			return m, nil
		}
		icon := m.iconInput.Value()
		if icon == "" {
			icon = "📚"
		}
		return m, func() tea.Msg {
			subject, err := db.AddSubject(name, icon)
			return SubjectAddedMsg{Subject: subject, Err: err}
		}
	}

	var cmd tea.Cmd
	if m.inputFocus == 0 {
		m.textInput, cmd = m.textInput.Update(msg)
	} else {
		m.iconInput, cmd = m.iconInput.Update(msg)
	}
	return m, cmd
}

func (m SubjectSelectModel) View() string {
	title := TitleStyle.Render("Choose Your Focus")

	if m.err != nil {
		return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n",
			title,
			ErrorStyle.Render("Error: "+m.err.Error()),
			HelpStyle.Render("esc/q back to menu"),
		)
	}

	if m.adding {
		return m.renderAddForm(title)
	}

	return m.renderList(title)
}

func (m SubjectSelectModel) renderAddForm(title string) string {
	form := fmt.Sprintf(
		"Name:\n%s\n\nIcon:\n%s",
		m.textInput.View(),
		m.iconInput.View(),
	)

	help := HelpStyle.Render("tab switch field • enter next/submit • esc cancel")

	return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", title, form, help)
}

func (m SubjectSelectModel) renderList(title string) string {
	if len(m.subjects) == 0 {
		empty := NormalStyle.Render("No subjects yet. Press 'a' to add one.")
		help := HelpStyle.Render("a add subject • esc/q back to menu")
		return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", title, empty, help)
	}

	var list string
	for i, s := range m.subjects {
		cursor := "  "
		style := NormalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = SelectedStyle
		}
		icon := IconStyle.Render(s.Icon)
		list += fmt.Sprintf("%s%s%s\n", cursor, icon, style.Render(s.Name))
	}

	help := HelpStyle.Render("↑/↓ navigate • enter select • a add • esc/q back")

	return fmt.Sprintf("\n  %s\n\n%s\n  %s\n", title, list, help)
}
