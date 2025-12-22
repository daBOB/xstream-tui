package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
)

func TestNewContentTypeModel(t *testing.T) {
	m := NewContentTypeModel()
	if m == nil {
		t.Fatal("NewContentTypeModel() returned nil")
	}
	if len(m.options) != 3 {
		t.Errorf("options count = %d, want 3", len(m.options))
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestContentTypeModel_SetSize(t *testing.T) {
	m := NewContentTypeModel()
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestContentTypeModel_SetUserInfo(t *testing.T) {
	m := NewContentTypeModel()
	m.SetUserInfo("test-user")

	if m.userInfo != "test-user" {
		t.Errorf("userInfo = %q, want 'test-user'", m.userInfo)
	}
}

func TestContentTypeModel_Selected(t *testing.T) {
	m := NewContentTypeModel()

	// Default selection
	if m.Selected() != tui.LiveContent {
		t.Errorf("Selected() = %v, want LiveContent", m.Selected())
	}

	// After navigation
	m.cursor = 1
	if m.Selected() != tui.VODContent {
		t.Errorf("Selected() = %v, want VODContent", m.Selected())
	}

	m.cursor = 2
	if m.Selected() != tui.SeriesContent {
		t.Errorf("Selected() = %v, want SeriesContent", m.Selected())
	}
}

func TestContentTypeModel_Update_Navigation(t *testing.T) {
	m := NewContentTypeModel()

	t.Run("down key", func(t *testing.T) {
		m.cursor = 0
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if m.cursor != 1 {
			t.Errorf("cursor = %d after down, want 1", m.cursor)
		}
	})

	t.Run("up key", func(t *testing.T) {
		m.cursor = 1
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if m.cursor != 0 {
			t.Errorf("cursor = %d after up, want 0", m.cursor)
		}
	})

	t.Run("boundary at top", func(t *testing.T) {
		m.cursor = 0
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if m.cursor != 0 {
			t.Errorf("cursor = %d, should stay at 0", m.cursor)
		}
	})

	t.Run("boundary at bottom", func(t *testing.T) {
		m.cursor = 2
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if m.cursor != 2 {
			t.Errorf("cursor = %d, should stay at 2", m.cursor)
		}
	})
}

func TestContentTypeModel_Update_QuickSelect(t *testing.T) {
	m := NewContentTypeModel()

	t.Run("key 1 selects Live", func(t *testing.T) {
		m.cursor = 2
		cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
		if m.cursor != 0 {
			t.Errorf("cursor = %d, want 0", m.cursor)
		}
		if cmd == nil {
			t.Error("cmd should not be nil")
		}
	})

	t.Run("key 2 selects VOD", func(t *testing.T) {
		m.cursor = 0
		cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
		if m.cursor != 1 {
			t.Errorf("cursor = %d, want 1", m.cursor)
		}
		if cmd == nil {
			t.Error("cmd should not be nil")
		}
	})

	t.Run("key 3 selects Series", func(t *testing.T) {
		m.cursor = 0
		cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
		if m.cursor != 2 {
			t.Errorf("cursor = %d, want 2", m.cursor)
		}
		if cmd == nil {
			t.Error("cmd should not be nil")
		}
	})
}

func TestContentTypeModel_Update_Enter(t *testing.T) {
	m := NewContentTypeModel()
	m.cursor = 1

	cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("enter should return a command")
	}

	// Execute command and check message type
	msg := cmd()
	if _, ok := msg.(tui.ContentTypeSelectedMsg); !ok {
		t.Errorf("msg type = %T, want ContentTypeSelectedMsg", msg)
	}
}

func TestContentTypeModel_View(t *testing.T) {
	m := NewContentTypeModel()
	m.SetSize(80, 40)

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// Test with user info
	m.SetUserInfo("test@example.com")
	viewWithInfo := m.View()
	if viewWithInfo == "" {
		t.Error("View() with user info returned empty string")
	}
}
