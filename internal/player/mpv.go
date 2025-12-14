package player

import (
	"context"
	"fmt"
	"os/exec"
)

// MPVPlayer controls mpv via IPC socket.
type MPVPlayer struct {
	cmd        *exec.Cmd
	ipc        *IPCClient
	socketPath string
	cancel     context.CancelFunc
}

// NewMPVPlayer spawns mpv with IPC socket and returns a controller.
func NewMPVPlayer(ctx context.Context, streamURL, title string) (*MPVPlayer, error) {
	ctx, cancel := context.WithCancel(ctx)
	socketPath := getSocketPath()

	args := []string{
		"--fs",                             // Fullscreen
		"--input-ipc-server=" + socketPath, // IPC socket
		"--no-terminal",                    // No terminal output
		"--idle=no",                        // Exit when done
		"--title=" + title,                 // Window title (hides URL)
		streamURL,
	}

	cmd := exec.CommandContext(ctx, "mpv", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start mpv: %w", err)
	}

	ipc, err := NewIPCClient(socketPath)
	if err != nil {
		cmd.Process.Kill()
		cancel()
		cleanupSocket(socketPath)
		return nil, fmt.Errorf("connect to mpv: %w", err)
	}

	return &MPVPlayer{
		cmd:        cmd,
		ipc:        ipc,
		socketPath: socketPath,
		cancel:     cancel,
	}, nil
}

// Pause pauses playback.
func (p *MPVPlayer) Pause() error {
	_, err := p.ipc.Command("set", "pause", true)
	return err
}

// Resume resumes playback.
func (p *MPVPlayer) Resume() error {
	_, err := p.ipc.Command("set", "pause", false)
	return err
}

// TogglePause toggles pause state.
func (p *MPVPlayer) TogglePause() error {
	_, err := p.ipc.Command("cycle", "pause")
	return err
}

// Seek seeks relative by seconds (positive = forward, negative = backward).
func (p *MPVPlayer) Seek(seconds int) error {
	_, err := p.ipc.Command("seek", seconds, "relative")
	return err
}

// SetVolume sets volume (0-100).
func (p *MPVPlayer) SetVolume(vol int) error {
	_, err := p.ipc.Command("set", "volume", vol)
	return err
}

// GetPosition returns current playback position in seconds.
func (p *MPVPlayer) GetPosition() (float64, error) {
	resp, err := p.ipc.Command("get_property", "playback-time")
	if err != nil {
		return 0, err
	}
	if f, ok := resp.Data.(float64); ok {
		return f, nil
	}
	return 0, fmt.Errorf("unexpected type for playback-time")
}

// GetDuration returns total duration in seconds.
func (p *MPVPlayer) GetDuration() (float64, error) {
	resp, err := p.ipc.Command("get_property", "duration")
	if err != nil {
		return 0, err
	}
	if f, ok := resp.Data.(float64); ok {
		return f, nil
	}
	return 0, fmt.Errorf("unexpected type for duration")
}

// IsPaused returns whether playback is paused.
func (p *MPVPlayer) IsPaused() (bool, error) {
	resp, err := p.ipc.Command("get_property", "pause")
	if err != nil {
		return false, err
	}
	if b, ok := resp.Data.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("unexpected type for pause")
}

// Wait blocks until mpv exits.
func (p *MPVPlayer) Wait() error {
	return p.cmd.Wait()
}

// Stop terminates mpv and cleans up.
func (p *MPVPlayer) Stop() error {
	// Try graceful quit first
	p.ipc.Command("quit")
	p.ipc.Close()
	p.cancel()
	cleanupSocket(p.socketPath)
	return nil
}
