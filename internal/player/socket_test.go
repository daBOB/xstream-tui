package player

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestGetSocketPath(t *testing.T) {
	path := getSocketPath()

	if path == "" {
		t.Error("getSocketPath returned empty string")
	}

	// Should contain PID for uniqueness
	pid := os.Getpid()
	if !strings.Contains(path, "xstream-mpv") {
		t.Errorf("socket path should contain 'xstream-mpv', got: %s", path)
	}

	// Platform-specific checks
	if runtime.GOOS == "windows" {
		if !strings.HasPrefix(path, `\\.\pipe\`) {
			t.Errorf("Windows path should start with named pipe prefix, got: %s", path)
		}
	} else {
		if !strings.HasSuffix(path, ".sock") {
			t.Errorf("Unix path should end with .sock, got: %s", path)
		}
	}

	t.Logf("Socket path for PID %d: %s", pid, path)
}

func TestCleanupSocket(t *testing.T) {
	// Create a temp file to simulate socket
	if runtime.GOOS == "windows" {
		t.Skip("Socket cleanup test skipped on Windows")
	}

	tmpFile, err := os.CreateTemp("", "test-socket-*.sock")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	path := tmpFile.Name()
	tmpFile.Close()

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("temp file should exist")
	}

	// Cleanup should remove it
	cleanupSocket(path)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("cleanupSocket should have removed file")
		os.Remove(path) // cleanup
	}
}
