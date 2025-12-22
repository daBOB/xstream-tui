package player

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Player is the minimal interface for media player implementations.
// All players must implement Wait and Stop.
type Player interface {
	Wait() error
	Stop() error
}

// ControllablePlayer extends Player with playback control capabilities.
// Only players with IPC support (like mpv) implement this interface.
type ControllablePlayer interface {
	Player
	Pause() error
	Resume() error
	TogglePause() error
	Seek(seconds int) error
	SetVolume(vol int) error
}

// PlayerType indicates which player backend is active.
type PlayerType string

const (
	TypeMPV PlayerType = "mpv"
	TypeVLC PlayerType = "vlc"
)

// Manager coordinates player lifecycle and provides unified control.
type Manager struct {
	mu         sync.Mutex
	current    Player
	playerType PlayerType
	onExit     func(error)
	playing    bool
	done       chan struct{} // signals monitor goroutine finished
}

// NewManager creates a new player manager.
func NewManager() *Manager {
	return &Manager{}
}

// Play starts playback with mpv (preferred) or VLC (fallback).
func (m *Manager) Play(ctx context.Context, url, title string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop any existing playback
	if m.current != nil {
		m.stopLocked()
	}

	// Try mpv first
	player, err := NewMPVPlayer(ctx, url, title)
	if err == nil {
		m.current = player
		m.playerType = TypeMPV
		m.playing = true
		m.done = make(chan struct{})
		m.monitorExitLocked()
		return nil
	}

	// Fallback to VLC
	vlcPlayer, vlcErr := NewVLCPlayer(ctx, url, title)
	if vlcErr != nil {
		return fmt.Errorf("no player available: mpv (%v), vlc (%v)", err, vlcErr)
	}

	m.current = vlcPlayer
	m.playerType = TypeVLC
	m.playing = true
	m.done = make(chan struct{})
	m.monitorExitLocked()
	return nil
}

// monitorExitLocked watches for player exit and calls callback.
// Must be called while holding m.mu.
func (m *Manager) monitorExitLocked() {
	player := m.current
	done := m.done

	go func() {
		defer close(done)

		err := player.Wait()

		m.mu.Lock()
		// Only update if this is still the active player
		if m.current == player {
			m.playing = false
		}
		callback := m.onExit
		m.mu.Unlock()

		if callback != nil {
			callback(err)
		}
	}()
}

// stopLocked terminates current playback (must hold m.mu).
func (m *Manager) stopLocked() error {
	if m.current == nil {
		return nil
	}

	err := m.current.Stop()
	done := m.done
	m.current = nil
	m.playing = false

	// Wait for monitor goroutine with timeout to prevent deadlock
	if done != nil {
		m.mu.Unlock()
		select {
		case <-done:
			// Monitor goroutine finished cleanly
		case <-time.After(5 * time.Second):
			// Timeout: monitor goroutine hung, continue without waiting
		}
		m.mu.Lock()
	}

	return err
}

// Stop terminates current playback.
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopLocked()
}

// TogglePause toggles pause state (only works if player is controllable).
func (m *Manager) TogglePause() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cp, ok := m.current.(ControllablePlayer); ok {
		return cp.TogglePause()
	}
	return nil
}

// Seek seeks by relative seconds (only works if player is controllable).
func (m *Manager) Seek(seconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cp, ok := m.current.(ControllablePlayer); ok {
		return cp.Seek(seconds)
	}
	return nil
}

// SetVolume sets player volume (0-100) (only works if player is controllable).
func (m *Manager) SetVolume(vol int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cp, ok := m.current.(ControllablePlayer); ok {
		return cp.SetVolume(vol)
	}
	return nil
}

// OnExit registers a callback for when playback ends.
func (m *Manager) OnExit(fn func(error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onExit = fn
}

// IsPlaying returns whether a player is currently active.
func (m *Manager) IsPlaying() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.playing
}

// Type returns the current player type (mpv or vlc).
func (m *Manager) Type() PlayerType {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.playerType
}

// IsControllable returns whether the current player supports playback control.
// VLC lacks IPC, so only mpv is controllable.
func (m *Manager) IsControllable() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.current.(ControllablePlayer)
	return ok
}
