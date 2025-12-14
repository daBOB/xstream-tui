// Package download provides a download manager with queue functionality.
// Only one download runs at a time; others wait in queue.
// Files are downloaded to a user-specified directory with sanitized filenames.
// Existing files with the same name will be overwritten.
package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ErrInvalidURL is returned when the URL scheme is not http or https.
var ErrInvalidURL = errors.New("invalid URL: only http and https are supported")

// Status represents the download state.
type Status int

const (
	StatusQueued Status = iota
	StatusDownloading
	StatusCompleted
	StatusFailed
	StatusCancelled
)

func (s Status) String() string {
	switch s {
	case StatusQueued:
		return "Queued"
	case StatusDownloading:
		return "Downloading"
	case StatusCompleted:
		return "Completed"
	case StatusFailed:
		return "Failed"
	case StatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// Item represents a download in the queue.
type Item struct {
	ID         string
	Name       string
	URL        string
	FilePath   string
	Status     Status
	Progress   float64 // 0.0 to 1.0
	Size       int64   // total bytes
	Downloaded int64   // bytes downloaded
	Error      error
	StartedAt  time.Time
	cancel     context.CancelFunc
}

// ProgressUpdate is sent when download progress changes.
type ProgressUpdate struct {
	ID         string
	Progress   float64
	Downloaded int64
	Size       int64
	Status     Status
	Error      error
}

// Manager handles the download queue with single concurrent download.
type Manager struct {
	mu           sync.RWMutex
	queue        []*Item
	downloadDir  string
	httpClient   *http.Client
	onProgress   func(ProgressUpdate)
	nextID       int
	isProcessing bool // prevents concurrent processQueue execution
}

// NewManager creates a download manager with specified download directory.
func NewManager(downloadDir string) *Manager {
	return &Manager{
		queue:       make([]*Item, 0),
		downloadDir: downloadDir,
		httpClient: &http.Client{
			Timeout: 0, // No timeout for downloads
		},
		nextID: 1,
	}
}

// SetProgressCallback sets the callback for progress updates.
func (m *Manager) SetProgressCallback(cb func(ProgressUpdate)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onProgress = cb
}

// Add adds a new download to the queue.
// Returns empty string and sets error if URL is invalid.
func (m *Manager) Add(name, urlStr string) string {
	return m.AddWithPath(name, urlStr, "")
}

// AddWithPath adds a download with optional subpath (e.g., "SeriesName/Season 1").
// Returns empty string and sets error if URL is invalid.
func (m *Manager) AddWithPath(name, urlStr, subPath string) string {
	// Validate URL before adding to queue
	if err := validateURL(urlStr); err != nil {
		return ""
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("dl-%d", m.nextID)
	m.nextID++

	// Sanitize filename
	safeName := sanitizeFilename(name)

	// Build file path with optional subfolder
	var filePath string
	if subPath != "" {
		safeSubPath := sanitizePath(subPath)
		filePath = filepath.Join(m.downloadDir, safeSubPath, safeName)
	} else {
		filePath = filepath.Join(m.downloadDir, safeName)
	}

	item := &Item{
		ID:       id,
		Name:     name,
		URL:      urlStr,
		FilePath: filePath,
		Status:   StatusQueued,
	}

	m.queue = append(m.queue, item)

	// Start processing if this is the only item
	go m.processQueue()

	return id
}

// validateURL checks that URL is valid and uses http or https scheme.
func validateURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrInvalidURL
	}
	return nil
}

// Cancel cancels a download by ID.
func (m *Manager) Cancel(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range m.queue {
		if item.ID == id {
			if item.Status == StatusDownloading && item.cancel != nil {
				item.cancel()
			}
			item.Status = StatusCancelled
			return true
		}
	}
	return false
}

// Remove removes a completed/failed/cancelled download from queue.
func (m *Manager) Remove(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, item := range m.queue {
		if item.ID == id && item.Status != StatusDownloading {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			return true
		}
	}
	return false
}

// Queue returns a copy of the current queue.
func (m *Manager) Queue() []Item {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Item, len(m.queue))
	for i, item := range m.queue {
		result[i] = *item
	}
	return result
}

