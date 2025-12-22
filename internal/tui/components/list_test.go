package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// testItem implements ListItem for testing
type testItem struct {
	title       string
	description string
	filterValue string
}

func (t testItem) Title() string       { return t.title }
func (t testItem) Description() string { return t.description }
func (t testItem) FilterValue() string { return t.filterValue }

func TestNewVirtualList(t *testing.T) {
	l := NewVirtualList()
	if l == nil {
		t.Fatal("NewVirtualList() returned nil")
	}
	if l.Len() != 0 {
		t.Error("new list should be empty")
	}
}

func TestVirtualList_SetItems(t *testing.T) {
	l := NewVirtualList()
	items := []ListItem{
		testItem{title: "Item 1"},
		testItem{title: "Item 2"},
		testItem{title: "Item 3"},
	}
	l.SetItems(items)

	if l.Len() != 3 {
		t.Errorf("Len() = %d, want 3", l.Len())
	}
	if l.TotalLen() != 3 {
		t.Errorf("TotalLen() = %d, want 3", l.TotalLen())
	}
}

func TestVirtualList_SetSize(t *testing.T) {
	l := NewVirtualList()
	l.SetSize(100, 20)

	if l.width != 100 {
		t.Errorf("width = %d, want 100", l.width)
	}
	if l.height != 20 {
		t.Errorf("height = %d, want 20", l.height)
	}
}

func TestVirtualList_Selected(t *testing.T) {
	l := NewVirtualList()

	// Empty list
	if l.Selected() != nil {
		t.Error("Selected() should be nil on empty list")
	}

	// With items
	items := []ListItem{
		testItem{title: "First"},
		testItem{title: "Second"},
	}
	l.SetItems(items)

	selected := l.Selected()
	if selected == nil {
		t.Fatal("Selected() returned nil with items")
	}
	if selected.Title() != "First" {
		t.Errorf("Selected().Title() = %q, want 'First'", selected.Title())
	}
}

func TestVirtualList_SelectedIndex(t *testing.T) {
	l := NewVirtualList()
	l.SetItems([]ListItem{testItem{title: "A"}, testItem{title: "B"}})

	if l.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, want 0", l.SelectedIndex())
	}
}

func TestVirtualList_Navigation(t *testing.T) {
	l := NewVirtualList()
	l.SetSize(80, 10)
	l.SetItems([]ListItem{
		testItem{title: "A"},
		testItem{title: "B"},
		testItem{title: "C"},
	})

	t.Run("down key", func(t *testing.T) {
		l.cursor = 0
		l.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if l.cursor != 1 {
			t.Errorf("cursor = %d after down, want 1", l.cursor)
		}
	})

	t.Run("up key", func(t *testing.T) {
		l.cursor = 1
		l.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if l.cursor != 0 {
			t.Errorf("cursor = %d after up, want 0", l.cursor)
		}
	})

	t.Run("home key", func(t *testing.T) {
		l.cursor = 2
		l.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
		if l.cursor != 0 {
			t.Errorf("cursor = %d after home, want 0", l.cursor)
		}
	})

	t.Run("end key", func(t *testing.T) {
		l.cursor = 0
		l.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
		if l.cursor != 2 {
			t.Errorf("cursor = %d after end, want 2", l.cursor)
		}
	})
}

func TestVirtualList_Filter(t *testing.T) {
	l := NewVirtualList()
	l.SetItems([]ListItem{
		testItem{title: "Apple", filterValue: "Apple"},
		testItem{title: "Banana", filterValue: "Banana"},
		testItem{title: "Cherry", filterValue: "Cherry"},
	})

	// Filter for "an"
	l.Filter("an")
	if l.Len() != 1 {
		t.Errorf("Len() after filter = %d, want 1", l.Len())
	}
	if l.Selected().Title() != "Banana" {
		t.Errorf("Selected = %q, want 'Banana'", l.Selected().Title())
	}

	// Clear filter
	l.ClearFilter()
	if l.Len() != 3 {
		t.Errorf("Len() after clear = %d, want 3", l.Len())
	}
}

func TestVirtualList_View(t *testing.T) {
	l := NewVirtualList()
	l.SetSize(80, 10)

	t.Run("empty list", func(t *testing.T) {
		view := l.View()
		if view == "" {
			t.Error("View() empty for empty list")
		}
	})

	t.Run("with items", func(t *testing.T) {
		l.SetItems([]ListItem{
			testItem{title: "Test Item", description: "A description"},
		})
		view := l.View()
		if view == "" {
			t.Error("View() empty with items")
		}
	})
}

func TestVirtualList_ScrollInfo(t *testing.T) {
	l := NewVirtualList()
	l.SetSize(80, 10)

	t.Run("empty list", func(t *testing.T) {
		info := l.ScrollInfo()
		if info == "" {
			t.Error("ScrollInfo() empty for empty list")
		}
	})

	t.Run("with items", func(t *testing.T) {
		l.SetItems([]ListItem{testItem{title: "A"}, testItem{title: "B"}})
		info := l.ScrollInfo()
		if info == "" {
			t.Error("ScrollInfo() empty with items")
		}
	})
}

func TestVirtualList_MoveUpDownBoundaries(t *testing.T) {
	l := NewVirtualList()
	l.SetSize(80, 10)
	l.SetItems([]ListItem{testItem{title: "A"}, testItem{title: "B"}})

	// Move up at top should stay at 0
	l.cursor = 0
	l.moveUp(1)
	if l.cursor != 0 {
		t.Errorf("cursor = %d after moveUp at top, want 0", l.cursor)
	}

	// Move down at bottom should stay at last
	l.cursor = 1
	l.moveDown(1)
	if l.cursor != 1 {
		t.Errorf("cursor = %d after moveDown at bottom, want 1", l.cursor)
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{123, "123"},
		{999999, "999999"},
	}
	for _, tt := range tests {
		if got := itoa(tt.n); got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
