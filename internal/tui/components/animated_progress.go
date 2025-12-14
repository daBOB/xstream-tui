// Package components provides reusable UI components for the TUI.
package components

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AnimatedProgress is an animated progress bar component.
type AnimatedProgress struct {
	progress progress.Model
	percent  float64
	width    int
}

// NewAnimatedProgress creates a new animated progress bar.
func NewAnimatedProgress() *AnimatedProgress {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	// Custom gradient colors
	p.FullColor = "#7D56F4"
	p.EmptyColor = "#383838"

	return &AnimatedProgress{
		progress: p,
		percent:  0,
		width:    40,
	}
}

// NewAnimatedProgressWithGradient creates a progress bar with custom gradient.
func NewAnimatedProgressWithGradient(startColor, endColor string) *AnimatedProgress {
	p := progress.New(
		progress.WithGradient(startColor, endColor),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return &AnimatedProgress{
		progress: p,
		percent:  0,
		width:    40,
	}
}

// SetWidth sets the progress bar width.
func (a *AnimatedProgress) SetWidth(width int) {
	a.width = width
	a.progress.Width = width
}

// SetPercent sets the progress percentage (0.0 to 1.0).
// Returns a command for smooth animation.
func (a *AnimatedProgress) SetPercent(percent float64) tea.Cmd {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}
	a.percent = percent
	return a.progress.SetPercent(percent)
}

// Update handles animation frames.
func (a *AnimatedProgress) Update(msg tea.Msg) (*AnimatedProgress, tea.Cmd) {
	var cmd tea.Cmd
	progressModel, cmd := a.progress.Update(msg)
	a.progress = progressModel.(progress.Model)
	return a, cmd
}

// View renders the progress bar.
func (a *AnimatedProgress) View() string {
	return a.progress.View()
}

// ViewWithLabel renders the progress bar with a label.
func (a *AnimatedProgress) ViewWithLabel(label string, percent float64) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginRight(1)

	percentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		MarginLeft(1)

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		labelStyle.Render(label),
		a.progress.ViewAs(percent),
		percentStyle.Render(formatPercent(percent)),
	)
}

// Percent returns the current percentage.
func (a *AnimatedProgress) Percent() float64 {
	return a.percent
}

func formatPercent(p float64) string {
	pct := int(p * 100)
	if pct < 10 {
		return "  " + itoa(pct) + "%"
	} else if pct < 100 {
		return " " + itoa(pct) + "%"
	}
	return itoa(pct) + "%"
}

// ProgressTickMsg is sent to animate progress updates.
type ProgressTickMsg struct {
	ID string
}

// MultiProgress manages multiple animated progress bars.
type MultiProgress struct {
	bars map[string]*AnimatedProgress
}

// NewMultiProgress creates a multi-progress manager.
func NewMultiProgress() *MultiProgress {
	return &MultiProgress{
		bars: make(map[string]*AnimatedProgress),
	}
}

// Get gets or creates a progress bar for the given ID.
func (m *MultiProgress) Get(id string) *AnimatedProgress {
	if bar, ok := m.bars[id]; ok {
		return bar
	}
	bar := NewAnimatedProgress()
	m.bars[id] = bar
	return bar
}

// Remove removes a progress bar.
func (m *MultiProgress) Remove(id string) {
	delete(m.bars, id)
}

// Update handles progress messages for all bars.
func (m *MultiProgress) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for _, bar := range m.bars {
		var cmd tea.Cmd
		bar, cmd = bar.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}
