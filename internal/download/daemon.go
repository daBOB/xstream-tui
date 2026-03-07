// Package download provides download management with daemon-style Bubble Tea integration.
package download

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Daemon tick intervals - adaptive based on activity state.
const (
	DaemonTickIntervalActive = 100 * time.Millisecond  // During active downloads
	DaemonTickIntervalIdle   = 1000 * time.Millisecond // When queue is empty/idle
)

// DaemonTickMsg signals the daemon to check for updates.
type DaemonTickMsg struct{}

// DaemonStatusMsg contains the current download queue status.
type DaemonStatusMsg struct {
	Items       []Item
	ActiveID    string
	ActiveItem  *Item
	QueueCount  int
	Downloading bool
}

// Daemon provides a daemon-style interface for Bubble Tea integration.
// It uses tick messages to poll the download manager and send status updates.
// Adaptive polling: 100ms during downloads, 1000ms when idle.
type Daemon struct {
	manager     *Manager
	running     bool
	wasActive   bool // Track if downloads were active last tick
}

// NewDaemon creates a new download daemon wrapping the manager.
func NewDaemon(manager *Manager) *Daemon {
	return &Daemon{
		manager: manager,
	}
}

// Manager returns the underlying manager.
func (d *Daemon) Manager() *Manager {
	return d.manager
}

// Start returns a command that begins the daemon tick loop.
func (d *Daemon) Start() tea.Cmd {
	d.running = true
	return d.tick(false) // Start with idle interval
}

// Stop stops the daemon.
func (d *Daemon) Stop() {
	d.running = false
}

// tick returns a command that waits then sends a tick message.
// Uses adaptive interval based on download activity.
func (d *Daemon) tick(active bool) tea.Cmd {
	interval := DaemonTickIntervalIdle
	if active {
		interval = DaemonTickIntervalActive
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return DaemonTickMsg{}
	})
}

// Update handles daemon tick messages and returns status updates.
// Returns a status message and a command to continue ticking.
func (d *Daemon) Update(msg tea.Msg) (tea.Msg, tea.Cmd) {
	switch msg.(type) {
	case DaemonTickMsg:
		if !d.running {
			return nil, nil
		}

		// Get current status from manager
		status := d.getStatus()

		// Determine if we have active downloads or queued items
		isActive := status.Downloading || status.QueueCount > 0
		d.wasActive = isActive

		// Continue ticking with adaptive interval
		return status, d.tick(isActive)
	}
	return nil, nil
}

// getStatus collects current queue status.
func (d *Daemon) getStatus() DaemonStatusMsg {
	items := d.manager.Queue()
	active := d.manager.ActiveDownload()

	status := DaemonStatusMsg{
		Items:      items,
		QueueCount: len(items),
	}

	if active != nil {
		status.ActiveID = active.ID
		status.ActiveItem = active
		status.Downloading = true
	}

	return status
}

// Add adds a download to the queue.
func (d *Daemon) Add(name, url string) error {
	return d.manager.Add(name, url)
}

// Cancel cancels a download.
func (d *Daemon) Cancel(id string) bool {
	return d.manager.Cancel(id)
}

// Remove removes a download from the queue.
func (d *Daemon) Remove(id string) bool {
	return d.manager.Remove(id)
}

// DownloadCompleteMsg is sent when a download finishes (success or failure).
type DownloadCompleteMsg struct {
	ID        string
	Name      string
	FilePath  string
	Success   bool
	Error     error
	Cancelled bool
}

// SetupCompletionCallback configures the manager to send completion messages.
func (d *Daemon) SetupCompletionCallback(send func(tea.Msg)) {
	d.manager.SetProgressCallback(func(update ProgressUpdate) {
		// Send completion messages for finished downloads
		if update.Status == StatusCompleted ||
			update.Status == StatusFailed ||
			update.Status == StatusCancelled {

			// Find the item for additional info
			items := d.manager.Queue()
			for _, item := range items {
				if item.ID == update.ID {
					send(DownloadCompleteMsg{
						ID:        update.ID,
						Name:      item.Name,
						FilePath:  item.FilePath,
						Success:   update.Status == StatusCompleted,
						Error:     update.Error,
						Cancelled: update.Status == StatusCancelled,
					})
					break
				}
			}
		}
	})
}
