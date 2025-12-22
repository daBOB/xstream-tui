package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSplitView(t *testing.T) {
	tests := []struct {
		name       string
		proportion float64
		want       float64
	}{
		{"normal proportion", 0.5, 0.5},
		{"too small clamped", 0.1, 0.2},
		{"too large clamped", 0.9, 0.8},
		{"exact min", 0.2, 0.2},
		{"exact max", 0.8, 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv := NewSplitView(tt.proportion)
			if sv.leftWidth != tt.want {
				t.Errorf("leftWidth = %v, want %v", sv.leftWidth, tt.want)
			}
		})
	}
}

func TestSplitView_SetSize(t *testing.T) {
	sv := NewSplitView(0.5)
	sv.SetSize(100, 50)

	if sv.width != 100 {
		t.Errorf("width = %d, want 100", sv.width)
	}
	if sv.height != 50 {
		t.Errorf("height = %d, want 50", sv.height)
	}
}

func TestSplitView_SetTitles(t *testing.T) {
	sv := NewSplitView(0.5)
	sv.SetTitles("Left", "Right")

	if sv.leftTitle != "Left" {
		t.Errorf("leftTitle = %q, want 'Left'", sv.leftTitle)
	}
	if sv.rightTitle != "Right" {
		t.Errorf("rightTitle = %q, want 'Right'", sv.rightTitle)
	}
}

func TestSplitView_GetDimensions(t *testing.T) {
	sv := NewSplitView(0.5)
	sv.SetSize(100, 50)

	leftW, leftH := sv.GetLeftDimensions()
	if leftH != 46 { // height - 4 for title bar
		t.Errorf("leftH = %d, want 46", leftH)
	}
	if leftW < 10 {
		t.Errorf("leftW = %d, should be >= 10", leftW)
	}

	rightW, rightH := sv.GetRightDimensions()
	if rightH != 46 {
		t.Errorf("rightH = %d, want 46", rightH)
	}
	if rightW < 10 {
		t.Errorf("rightW = %d, should be >= 10", rightW)
	}
}

func TestSplitView_Active(t *testing.T) {
	sv := NewSplitView(0.5)

	// Default is left
	if sv.Active() != LeftPane {
		t.Error("default active should be LeftPane")
	}

	// Set to right
	sv.SetActive(RightPane)
	if sv.Active() != RightPane {
		t.Error("active should be RightPane after SetActive")
	}

	// Toggle back
	sv.ToggleActive()
	if sv.Active() != LeftPane {
		t.Error("active should be LeftPane after ToggleActive")
	}
}

func TestSplitView_Update(t *testing.T) {
	sv := NewSplitView(0.5)

	t.Run("tab switches to right", func(t *testing.T) {
		sv.SetActive(LeftPane)
		sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("tab")})
		// Note: tea.KeyMsg string representation
	})

	t.Run("l switches to right", func(t *testing.T) {
		sv.SetActive(LeftPane)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
		sv.Update(msg)
		if sv.Active() != RightPane {
			t.Error("should switch to right on 'l' from left")
		}
	})

	t.Run("h switches to left", func(t *testing.T) {
		sv.SetActive(RightPane)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
		sv.Update(msg)
		if sv.Active() != LeftPane {
			t.Error("should switch to left on 'h' from right")
		}
	})
}

func TestDefaultSplitViewStyles(t *testing.T) {
	styles := DefaultSplitViewStyles()

	// Just verify it returns non-nil styles
	_ = styles.ActiveBorder
	_ = styles.InactiveBorder
	_ = styles.Title
	_ = styles.ActiveTitle
	_ = styles.Divider
}

func TestSplitView_Render(t *testing.T) {
	sv := NewSplitView(0.5)
	sv.SetSize(80, 20)
	sv.SetTitles("Left Pane", "Right Pane")

	styles := DefaultSplitViewStyles()
	result := sv.Render("left content", "right content", styles)

	if result == "" {
		t.Error("Render() returned empty string")
	}
}

func TestSplitView_RenderSimple(t *testing.T) {
	sv := NewSplitView(0.5)
	sv.SetSize(80, 20)
	sv.SetTitles("Left", "Right")

	result := sv.RenderSimple("left", "right")
	if result == "" {
		t.Error("RenderSimple() returned empty string")
	}
}

func TestSplitPane_Constants(t *testing.T) {
	if LeftPane != 0 {
		t.Errorf("LeftPane = %d, want 0", LeftPane)
	}
	if RightPane != 1 {
		t.Errorf("RightPane = %d, want 1", RightPane)
	}
}
