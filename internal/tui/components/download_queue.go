// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/download"
)

// DownloadQueue displays the download queue as an overlay panel.
type DownloadQueue struct {
	items      []download.Item
	selected   int
	width      int
	height     int
	visible    bool
	progressBars map[string]progress.Model
}

// NewDownloadQueue creates a new download queue component.
func NewDownloadQueue() *DownloadQueue {
	return &DownloadQueue{
		items:        make([]download.Item, 0),
		visible:      false,
		progressBars: make(map[string]progress.Model),
	}
}

// getProgressBar returns or creates an animated progress bar for the given ID.
func (d *DownloadQueue) getProgressBar(id string) progress.Model {
	if bar, ok := d.progressBars[id]; ok {
		return bar
	}

	bar := progress.New(
		progress.WithGradient("#7D56F4", "#00D9FF"),
		progress.WithWidth(30),
		progress.WithoutPercentage(),
	)
	d.progressBars[id] = bar
	return bar
}

// SetSize sets the component dimensions.
func (d *DownloadQueue) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// SetItems updates the queue items.
func (d *DownloadQueue) SetItems(items []download.Item) {
	d.items = items
	if d.selected >= len(items) && len(items) > 0 {
		d.selected = len(items) - 1
	}

	// Clean up progress bars for removed items
	activeIDs := make(map[string]bool)
	for _, item := range items {
		activeIDs[item.ID] = true
	}
	for id := range d.progressBars {
		if !activeIDs[id] {
			delete(d.progressBars, id)
		}
	}
}

// Toggle toggles visibility.
func (d *DownloadQueue) Toggle() {
	d.visible = !d.visible
}

// Show shows the queue.
func (d *DownloadQueue) Show() {
	d.visible = true
}

// Hide hides the queue.
func (d *DownloadQueue) Hide() {
	d.visible = false
}

// Visible returns whether the queue is visible.
func (d *DownloadQueue) Visible() bool {
	return d.visible
}

// SelectedID returns the ID of the selected item.
func (d *DownloadQueue) SelectedID() string {
	if d.selected >= 0 && d.selected < len(d.items) {
		return d.items[d.selected].ID
	}
	return ""
}

// Update handles keyboard input and progress bar animations.
func (d *DownloadQueue) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	// Update all progress bars for smooth animation
	switch msg.(type) {
	case progress.FrameMsg:
		for id, bar := range d.progressBars {
			newBar, cmd := bar.Update(msg)
			d.progressBars[id] = newBar.(progress.Model)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if !d.visible {
		return tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if d.selected > 0 {
				d.selected--
			}
		case "down", "j":
			if d.selected < len(d.items)-1 {
				d.selected++
			}
		}
	}

	return tea.Batch(cmds...)
}

// UpdateProgress updates the progress for a specific download with animation.
func (d *DownloadQueue) UpdateProgress(id string, percent float64) tea.Cmd {
	bar := d.getProgressBar(id)
	cmd := bar.SetPercent(percent)
	d.progressBars[id] = bar
	return cmd
}

// View renders the download queue panel.
func (d *DownloadQueue) View() string {
	if !d.visible || len(d.items) == 0 {
		return ""
	}

	// Panel styling
	panelWidth := 60
	if d.width > 0 && d.width < panelWidth+10 {
		panelWidth = d.width - 10
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginBottom(1)

	itemStyle := lipgloss.NewStyle().
		Width(panelWidth - 4)

	selectedStyle := lipgloss.NewStyle().
		Width(panelWidth - 4).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	var b strings.Builder

	b.WriteString(titleStyle.Render("Downloads"))
	b.WriteString("\n")

	for i, item := range d.items {
		style := itemStyle
		if i == d.selected {
			style = selectedStyle
		}

		// Status icon
		icon := statusIcon(item.Status)

		// Progress bar for downloading items
		var progressStr string
		if item.Status == download.StatusDownloading {
			bar := d.getProgressBar(item.ID)
			progressStr = bar.ViewAs(item.Progress)
			progressStr += fmt.Sprintf(" %.0f%%", item.Progress*100)

			// Speed/size info
			if item.Size > 0 {
				progressStr += fmt.Sprintf(" (%s / %s)",
					formatBytes(item.Downloaded),
					formatBytes(item.Size))
			}
		} else if item.Status == download.StatusCompleted {
			progressStr = successStyle.Render("✓ Complete")
		} else if item.Status == download.StatusFailed && item.Error != nil {
			progressStr = errorStyle.Render("✗ " + truncate(item.Error.Error(), 25))
		} else if item.Status == download.StatusCancelled {
			progressStr = dimStyle.Render("⊘ Cancelled")
		} else {
			progressStr = dimStyle.Render("⏳ " + item.Status.String())
		}

		// Truncate name
		name := truncate(item.Name, panelWidth-10)

		line := fmt.Sprintf("%s %s", icon, name)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
		b.WriteString("  " + progressStr)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("[d] Cancel  [x] Remove  [Esc] Close"))

	// Panel border
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(panelWidth)

	return panelStyle.Render(b.String())
}

func statusIcon(status download.Status) string {
	switch status {
	case download.StatusQueued:
		return "⏳"
	case download.StatusDownloading:
		return "⬇"
	case download.StatusCompleted:
		return "✓"
	case download.StatusFailed:
		return "✗"
	case download.StatusCancelled:
		return "⊘"
	default:
		return "?"
	}
}

func renderProgressBar(progress float64, width int) string {
	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return "[" + bar + "]"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
