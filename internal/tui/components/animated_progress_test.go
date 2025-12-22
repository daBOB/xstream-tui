package components

import (
	"testing"

	"github.com/charmbracelet/bubbles/progress"
)

func TestNewAnimatedProgress(t *testing.T) {
	p := NewAnimatedProgress()
	if p == nil {
		t.Fatal("NewAnimatedProgress() returned nil")
	}
	if p.Percent() != 0 {
		t.Errorf("Percent() = %f, want 0", p.Percent())
	}
	if p.width != 40 {
		t.Errorf("width = %d, want 40", p.width)
	}
}

func TestNewAnimatedProgressWithGradient(t *testing.T) {
	p := NewAnimatedProgressWithGradient("#FF0000", "#00FF00")
	if p == nil {
		t.Fatal("NewAnimatedProgressWithGradient() returned nil")
	}
	if p.width != 40 {
		t.Errorf("width = %d, want 40", p.width)
	}
}

func TestAnimatedProgress_SetWidth(t *testing.T) {
	p := NewAnimatedProgress()
	p.SetWidth(100)
	if p.width != 100 {
		t.Errorf("width = %d, want 100", p.width)
	}
}

func TestAnimatedProgress_SetPercent(t *testing.T) {
	p := NewAnimatedProgress()

	t.Run("normal value", func(t *testing.T) {
		p.SetPercent(0.5)
		if p.Percent() != 0.5 {
			t.Errorf("Percent() = %f, want 0.5", p.Percent())
		}
	})

	t.Run("clamp below zero", func(t *testing.T) {
		p.SetPercent(-0.5)
		if p.Percent() != 0 {
			t.Errorf("Percent() = %f, want 0", p.Percent())
		}
	})

	t.Run("clamp above one", func(t *testing.T) {
		p.SetPercent(1.5)
		if p.Percent() != 1 {
			t.Errorf("Percent() = %f, want 1", p.Percent())
		}
	})
}

func TestAnimatedProgress_Update(t *testing.T) {
	p := NewAnimatedProgress()
	result, cmd := p.Update(progress.FrameMsg{})
	if result == nil {
		t.Error("Update() returned nil")
	}
	_ = cmd // cmd may be nil or have a value
}

func TestAnimatedProgress_View(t *testing.T) {
	p := NewAnimatedProgress()
	p.SetPercent(0.5)
	view := p.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestAnimatedProgress_ViewWithLabel(t *testing.T) {
	p := NewAnimatedProgress()
	p.SetWidth(40)
	view := p.ViewWithLabel("Progress", 0.5)
	if view == "" {
		t.Error("ViewWithLabel() returned empty string")
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		percent float64
		want    string
	}{
		{0.05, "  5%"},
		{0.50, " 50%"},
		{1.00, "100%"},
		{0.0, "  0%"},
	}
	for _, tt := range tests {
		if got := formatPercent(tt.percent); got != tt.want {
			t.Errorf("formatPercent(%f) = %q, want %q", tt.percent, got, tt.want)
		}
	}
}

func TestProgressTickMsg(t *testing.T) {
	msg := ProgressTickMsg{ID: "test-id"}
	if msg.ID != "test-id" {
		t.Errorf("ID = %q, want 'test-id'", msg.ID)
	}
}

func TestNewMultiProgress(t *testing.T) {
	mp := NewMultiProgress()
	if mp == nil {
		t.Fatal("NewMultiProgress() returned nil")
	}
	if mp.bars == nil {
		t.Error("bars should be initialized")
	}
}

func TestMultiProgress_Get(t *testing.T) {
	mp := NewMultiProgress()

	// Get creates new bar
	bar1 := mp.Get("id1")
	if bar1 == nil {
		t.Fatal("Get() returned nil")
	}

	// Get same ID returns same bar
	bar2 := mp.Get("id1")
	if bar1 != bar2 {
		t.Error("Get() should return same bar for same ID")
	}

	// Get different ID creates new bar
	bar3 := mp.Get("id2")
	if bar1 == bar3 {
		t.Error("Get() should create new bar for new ID")
	}
}

func TestMultiProgress_Remove(t *testing.T) {
	mp := NewMultiProgress()
	mp.Get("id1")

	if len(mp.bars) != 1 {
		t.Errorf("bars count = %d, want 1", len(mp.bars))
	}

	mp.Remove("id1")
	if len(mp.bars) != 0 {
		t.Errorf("bars count after remove = %d, want 0", len(mp.bars))
	}
}

func TestMultiProgress_Update(t *testing.T) {
	mp := NewMultiProgress()
	mp.Get("id1")
	mp.Get("id2")

	cmd := mp.Update(progress.FrameMsg{})
	_ = cmd // cmd may be nil or batch command
}
