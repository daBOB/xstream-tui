package player

import (
	"context"
	"os/exec"
	"runtime"
)

// VLCPlayer is a fallback player when mpv is unavailable.
// VLC has limited IPC capabilities, so control methods are no-ops.
type VLCPlayer struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

// NewVLCPlayer spawns VLC with fullscreen and play-and-exit mode.
func NewVLCPlayer(ctx context.Context, streamURL, title string) (*VLCPlayer, error) {
	ctx, cancel := context.WithCancel(ctx)

	vlcPath := getVLCPath()
	args := []string{
		"--fullscreen",
		"--play-and-exit",
		"--no-video-title-show",
		"--meta-title=" + title,
		streamURL,
	}

	cmd := exec.CommandContext(ctx, vlcPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}

	return &VLCPlayer{cmd: cmd, cancel: cancel}, nil
}

// getVLCPath returns platform-specific VLC executable path.
func getVLCPath() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Applications/VLC.app/Contents/MacOS/VLC"
	case "windows":
		return "C:\\Program Files\\VideoLAN\\VLC\\vlc.exe"
	default:
		return "vlc"
	}
}

// Wait blocks until VLC exits.
func (p *VLCPlayer) Wait() error {
	return p.cmd.Wait()
}

// Stop terminates VLC.
func (p *VLCPlayer) Stop() error {
	p.cancel()
	return p.cmd.Wait()
}

// VLC has no practical IPC - these are no-ops.

func (p *VLCPlayer) Pause() error       { return nil }
func (p *VLCPlayer) Resume() error      { return nil }
func (p *VLCPlayer) TogglePause() error { return nil }
func (p *VLCPlayer) Seek(int) error     { return nil }
func (p *VLCPlayer) SetVolume(int) error { return nil }
