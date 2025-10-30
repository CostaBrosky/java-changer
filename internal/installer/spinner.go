package installer

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type spinnerFinishedMsg struct {
	err error
}

type spinnerModel struct {
	spinner  spinner.Model
	message  string
	quitting bool
	err      error
}

func newSpinnerModel(message string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))

	return spinnerModel{
		spinner: s,
		message: message,
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case spinnerFinishedMsg:
		m.err = msg.err
		m.quitting = true
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m spinnerModel) View() string {
	if m.quitting {
		if m.err != nil {
			return ""
		}
		return ""
	}
	return fmt.Sprintf("\n %s %s\n\n", m.spinner.View(), m.message)
}

// WithSpinner runs a function with a spinner animation
func WithSpinner(message string, fn func() error) error {
	p := tea.NewProgram(newSpinnerModel(message))

	// Run function in background
	go func() {
		time.Sleep(100 * time.Millisecond) // Give UI time to start
		err := fn()
		p.Send(spinnerFinishedMsg{err: err})
	}()

	// Run the spinner UI
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
