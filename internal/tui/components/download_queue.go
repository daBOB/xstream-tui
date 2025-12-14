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
	items        []download.Item
	selected     int
	offset       int // scroll offset for viewport
	width        int
	height       int
	visible      bool
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
	if d.selected < 0 {
		d.selected = 0
	}
	// Adjust offset to ensure selection is visible
	d.adjustOffset()

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
			d.moveUp(1)
		case "down", "j":
			d.moveDown(1)
		case "pgup", "ctrl+u":
			d.moveUp(d.visibleItems() / 2)
		case "pgdown", "ctrl+d":
			d.moveDown(d.visibleItems() / 2)
		case "home", "g":
			d.selected = 0
			d.offset = 0
		case "end", "G":
			if len(d.items) > 0 {
				d.selected = len(d.items) - 1
				d.adjustOffset()
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

// visibleItems returns how many items can fit in the viewport.
// Each item takes 3 lines (name, progress, blank line).
func (d *DownloadQueue) visibleItems() int {
	// Account for header (title + stats + spacing) ~6 lines and footer help ~2 lines
	availableHeight := d.height - 10
	if availableHeight < 3 {
		return 1
	}
	// Each item = 3 lines (name + progress + blank)
	return availableHeight / 3
}

// moveUp moves selection up by n items and adjusts scroll offset.
func (d *DownloadQueue) moveUp(n int) {
	if d.selected > 0 {
		d.selected -= n
		if d.selected < 0 {
			d.selected = 0
		}
		if d.selected < d.offset {
			d.offset = d.selected
		}
	}
}

// moveDown moves selection down by n items and adjusts scroll offset.
func (d *DownloadQueue) moveDown(n int) {
	if len(d.items) == 0 {
		return
	}
	d.selected += n
	if d.selected >= len(d.items) {
		d.selected = len(d.items) - 1
	}
	d.adjustOffset()
}

// adjustOffset ensures the selected item is visible in the viewport.
func (d *DownloadQueue) adjustOffset() {
	visible := d.visibleItems()
	if visible <= 0 {
		visible = 1
	}
	if d.selected >= d.offset+visible {
		d.offset = d.selected - visible + 1
	}
	if d.offset < 0 {
		d.offset = 0
	}
}

// View renders the download queue as a full-screen view.
func (d *DownloadQueue) View() string {
	if !d.visible {
		return ""
	}

	// Use full width with padding
	contentWidth := d.width - 8
	if contentWidth < 40 {
		contentWidth = 40
	}

	// Progress bar width scales with screen
	progressBarWidth := contentWidth - 30
	if progressBarWidth < 20 {
		progressBarWidth = 20
	}
	if progressBarWidth > 60 {
		progressBarWidth = 60
	}

	// Update progress bar widths
	for id, bar := range d.progressBars {
		bar.Width = progressBarWidth
		d.progressBars[id] = bar
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginBottom(1)

	itemStyle := lipgloss.NewStyle().
		Width(contentWidth)

	selectedStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("📥 Download Queue"))
	b.WriteString("\n\n")

	if len(d.items) == 0 {
		b.WriteString(dimStyle.Render("No downloads in queue"))
		b.WriteString("\n")
	} else {
		// Stats
		queued, downloading, completed, failed := 0, 0, 0, 0
		for _, item := range d.items {
			switch item.Status {
			case download.StatusQueued:
				queued++
			case download.StatusDownloading:
				downloading++
			case download.StatusCompleted:
				completed++
			case download.StatusFailed, download.StatusCancelled:
				failed++
			}
		}
		stats := fmt.Sprintf("Total: %d  │  Queued: %d  │  Downloading: %d  │  Completed: %d  │  Failed: %d",
			len(d.items), queued, downloading, completed, failed)
		// Scroll indicator
		visible := d.visibleItems()
		if len(d.items) > visible {
			scrollInfo := fmt.Sprintf("  │  Showing %d-%d", d.offset+1, min(d.offset+visible, len(d.items)))
			b.WriteString(dimStyle.Render(stats + scrollInfo))
		} else {
			b.WriteString(dimStyle.Render(stats))
		}
		b.WriteString("\n\n")

		// Items - only render visible portion
		endIdx := d.offset + visible
		if endIdx > len(d.items) {
			endIdx = len(d.items)
		}

		for i := d.offset; i < endIdx; i++ {
			item := d.items[i]
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
					progressStr += fmt.Sprintf("  (%s / %s)",
						formatBytes(item.Downloaded),
						formatBytes(item.Size))
				}
			} else if item.Status == download.StatusCompleted {
				progressStr = successStyle.Render("✓ Complete")
			} else if item.Status == download.StatusFailed && item.Error != nil {
				progressStr = errorStyle.Render("✗ " + truncate(item.Error.Error(), 40))
			} else if item.Status == download.StatusCancelled {
				progressStr = dimStyle.Render("⊘ Cancelled")
			} else {
				progressStr = dimStyle.Render("⏳ " + item.Status.String())
			}

			// Truncate name to fit screen
			name := truncate(item.Name, contentWidth-10)

			line := fmt.Sprintf("%s %s", icon, name)
			b.WriteString(style.Render(line))
			b.WriteString("\n")
			b.WriteString("    " + progressStr)
			b.WriteString("\n\n")
		}
	}

	// Help text at bottom
	b.WriteString("\n")
	helpKeys := "[↑↓/jk] Navigate  [PgUp/PgDn] Scroll  [g/G] Top/Bottom  [d] Cancel  [x] Remove  [Esc] Close"
	b.WriteString(dimStyle.Render(helpKeys))

	// Full screen container with padding
	containerStyle := lipgloss.NewStyle().
		Width(d.width).
		Height(d.height).
		Padding(2, 4)

	return containerStyle.Render(b.String())
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
