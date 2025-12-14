package player

import (
	"os/exec"
	"runtime"
)

// Availability reports which players are installed.
type Availability struct {
	MPV bool
	VLC bool
}

// Detect checks for available media players.
func Detect() Availability {
	return Availability{
		MPV: detectMPV(),
		VLC: detectVLC(),
	}
}

func detectMPV() bool {
	_, err := exec.LookPath("mpv")
	return err == nil
}

func detectVLC() bool {
	switch runtime.GOOS {
	case "darwin":
		// macOS: Check standard application location
		_, err := exec.LookPath("/Applications/VLC.app/Contents/MacOS/VLC")
		return err == nil
	case "windows":
		// Windows: Check common install paths
		paths := []string{
			"C:\\Program Files\\VideoLAN\\VLC\\vlc.exe",
			"C:\\Program Files (x86)\\VideoLAN\\VLC\\vlc.exe",
		}
		for _, p := range paths {
			if _, err := exec.LookPath(p); err == nil {
				return true
			}
		}
		// Also check PATH
		_, err := exec.LookPath("vlc")
		return err == nil
	default:
		// Linux/BSD: Check PATH
		_, err := exec.LookPath("vlc")
		return err == nil
	}
}

// Preferred returns the best available player type.
func (a Availability) Preferred() PlayerType {
	if a.MPV {
		return TypeMPV
	}
	if a.VLC {
		return TypeVLC
	}
	return ""
}

// String returns a human-readable availability summary.
func (a Availability) String() string {
	switch {
	case a.MPV && a.VLC:
		return "mpv, vlc"
	case a.MPV:
		return "mpv"
	case a.VLC:
		return "vlc"
	default:
		return "none"
	}
}
