// Package components provides reusable TUI components.
package components

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/download"
)

// DownloadQueue displays the download queue as an overlay panel.
type DownloadQueue struct {
	items        []download.Item
	selected     int
	offset       int
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

// getProgressBar returns or creates an animated progress bar.
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

// UpdateProgress updates the progress for a specific download.
func (d *DownloadQueue) UpdateProgress(id string, percent float64) tea.Cmd {
	bar := d.getProgressBar(id)
	cmd := bar.SetPercent(percent)
	d.progressBars[id] = bar
	return cmd
}
