package download

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager("/tmp/test-downloads")
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.downloadDir != "/tmp/test-downloads" {
		t.Errorf("expected downloadDir /tmp/test-downloads, got %s", m.downloadDir)
	}
}

func TestAddToQueue(t *testing.T) {
	m := NewManager("/tmp/test-downloads")

	if err := m.Add("test-file.mp4", "http://example.com/file.mp4"); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	queue := m.Queue()
	if len(queue) != 1 {
		t.Fatalf("expected 1 item in queue, got %d", len(queue))
	}

	if queue[0].Name != "test-file.mp4" {
		t.Errorf("expected name 'test-file.mp4', got %s", queue[0].Name)
	}
}

func TestQueueCount(t *testing.T) {
	m := NewManager("/tmp/test-downloads")

	if m.QueueCount() != 0 {
		t.Error("expected empty queue")
	}

	m.Add("file1.mp4", "http://example.com/1")
	m.Add("file2.mp4", "http://example.com/2")

	// Give goroutines time to start
	time.Sleep(50 * time.Millisecond)

	if m.QueueCount() != 2 {
		t.Errorf("expected 2 items, got %d", m.QueueCount())
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal.mp4", "normal.mp4"},
		{"file/with/slashes.mp4", "file_with_slashes.mp4"},
		{"file:with:colons.mp4", "file_with_colons.mp4"},
		{"file*with*stars.mp4", "file_with_stars.mp4"},
		{"", "download"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDownloadWithServer(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "12")
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	// Create temp directory
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Track progress
	var progressUpdates []ProgressUpdate
	m.SetProgressCallback(func(update ProgressUpdate) {
		progressUpdates = append(progressUpdates, update)
	})

	// Add download
	if err := m.Add("test.txt", server.URL); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	// Wait for download to complete
	time.Sleep(500 * time.Millisecond)

	// Check file exists
	expectedPath := filepath.Join(tmpDir, "test.txt")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("downloaded file does not exist")
	}

	// Check content
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("unexpected content: %s", string(content))
	}

	// Check queue status
	queue := m.Queue()
	for _, item := range queue {
		if item.Status != StatusCompleted {
			t.Errorf("expected status Completed, got %s", item.Status.String())
		}
	}
}

func TestCancel(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000000")
		for i := 0; i < 100; i++ {
			time.Sleep(50 * time.Millisecond)
			w.Write(make([]byte, 10000))
		}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	if err := m.Add("slow.bin", server.URL); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	// Wait for download to start
	time.Sleep(100 * time.Millisecond)

	// Cancel first item (dl-1)
	if !m.Cancel("dl-1") {
		t.Error("Cancel returned false")
	}

	// Wait a bit
	time.Sleep(200 * time.Millisecond)

	// Check status
	queue := m.Queue()
	for _, item := range queue {
		if item.Status != StatusCancelled {
			t.Errorf("expected status Cancelled, got %s", item.Status.String())
		}
	}
}

func TestRemove(t *testing.T) {
	m := NewManager("/tmp/test-downloads")

	if err := m.Add("file.mp4", "http://example.com/file"); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	// Can't remove while downloading (give it time to start)
	time.Sleep(50 * time.Millisecond)

	// Cancel first
	m.Cancel("dl-1")
	time.Sleep(50 * time.Millisecond)

	// Now remove
	if !m.Remove("dl-1") {
		t.Error("Remove returned false")
	}

	if m.QueueCount() != 0 {
		t.Error("expected empty queue after remove")
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusQueued, "Queued"},
		{StatusDownloading, "Downloading"},
		{StatusCompleted, "Completed"},
		{StatusFailed, "Failed"},
		{StatusCancelled, "Cancelled"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("Status(%d).String() = %q, want %q", tt.status, tt.status.String(), tt.expected)
		}
	}
}
