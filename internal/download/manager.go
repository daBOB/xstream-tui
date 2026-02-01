// Package download provides a download manager with queue functionality.
// Only one download runs at a time; others wait in queue.
package download

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

// ProgressThrottleInterval limits progress update frequency to avoid UI flooding.
const ProgressThrottleInterval = 100 * time.Millisecond

// Item represents a download in the queue.
type Item struct {
	ID               string
	Name             string
	URL              string
	FilePath         string
	Status           Status
	Progress         float64
	Size             int64
	Downloaded       int64
	Error            error
	StartedAt        time.Time
	cancel           func()
	lastProgressTime time.Time // For throttling progress updates
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
	isProcessing bool
}

// NewManager creates a download manager with specified download directory.
func NewManager(downloadDir string) *Manager {
	return &Manager{
		queue:       make([]*Item, 0),
		downloadDir: downloadDir,
		httpClient:  &http.Client{Timeout: 0},
		nextID:      1,
	}
}

// SetProgressCallback sets the callback for progress updates.
func (m *Manager) SetProgressCallback(cb func(ProgressUpdate)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onProgress = cb
}

// Add adds a new download to the queue.
func (m *Manager) Add(name, urlStr string) string {
	return m.AddWithPath(name, urlStr, "")
}

// AddWithPath adds a download with optional subpath.
func (m *Manager) AddWithPath(name, urlStr, subPath string) string {
	if err := validateURL(urlStr); err != nil {
		return ""
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("dl-%d", m.nextID)
	m.nextID++

	safeName := sanitizeFilename(name)

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
	go m.processQueue()

	return id
}

// validateURL checks that URL uses http or https scheme.
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

// sanitizeFilename replaces problematic characters in filenames.
func sanitizeFilename(name string) string {
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

// sanitizePath sanitizes a path with multiple segments.
func sanitizePath(path string) string {
	segments := strings.Split(path, "/")
	sanitized := make([]string, 0, len(segments))

	for _, seg := range segments {
		if seg == "" {
			continue
		}
		safeSeg := sanitizeFilename(seg)
		sanitized = append(sanitized, safeSeg)
	}

	return filepath.Join(sanitized...)
}
