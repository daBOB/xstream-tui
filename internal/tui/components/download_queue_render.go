package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/download"
)

// View renders the download queue as a full-screen view.
func (d *DownloadQueue) View() string {
	if !d.visible {
		return ""
	}

	contentWidth := d.width - 8
	if contentWidth < 40 {
		contentWidth = 40
	}

	progressBarWidth := contentWidth - 30
	if progressBarWidth < 20 {
		progressBarWidth = 20
	}
	if progressBarWidth > 60 {
		progressBarWidth = 60
	}

	for id, bar := range d.progressBars {
		bar.Width = progressBarWidth
		d.progressBars[id] = bar
	}

	styles := d.createStyles(contentWidth)
	var b strings.Builder

	b.WriteString(styles.title.Render("📥 Download Queue"))
	b.WriteString("\n\n")

	if len(d.items) == 0 {
		b.WriteString(styles.dim.Render("No downloads in queue"))
		b.WriteString("\n")
	} else {
		d.renderStats(&b, styles.dim)
		d.renderItems(&b, contentWidth, styles)
	}

	b.WriteString("\n")
	helpKeys := "[↑↓/jk] Navigate  [PgUp/PgDn] Scroll  [g/G] Top/Bottom  [d] Cancel  [x] Remove  [Esc] Close"
	b.WriteString(styles.dim.Render(helpKeys))

	containerStyle := lipgloss.NewStyle().
		Width(d.width).
		Height(d.height).
		Padding(2, 4)

	return containerStyle.Render(b.String())
}

// renderStyles holds styles for download queue rendering.
type renderStyles struct {
	title    lipgloss.Style
	item     lipgloss.Style
	selected lipgloss.Style
	dim      lipgloss.Style
	success  lipgloss.Style
	error    lipgloss.Style
}

// createStyles creates render styles.
func (d *DownloadQueue) createStyles(contentWidth int) renderStyles {
	return renderStyles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			MarginBottom(1),
		item:     lipgloss.NewStyle().Width(contentWidth),
		selected: lipgloss.NewStyle().Width(contentWidth).Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")),
		dim:      lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		success:  lipgloss.NewStyle().Foreground(lipgloss.Color("82")),
		error:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
	}
}

// renderStats renders download statistics.
func (d *DownloadQueue) renderStats(b *strings.Builder, dimStyle lipgloss.Style) {
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

	visible := d.visibleItems()
	if len(d.items) > visible {
		scrollInfo := fmt.Sprintf("  │  Showing %d-%d", d.offset+1, min(d.offset+visible, len(d.items)))
		b.WriteString(dimStyle.Render(stats + scrollInfo))
	} else {
		b.WriteString(dimStyle.Render(stats))
	}
	b.WriteString("\n\n")
}

// renderItems renders the visible queue items.
func (d *DownloadQueue) renderItems(b *strings.Builder, contentWidth int, styles renderStyles) {
	visible := d.visibleItems()
	endIdx := d.offset + visible
	if endIdx > len(d.items) {
		endIdx = len(d.items)
	}

	for i := d.offset; i < endIdx; i++ {
		item := d.items[i]
		style := styles.item
		if i == d.selected {
			style = styles.selected
		}

		icon := statusIcon(item.Status)
		progressStr := d.formatItemProgress(item, styles)

		name := truncate(item.Name, contentWidth-10)
		line := fmt.Sprintf("%s %s", icon, name)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
		b.WriteString("    " + progressStr)
		b.WriteString("\n\n")
	}
}

// formatItemProgress formats the progress display for an item.
func (d *DownloadQueue) formatItemProgress(item download.Item, styles renderStyles) string {
	switch item.Status {
	case download.StatusDownloading:
		bar := d.getProgressBar(item.ID)
		progressStr := bar.ViewAs(item.Progress)
		progressStr += fmt.Sprintf(" %.0f%%", item.Progress*100)
		if item.Size > 0 {
			progressStr += fmt.Sprintf("  (%s / %s)",
				formatBytes(item.Downloaded),
				formatBytes(item.Size))
		}
		return progressStr
	case download.StatusCompleted:
		return styles.success.Render("✓ Complete")
	case download.StatusFailed:
		if item.Error != nil {
			return styles.error.Render("✗ " + truncate(item.Error.Error(), 40))
		}
		return styles.error.Render("✗ Failed")
	case download.StatusCancelled:
		return styles.dim.Render("⊘ Cancelled")
	default:
		return styles.dim.Render("⏳ " + item.Status.String())
	}
}
