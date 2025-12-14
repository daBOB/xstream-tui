// Package components provides reusable TUI components.
package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/download"
)

// DownloadQueue displays the download queue as an overlay panel.
type DownloadQueue struct {
	items    []download.Item
	selected int
	width    int
	height   int
	visible  bool
}

// NewDownloadQueue creates a new download queue component.
func NewDownloadQueue() *DownloadQueue {
	return &DownloadQueue{
		items:   make([]download.Item, 0),
		visible: false,
	}
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

// Update handles keyboard input.
func (d *DownloadQueue) Update(msg tea.Msg) tea.Cmd {
	if !d.visible {
		return nil
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
	return nil
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
		var progress string
		if item.Status == download.StatusDownloading {
			progress = renderProgressBar(item.Progress, 20)
			progress += fmt.Sprintf(" %.0f%%", item.Progress*100)
		} else if item.Status == download.StatusCompleted {
			progress = "Complete"
		} else if item.Status == download.StatusFailed && item.Error != nil {
			progress = "Error: " + truncate(item.Error.Error(), 20)
		} else {
			progress = item.Status.String()
		}

		// Truncate name
		name := truncate(item.Name, panelWidth-10)

		line := fmt.Sprintf("%s %s", icon, name)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  " + progress))
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
