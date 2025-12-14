// Package player implements media playback with mpv (primary) and VLC (fallback).
package player

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const dialTimeout = 2 * time.Second

// getSocketPath returns platform-specific IPC socket path.
// Linux/macOS: Unix socket in XDG_RUNTIME_DIR or temp
// Windows: Named pipe (note: requires go-winio for full support)
func getSocketPath() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`\\.\pipe\xstream-mpv-%d`, os.Getpid())
	}

	// Prefer XDG_RUNTIME_DIR for security, fallback to temp
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, fmt.Sprintf("xstream-mpv-%d.sock", os.Getpid()))
}

// dialSocket connects to mpv IPC socket with timeout.
// Unix socket on Linux/macOS.
// Note: Windows named pipe support limited - mpv may work but untested.
func dialSocket(path string) (net.Conn, error) {
	return net.DialTimeout("unix", path, dialTimeout)
}

// cleanupSocket removes the socket file (Unix only).
// Note: Socket file permissions controlled by umask and mpv.
// mpv creates socket with restrictive permissions by default.
func cleanupSocket(path string) {
	if runtime.GOOS != "windows" {
		os.Remove(path)
	}
}
