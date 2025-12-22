package components

import (
	"fmt"

	"github.com/altmueller/xstream-tui/internal/download"
)

// statusIcon returns an emoji icon for the download status.
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

// truncate shortens a string to maxLen with ellipsis.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatBytes converts bytes to human-readable format.
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

// visibleItems returns how many items can fit in the viewport.
func (d *DownloadQueue) visibleItems() int {
	availableHeight := d.height - 10
	if availableHeight < 3 {
		return 1
	}
	return availableHeight / 3
}

// moveUp moves selection up by n items.
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

// moveDown moves selection down by n items.
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

// adjustOffset ensures the selected item is visible.
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
