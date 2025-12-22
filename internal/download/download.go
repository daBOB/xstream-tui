package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// processQueue finds and starts the next queued download.
func (m *Manager) processQueue() {
	m.mu.Lock()

	if m.isProcessing {
		m.mu.Unlock()
		return
	}

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

	go func() {
		err := m.download(ctx, next)
		m.mu.Lock()
		m.isProcessing = false
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

		m.processQueue()
	}()
}

// download performs the actual file download.
func (m *Manager) download(ctx context.Context, item *Item) error {
	fileDir := filepath.Dir(item.FilePath)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
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

	m.mu.Lock()
	item.Size = resp.ContentLength
	m.mu.Unlock()

	tmpPath := item.FilePath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if err := m.copyWithProgress(ctx, file, resp.Body, item, tmpPath); err != nil {
		return err
	}

	file.Close()

	if err := os.Rename(tmpPath, item.FilePath); err != nil {
		return fmt.Errorf("rename file: %w", err)
	}

	return nil
}

// copyWithProgress copies data while tracking progress.
func (m *Manager) copyWithProgress(ctx context.Context, dst *os.File, src io.Reader, item *Item, tmpPath string) error {
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			dst.Close()
			os.Remove(tmpPath)
			return ctx.Err()
		default:
		}

		n, readErr := src.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				dst.Close()
				os.Remove(tmpPath)
				return fmt.Errorf("write file: %w", writeErr)
			}

			m.updateProgress(item, int64(n))
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			dst.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("read body: %w", readErr)
		}
	}
	return nil
}

// updateProgress updates item progress and notifies callback.
func (m *Manager) updateProgress(item *Item, bytesRead int64) {
	m.mu.Lock()
	item.Downloaded += bytesRead
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

// notifyProgress calls the progress callback if set.
func (m *Manager) notifyProgress(update ProgressUpdate) {
	m.mu.RLock()
	cb := m.onProgress
	m.mu.RUnlock()

	if cb != nil {
		cb(update)
	}
}
