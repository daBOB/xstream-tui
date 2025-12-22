package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/download"
)

func TestNewDownloadQueue(t *testing.T) {
	dq := NewDownloadQueue()
	if dq == nil {
		t.Fatal("NewDownloadQueue() returned nil")
	}
	if dq.visible {
		t.Error("new queue should be hidden")
	}
	if dq.progressBars == nil {
		t.Error("progressBars should be initialized")
	}
}

func TestDownloadQueue_Visibility(t *testing.T) {
	dq := NewDownloadQueue()

	if dq.Visible() {
		t.Error("should start hidden")
	}

	dq.Show()
	if !dq.Visible() {
		t.Error("should be visible after Show()")
	}

	dq.Hide()
	if dq.Visible() {
		t.Error("should be hidden after Hide()")
	}

	dq.Toggle()
	if !dq.Visible() {
		t.Error("should be visible after Toggle()")
	}

	dq.Toggle()
	if dq.Visible() {
		t.Error("should be hidden after second Toggle()")
	}
}

func TestDownloadQueue_SetItems(t *testing.T) {
	dq := NewDownloadQueue()

	items := []download.Item{
		{ID: "1", Name: "file1.mp4", Status: download.StatusQueued},
		{ID: "2", Name: "file2.mp4", Status: download.StatusDownloading},
	}

	dq.SetItems(items)

	// Selection should be valid
	if dq.selected < 0 || dq.selected >= len(items) {
		t.Errorf("selected = %d, out of range", dq.selected)
	}
}

func TestDownloadQueue_SetItems_AdjustsSelection(t *testing.T) {
	dq := NewDownloadQueue()

	// Add 3 items
	items := []download.Item{
		{ID: "1"}, {ID: "2"}, {ID: "3"},
	}
	dq.SetItems(items)
	dq.selected = 2 // Select last

	// Reduce to 1 item
	dq.SetItems([]download.Item{{ID: "1"}})

	if dq.selected != 0 {
		t.Errorf("selected should be adjusted to 0, got %d", dq.selected)
	}
}

func TestDownloadQueue_SelectedID(t *testing.T) {
	dq := NewDownloadQueue()

	// Empty queue
	if id := dq.SelectedID(); id != "" {
		t.Errorf("empty queue SelectedID = %q, want ''", id)
	}

	// With items
	dq.SetItems([]download.Item{{ID: "test-id"}})
	if id := dq.SelectedID(); id != "test-id" {
		t.Errorf("SelectedID = %q, want 'test-id'", id)
	}
}

func TestDownloadQueue_SetSize(t *testing.T) {
	dq := NewDownloadQueue()
	dq.SetSize(100, 50)

	if dq.width != 100 {
		t.Errorf("width = %d, want 100", dq.width)
	}
	if dq.height != 50 {
		t.Errorf("height = %d, want 50", dq.height)
	}
}

func TestDownloadQueue_Update_Navigation(t *testing.T) {
	dq := NewDownloadQueue()
	dq.Show()
	dq.SetSize(80, 40)
	dq.SetItems([]download.Item{
		{ID: "1"}, {ID: "2"}, {ID: "3"},
	})

	t.Run("down key", func(t *testing.T) {
		dq.selected = 0
		dq.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if dq.selected != 1 {
			t.Errorf("selected = %d after down, want 1", dq.selected)
		}
	})

	t.Run("up key", func(t *testing.T) {
		dq.selected = 1
		dq.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if dq.selected != 0 {
			t.Errorf("selected = %d after up, want 0", dq.selected)
		}
	})

	t.Run("home key", func(t *testing.T) {
		dq.selected = 2
		dq.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
		if dq.selected != 0 {
			t.Errorf("selected = %d after home, want 0", dq.selected)
		}
	})

	t.Run("end key", func(t *testing.T) {
		dq.selected = 0
		dq.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
		if dq.selected != 2 {
			t.Errorf("selected = %d after end, want 2", dq.selected)
		}
	})
}

func TestDownloadQueue_View_Empty(t *testing.T) {
	dq := NewDownloadQueue()
	dq.Show()
	dq.SetSize(80, 40)

	view := dq.View()
	if view == "" {
		t.Error("View() returned empty for visible queue")
	}
}

func TestDownloadQueue_View_Hidden(t *testing.T) {
	dq := NewDownloadQueue()
	// Default is hidden

	view := dq.View()
	if view != "" {
		t.Error("View() should return empty string when hidden")
	}
}

func TestDownloadQueue_View_WithItems(t *testing.T) {
	dq := NewDownloadQueue()
	dq.Show()
	dq.SetSize(80, 40)
	dq.SetItems([]download.Item{
		{ID: "1", Name: "test.mp4", Status: download.StatusDownloading, Progress: 0.5},
		{ID: "2", Name: "done.mp4", Status: download.StatusCompleted},
		{ID: "3", Name: "failed.mp4", Status: download.StatusFailed},
	})

	view := dq.View()
	if view == "" {
		t.Error("View() returned empty with items")
	}
}

func TestDownloadQueue_UpdateProgress(t *testing.T) {
	dq := NewDownloadQueue()
	cmd := dq.UpdateProgress("test-id", 0.5)

	// Should create a progress bar
	if _, ok := dq.progressBars["test-id"]; !ok {
		t.Error("progress bar should be created for new ID")
	}

	// cmd may be nil or return animation frame
	_ = cmd
}

func TestDownloadQueue_Helpers(t *testing.T) {
	t.Run("truncate", func(t *testing.T) {
		if got := truncate("short", 10); got != "short" {
			t.Errorf("truncate short = %q", got)
		}
		if got := truncate("this is a long string", 10); got != "this is..." {
			t.Errorf("truncate long = %q", got)
		}
		if got := truncate("ab", 2); got != "ab" {
			t.Errorf("truncate exact = %q", got)
		}
	})

	t.Run("formatBytes", func(t *testing.T) {
		tests := []struct {
			bytes int64
			want  string
		}{
			{500, "500 B"},
			{1024, "1.0 KB"},
			{1024 * 1024, "1.0 MB"},
			{1024 * 1024 * 1024, "1.0 GB"},
		}
		for _, tt := range tests {
			if got := formatBytes(tt.bytes); got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		}
	})

	t.Run("statusIcon", func(t *testing.T) {
		if got := statusIcon(download.StatusQueued); got != "⏳" {
			t.Errorf("queued icon = %q", got)
		}
		if got := statusIcon(download.StatusDownloading); got != "⬇" {
			t.Errorf("downloading icon = %q", got)
		}
		if got := statusIcon(download.StatusCompleted); got != "✓" {
			t.Errorf("completed icon = %q", got)
		}
		if got := statusIcon(download.StatusFailed); got != "✗" {
			t.Errorf("failed icon = %q", got)
		}
		if got := statusIcon(download.StatusCancelled); got != "⊘" {
			t.Errorf("cancelled icon = %q", got)
		}
	})
}