// QueueCount returns the number of items in queue.
func (m *Manager) QueueCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.queue)
}

// ActiveDownload returns the currently downloading item, if any.
func (m *Manager) ActiveDownload() *Item {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, item := range m.queue {
		if item.Status == StatusDownloading {
			copied := *item
			return &copied
		}
	}
	return nil
}

func (m *Manager) processQueue() {
	m.mu.Lock()

	// Check if already processing (prevents race condition)
	if m.isProcessing {
		m.mu.Unlock()
		return
	}

	// Find next queued item
	var next *Item
	for _, item := range m.queue {
		if item.Status == StatusQueued {
			next = item
			break
		}
	}

	if next == nil {
		m.mu.Unlock()
		return
	}

	m.isProcessing = true
	next.Status = StatusDownloading
	next.StartedAt = time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	next.cancel = cancel

	m.mu.Unlock()

	// Download in goroutine
	go func() {
		err := m.download(ctx, next)
		m.mu.Lock()
		m.isProcessing = false // Reset processing flag
		if err != nil {
			if ctx.Err() == context.Canceled {
				next.Status = StatusCancelled
			} else {
				next.Status = StatusFailed
				next.Error = err
			}
		} else {
			next.Status = StatusCompleted
			next.Progress = 1.0
		}
		m.mu.Unlock()

		m.notifyProgress(ProgressUpdate{
			ID:         next.ID,
			Progress:   next.Progress,
			Downloaded: next.Downloaded,
			Size:       next.Size,
			Status:     next.Status,
			Error:      next.Error,
		})

		// Process next item
		m.processQueue()
	}()
}

func (m *Manager) download(ctx context.Context, item *Item) error {
	// Ensure download directory exists
	if err := os.MkdirAll(m.downloadDir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, item.URL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status: %d", resp.StatusCode)
	}

	// Get total size
	m.mu.Lock()
	item.Size = resp.ContentLength
	m.mu.Unlock()

	// Create temp file
	tmpPath := item.FilePath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	// Download with progress tracking
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		select {
		case <-ctx.Done():
			file.Close()
			os.Remove(tmpPath)
			return ctx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				file.Close()
				os.Remove(tmpPath)
				return fmt.Errorf("write file: %w", writeErr)
			}

			m.mu.Lock()
			item.Downloaded += int64(n)
			if item.Size > 0 {
				item.Progress = float64(item.Downloaded) / float64(item.Size)
			}
			downloaded := item.Downloaded
			size := item.Size
			progress := item.Progress
			m.mu.Unlock()

			m.notifyProgress(ProgressUpdate{
				ID:         item.ID,
				Progress:   progress,
				Downloaded: downloaded,
				Size:       size,
				Status:     StatusDownloading,
			})
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			file.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("read body: %w", readErr)
		}
	}

	file.Close()

	// Rename temp file to final name
	if err := os.Rename(tmpPath, item.FilePath); err != nil {
		return fmt.Errorf("rename file: %w", err)
	}

	return nil
}

func (m *Manager) notifyProgress(update ProgressUpdate) {
	m.mu.RLock()
	cb := m.onProgress
	m.mu.RUnlock()

	if cb != nil {
		cb(update)
	}
}

func sanitizeFilename(name string) string {
	// Replace problematic characters
	result := make([]rune, 0, len(name))
	for _, r := range name {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			result = append(result, '_')
		default:
			result = append(result, r)
		}
	}

	s := string(result)
	if len(s) > 200 {
		s = s[:200]
	}
	if s == "" {
		s = "download"
	}
	return s
}

// sanitizePath sanitizes a path with multiple segments (e.g., "Series/Season 1").
// Preserves path separators but sanitizes each segment.
func sanitizePath(path string) string {
	// Split by path separators
	segments := strings.Split(path, "/")
	sanitized := make([]string, 0, len(segments))

	for _, seg := range segments {
		if seg == "" {
			continue
		}
		// Sanitize each segment like a filename
		safeSeg := sanitizeFilename(seg)
		sanitized = append(sanitized, safeSeg)
	}

	return filepath.Join(sanitized...)
}
