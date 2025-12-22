package player

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}

	if m.IsPlaying() {
		t.Error("new manager should not be playing")
	}

	if m.Type() != "" {
		t.Errorf("new manager should have empty type, got: %s", m.Type())
	}
}

func TestManagerControlsWithoutPlayer(t *testing.T) {
	m := NewManager()

	// These should not panic when no player is active
	if err := m.TogglePause(); err != nil {
		t.Errorf("TogglePause with no player should return nil, got: %v", err)
	}

	if err := m.Seek(10); err != nil {
		t.Errorf("Seek with no player should return nil, got: %v", err)
	}

	if err := m.SetVolume(50); err != nil {
		t.Errorf("SetVolume with no player should return nil, got: %v", err)
	}

	if err := m.Stop(); err != nil {
		t.Errorf("Stop with no player should return nil, got: %v", err)
	}
}

func TestManagerOnExit(t *testing.T) {
	m := NewManager()

	called := false
	m.OnExit(func(err error) {
		called = true
	})

	// The callback is stored but won't be called until playback ends
	// Just verify it doesn't panic
	if called {
		t.Error("callback should not be called before playback")
	}
}

func TestPlayerInterfaces(t *testing.T) {
	// VLCPlayer implements minimal Player interface (no control methods)
	var _ Player = (*VLCPlayer)(nil)

	// MPVPlayer implements ControllablePlayer (full control via IPC)
	var _ ControllablePlayer = (*MPVPlayer)(nil)
}

func TestManagerIsControllable(t *testing.T) {
	m := NewManager()

	// No player = not controllable
	if m.IsControllable() {
		t.Error("manager with no player should not be controllable")
	}
}
