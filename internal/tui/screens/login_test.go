package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewLoginModel(t *testing.T) {
	m := NewLoginModel()
	if m == nil {
		t.Fatal("NewLoginModel() returned nil")
	}
	if len(m.inputs) != 4 {
		t.Errorf("inputs count = %d, want 4", len(m.inputs))
	}
	if m.focused != 0 {
		t.Errorf("focused = %d, want 0", m.focused)
	}
}

func TestLoginModel_SetSize(t *testing.T) {
	m := NewLoginModel()
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestLoginModel_SetError(t *testing.T) {
	m := NewLoginModel()
	m.SetError("test error")

	if m.err != "test error" {
		t.Errorf("err = %q, want 'test error'", m.err)
	}
}

func TestLoginModel_ClearError(t *testing.T) {
	m := NewLoginModel()
	m.SetError("test error")
	m.ClearError()

	if m.err != "" {
		t.Errorf("err = %q, want ''", m.err)
	}
}

func TestLoginModel_Focus(t *testing.T) {
	m := NewLoginModel()
	m.focused = 2

	cmd := m.Focus()
	if m.focused != 0 {
		t.Errorf("focused = %d after Focus(), want 0", m.focused)
	}
	_ = cmd // cmd focuses the input
}

func TestLoginModel_Update_Navigation(t *testing.T) {
	m := NewLoginModel()

	t.Run("tab moves to next", func(t *testing.T) {
		m.focused = 0
		m.Update(tea.KeyMsg{Type: tea.KeyTab})
		if m.focused != 1 {
			t.Errorf("focused = %d after tab, want 1", m.focused)
		}
	})

	t.Run("down moves to next", func(t *testing.T) {
		m.focused = 1
		m.Update(tea.KeyMsg{Type: tea.KeyDown})
		if m.focused != 2 {
			t.Errorf("focused = %d after down, want 2", m.focused)
		}
	})

	t.Run("shift+tab moves to prev", func(t *testing.T) {
		m.focused = 2
		m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if m.focused != 1 {
			t.Errorf("focused = %d after shift+tab, want 1", m.focused)
		}
	})

	t.Run("up moves to prev", func(t *testing.T) {
		m.focused = 1
		m.Update(tea.KeyMsg{Type: tea.KeyUp})
		if m.focused != 0 {
			t.Errorf("focused = %d after up, want 0", m.focused)
		}
	})

	t.Run("wraps around at end", func(t *testing.T) {
		m.focused = 3
		m.Update(tea.KeyMsg{Type: tea.KeyTab})
		if m.focused != 0 {
			t.Errorf("focused = %d, should wrap to 0", m.focused)
		}
	})

	t.Run("wraps around at start", func(t *testing.T) {
		m.focused = 0
		m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		if m.focused != 3 {
			t.Errorf("focused = %d, should wrap to 3", m.focused)
		}
	})
}

func TestLoginModel_Update_ClearsError(t *testing.T) {
	m := NewLoginModel()
	m.err = "some error"

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if m.err != "" {
		t.Errorf("err = %q, should be cleared on keypress", m.err)
	}
}

func TestLoginModel_Update_Enter_NotOnPassword(t *testing.T) {
	m := NewLoginModel()
	m.focused = 0 // host field

	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.focused != 1 {
		t.Errorf("focused = %d, enter should move to next field", m.focused)
	}
}

func TestLoginModel_Submit_Validation(t *testing.T) {
	m := NewLoginModel()
	m.focused = inputPass

	t.Run("empty host", func(t *testing.T) {
		m.inputs[inputHost].SetValue("")
		m.inputs[inputUser].SetValue("user")
		m.inputs[inputPass].SetValue("pass")

		cmd := m.submit()
		if cmd != nil {
			t.Error("should return nil for validation error")
		}
		if m.err == "" {
			t.Error("err should be set for missing host")
		}
	})

	t.Run("empty username", func(t *testing.T) {
		m.inputs[inputHost].SetValue("example.com")
		m.inputs[inputUser].SetValue("")
		m.inputs[inputPass].SetValue("pass")

		cmd := m.submit()
		if cmd != nil {
			t.Error("should return nil for validation error")
		}
		if m.err == "" {
			t.Error("err should be set for missing username")
		}
	})

	t.Run("empty password", func(t *testing.T) {
		m.inputs[inputHost].SetValue("example.com")
		m.inputs[inputUser].SetValue("user")
		m.inputs[inputPass].SetValue("")

		cmd := m.submit()
		if cmd != nil {
			t.Error("should return nil for validation error")
		}
		if m.err == "" {
			t.Error("err should be set for missing password")
		}
	})
}

func TestLoginModel_Submit_ValidInput(t *testing.T) {
	m := NewLoginModel()
	m.inputs[inputHost].SetValue("example.com")
	m.inputs[inputPort].SetValue("8080")
	m.inputs[inputUser].SetValue("user")
	m.inputs[inputPass].SetValue("pass")

	cmd := m.submit()
	if cmd == nil {
		t.Error("submit() should return command for valid input")
	}

	// Password should be cleared after submit
	if m.inputs[inputPass].Value() != "" {
		t.Error("password should be cleared after submit")
	}
}

func TestLoginModel_Submit_DefaultPort(t *testing.T) {
	m := NewLoginModel()
	m.inputs[inputHost].SetValue("example.com")
	m.inputs[inputPort].SetValue("") // empty port
	m.inputs[inputUser].SetValue("user")
	m.inputs[inputPass].SetValue("pass")

	cmd := m.submit()
	if cmd == nil {
		t.Error("submit() should return command even with empty port")
	}
}

func TestLoginModel_View(t *testing.T) {
	m := NewLoginModel()
	m.SetSize(80, 40)

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestLoginModel_View_WithError(t *testing.T) {
	m := NewLoginModel()
	m.SetSize(80, 40)
	m.SetError("Authentication failed")

	view := m.View()
	if view == "" {
		t.Error("View() with error returned empty string")
	}
}

func TestLoginModel_FocusNext(t *testing.T) {
	m := NewLoginModel()
	m.focused = 0

	m.focusNext()
	if m.focused != 1 {
		t.Errorf("focused = %d after focusNext, want 1", m.focused)
	}
}

func TestLoginModel_FocusPrev(t *testing.T) {
	m := NewLoginModel()
	m.focused = 1

	m.focusPrev()
	if m.focused != 0 {
		t.Errorf("focused = %d after focusPrev, want 0", m.focused)
	}
}
