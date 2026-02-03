package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"Beot/db"
)

type QuotesModel struct {
	quotes      []db.Quote
	cursor      int
	adding      bool
	textInput   textinput.Model
	sourceInput textinput.Model
	inputFocus  int // 0 = text, 1 = source
	err         error
}

func NewQuotesModel() QuotesModel {
	ti := textinput.New()
	ti.Placeholder = "Enter quote text..."
	ti.CharLimit = 500
	ti.Width = 60

	si := textinput.New()
	si.Placeholder = "Source (optional)"
	si.CharLimit = 100
	si.Width = 40

	return QuotesModel{
		textInput:   ti,
		sourceInput: si,
	}
}

func (m *QuotesModel) LoadQuotes() tea.Cmd {
	return func() tea.Msg {
		quotes, err := db.GetAllQuotes()
		if err != nil {
			return QuotesLoadedMsg{Err: err}
		}
		return QuotesLoadedMsg{Quotes: quotes}
	}
}

type QuotesLoadedMsg struct {
	Quotes []db.Quote
	Err    error
}

type QuoteAddedMsg struct {
	Quote *db.Quote
	Err   error
}

type QuoteDeletedMsg struct {
	Err error
}

func (m QuotesModel) Init() tea.Cmd {
	return m.LoadQuotes()
}

func (m QuotesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case QuotesLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.quotes = msg.Quotes
		}
		return m, nil

	case QuoteAddedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.quotes = append(m.quotes, *msg.Quote)
			m.adding = false
			m.textInput.Reset()
			m.sourceInput.Reset()
		}
		return m, nil

	case QuoteDeletedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		return m, m.LoadQuotes()

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
			if m.cursor < len(m.quotes)-1 {
				m.cursor++
			}
		case "a":
			m.adding = true
			m.textInput.Focus()
			m.inputFocus = 0
			return m, textinput.Blink
		case "d", "delete":
			if len(m.quotes) > 0 {
				return m, m.deleteCurrentQuote()
			}
		}
	}

	return m, nil
}

func (m QuotesModel) handleAddingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.adding = false
		m.textInput.Reset()
		m.sourceInput.Reset()
		return m, nil
	case "tab":
		if m.inputFocus == 0 {
			m.inputFocus = 1
			m.textInput.Blur()
			m.sourceInput.Focus()
		} else {
			m.inputFocus = 0
			m.sourceInput.Blur()
			m.textInput.Focus()
		}
		return m, nil
	case "enter":
		if m.inputFocus == 0 {
			// Move to source input
			m.inputFocus = 1
			m.textInput.Blur()
			m.sourceInput.Focus()
			return m, nil
		}
		// Submit the quote
		text := m.textInput.Value()
		if text == "" {
			return m, nil
		}
		source := m.sourceInput.Value()
		return m, func() tea.Msg {
			quote, err := db.AddQuote(text, source)
			return QuoteAddedMsg{Quote: quote, Err: err}
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	if m.inputFocus == 0 {
		m.textInput, cmd = m.textInput.Update(msg)
	} else {
		m.sourceInput, cmd = m.sourceInput.Update(msg)
	}
	return m, cmd
}

func (m QuotesModel) deleteCurrentQuote() tea.Cmd {
	if m.cursor >= len(m.quotes) {
		return nil
	}
	id := m.quotes[m.cursor].ID
	return func() tea.Msg {
		err := db.DeleteQuote(id)
		return QuoteDeletedMsg{Err: err}
	}
}

func (m QuotesModel) View() string {
	title := TitleStyle.Render("💬 Manage Quotes")

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

func (m QuotesModel) renderAddForm(title string) string {
	form := fmt.Sprintf(
		"Quote:\n%s\n\nSource:\n%s",
		m.textInput.View(),
		m.sourceInput.View(),
	)

	help := HelpStyle.Render("tab switch field • enter next/submit • esc cancel")

	return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", title, form, help)
}

func (m QuotesModel) renderList(title string) string {
	if len(m.quotes) == 0 {
		empty := NormalStyle.Render("No quotes yet. Press 'a' to add one.")
		help := HelpStyle.Render("a add • esc/q back to menu")
		return fmt.Sprintf("\n  %s\n\n  %s\n\n  %s\n", title, empty, help)
	}

	var list string
	for i, q := range m.quotes {
		cursor := "  "
		style := NormalStyle
		if i == m.cursor {
			cursor = "▸ "
			style = SelectedStyle
		}

		text := q.Text
		if len(text) > 50 {
			text = text[:50] + "..."
		}
		if q.Source != "" {
			text += " — " + q.Source
		}
		list += fmt.Sprintf("%s%s\n", cursor, style.Render(text))
	}

	help := HelpStyle.Render("↑/↓ navigate • a add • d delete • esc/q back")

	return fmt.Sprintf("\n  %s\n\n%s\n  %s\n", title, list, help)
}

// BackToMenuMsg signals to return to the main menu
type BackToMenuMsg struct{}
