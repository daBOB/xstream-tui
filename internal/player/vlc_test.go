package player

import (
	"runtime"
	"testing"
)

func TestGetVLCPath(t *testing.T) {
	path := getVLCPath()
	switch runtime.GOOS {
	case "darwin":
		if path != "/Applications/VLC.app/Contents/MacOS/VLC" {
			t.Errorf("darwin path = %q, want macOS path", path)
		}
	case "windows":
		if path != "C:\\Program Files\\VideoLAN\\VLC\\vlc.exe" {
			t.Errorf("windows path = %q, want windows path", path)
		}
	default:
		if path != "vlc" {
			t.Errorf("linux path = %q, want 'vlc'", path)
		}
	}
}

func TestVLCPlayer_ImplementsPlayer(t *testing.T) {
	// VLCPlayer implements the minimal Player interface (Wait, Stop only).
	// Control methods are not available on VLC due to lack of IPC.
	var _ Player = (*VLCPlayer)(nil)
}
