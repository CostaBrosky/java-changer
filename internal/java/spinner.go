package java

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type spinnerFinishedMsg struct{}

type scannerModel struct {
	spinner  spinner.Model
	quitting bool
}

func newScannerModel() scannerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))

	return scannerModel{
		spinner: s,
	}
}

func (m scannerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m scannerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case spinnerFinishedMsg:
		m.quitting = true
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m scannerModel) View() string {
	if m.quitting {
		return ""
	}
	return fmt.Sprintf(" %s Scanning for Java installations...\n", m.spinner.View())
}

// WithScanner runs a function with a scanner animation
func WithScanner(fn func() error) error {
	p := tea.NewProgram(newScannerModel())

	// Run function in background
	go func() {
		time.Sleep(50 * time.Millisecond) // Give UI time to start
		fn()
		p.Send(spinnerFinishedMsg{})
	}()

	// Run the spinner UI
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
