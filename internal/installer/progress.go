package installer

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding  = 2
	maxWidth = 80
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

type progressMsg struct {
	percent    float64
	downloaded int64
	speed      string
}

type progressErrMsg struct{ err error }

type downloadCompleteMsg struct{}

// ProgressWriter wraps an io.Writer and tracks download progress
type ProgressWriter struct {
	total      int64
	downloaded int64
	startTime  time.Time
	onProgress func(float64)
}

func NewProgressWriter(total int64, onProgress func(float64)) *ProgressWriter {
	return &ProgressWriter{
		total:      total,
		downloaded: 0,
		startTime:  time.Now(),
		onProgress: onProgress,
	}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)

	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(float64(pw.downloaded) / float64(pw.total))
	}

	return n, nil
}

// GetSpeed returns current download speed in bytes per second
func (pw *ProgressWriter) GetSpeed() float64 {
	elapsed := time.Since(pw.startTime).Seconds()
	if elapsed > 0 {
		return float64(pw.downloaded) / elapsed
	}
	return 0
}

// FormatSpeed formats speed in human-readable format
func (pw *ProgressWriter) FormatSpeed() string {
	speed := pw.GetSpeed()
	if speed >= 1024*1024 {
		return fmt.Sprintf("%.2f MB/s", speed/(1024*1024))
	} else if speed >= 1024 {
		return fmt.Sprintf("%.2f KB/s", speed/1024)
	}
	return fmt.Sprintf("%.0f B/s", speed)
}

// FormatSize formats bytes in human-readable format
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ProgressModel represents the Bubble Tea model for download progress
type ProgressModel struct {
	progress   progress.Model
	totalBytes int64
	downloaded int64
	speed      string
	err        error
	done       bool
}

func NewProgressModel(totalBytes int64) ProgressModel {
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return ProgressModel{
		progress:   prog,
		totalBytes: totalBytes,
		downloaded: 0,
		speed:      "0 B/s",
		done:       false,
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return nil
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil

	case progressMsg:
		if m.done {
			return m, tea.Quit
		}

		// Update progress data
		m.downloaded = msg.downloaded
		m.speed = msg.speed

		// Update progress bar
		cmd := m.progress.SetPercent(msg.percent)
		if msg.percent >= 1.0 {
			m.done = true
			return m, tea.Sequence(cmd, tea.Quit)
		}
		return m, cmd

	case downloadCompleteMsg:
		m.done = true
		return m, tea.Quit

	case progressErrMsg:
		m.err = msg.err
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m ProgressModel) View() string {
	if m.err != nil {
		return "Error downloading: " + m.err.Error() + "\n"
	}

	if m.done {
		return ""
	}

	pad := strings.Repeat(" ", padding)

	// Progress bar
	progressBar := m.progress.View()

	// Calculate percentage
	percent := m.progress.Percent() * 100

	// Downloaded / Total
	downloaded := FormatSize(m.downloaded)
	total := FormatSize(m.totalBytes)

	// Build the view
	info := fmt.Sprintf("%s / %s (%.0f%%) - %s",
		downloaded, total, percent, m.speed)

	return "\n" +
		pad + progressBar + "\n" +
		pad + helpStyle(info) + "\n"
}

// progressWriter is an io.Writer that sends progress updates to Bubble Tea
type progressWriter struct {
	total      int64
	downloaded int64
	startTime  time.Time
	program    *tea.Program
}

func newProgressWriter(total int64, program *tea.Program) *progressWriter {
	return &progressWriter{
		total:      total,
		downloaded: 0,
		startTime:  time.Now(),
		program:    program,
	}
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)

	if pw.total > 0 && pw.program != nil {
		percent := float64(pw.downloaded) / float64(pw.total)
		speed := pw.GetSpeed()

		pw.program.Send(progressMsg{
			percent:    percent,
			downloaded: pw.downloaded,
			speed:      speed,
		})
	}

	return n, nil
}

func (pw *progressWriter) GetSpeed() string {
	elapsed := time.Since(pw.startTime).Seconds()
	if elapsed > 0 {
		speed := float64(pw.downloaded) / elapsed
		if speed >= 1024*1024 {
			return fmt.Sprintf("%.2f MB/s", speed/(1024*1024))
		} else if speed >= 1024 {
			return fmt.Sprintf("%.2f KB/s", speed/1024)
		}
		return fmt.Sprintf("%.0f B/s", speed)
	}
	return "0 B/s"
}
