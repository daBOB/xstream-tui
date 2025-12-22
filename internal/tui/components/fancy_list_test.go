package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// testFancyItem implements FancyListItem
type testFancyItem struct {
	title       string
	description string
	filterValue string
}

func (t testFancyItem) Title() string       { return t.title }
func (t testFancyItem) Description() string { return t.description }
func (t testFancyItem) FilterValue() string { return t.filterValue }

func TestDefaultFancyListStyles(t *testing.T) {
	styles := DefaultFancyListStyles()

	// Verify styles are initialized
	_ = styles.Title.String()
	_ = styles.TitleSelected.String()
	_ = styles.Desc.String()
	_ = styles.DescSelected.String()
	_ = styles.FilterMatch.String()
}

func TestNewFancyList(t *testing.T) {
	l := NewFancyList("Test Title", true)
	if l == nil {
		t.Fatal("NewFancyList() returned nil")
	}
	if l.Len() != 0 {
		t.Errorf("Len() = %d, want 0", l.Len())
	}
}

func TestFancyList_SetItems(t *testing.T) {
	l := NewFancyList("Test", false)
	items := []FancyListItem{
		testFancyItem{title: "Item 1"},
		testFancyItem{title: "Item 2"},
	}
	l.SetItems(items)

	if l.Len() != 2 {
		t.Errorf("Len() = %d, want 2", l.Len())
	}
}

func TestFancyList_SetSize(t *testing.T) {
	l := NewFancyList("Test", false)
	l.SetSize(100, 50)
	// SetSize doesn't have public getters, just verify no panic
}

func TestFancyList_Selected(t *testing.T) {
	l := NewFancyList("Test", false)

	// Empty list
	if l.Selected() != nil {
		t.Error("Selected() should be nil on empty list")
	}

	// With items
	l.SetItems([]FancyListItem{
		testFancyItem{title: "First"},
	})

	selected := l.Selected()
	if selected == nil {
		t.Fatal("Selected() returned nil with items")
	}
	if selected.Title() != "First" {
		t.Errorf("Selected().Title() = %q, want 'First'", selected.Title())
	}
}

func TestFancyList_Update(t *testing.T) {
	l := NewFancyList("Test", false)
	l.SetItems([]FancyListItem{
		testFancyItem{title: "A"},
		testFancyItem{title: "B"},
	})
	l.SetSize(80, 20)

	// Update with key message
	result, cmd := l.Update(tea.KeyMsg{Type: tea.KeyDown})
	if result == nil {
		t.Error("Update() returned nil")
	}
	_ = cmd // cmd may be nil or not
}

func TestFancyList_View(t *testing.T) {
	l := NewFancyList("Test", true)
	l.SetSize(80, 20)
	l.SetItems([]FancyListItem{
		testFancyItem{title: "Item", description: "Desc"},
	})

	view := l.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestFancyList_Filtering(t *testing.T) {
	l := NewFancyList("Test", false)
	// Initially not filtering
	if l.Filtering() {
		t.Error("Filtering() should be false initially")
	}
}

func TestFancyList_SetTitle(t *testing.T) {
	l := NewFancyList("Initial", false)
	l.SetTitle("New Title")
	// No public getter, but verify no panic
}

func TestFancyList_Index(t *testing.T) {
	l := NewFancyList("Test", false)
	l.SetItems([]FancyListItem{
		testFancyItem{title: "A"},
		testFancyItem{title: "B"},
	})

	if l.Index() < 0 {
		t.Errorf("Index() = %d, should be >= 0", l.Index())
	}
}

func TestFancyItemDelegate_Height(t *testing.T) {
	t.Run("without description", func(t *testing.T) {
		d := fancyItemDelegate{showDesc: false}
		if d.Height() != 1 {
			t.Errorf("Height() = %d, want 1", d.Height())
		}
	})

	t.Run("with description", func(t *testing.T) {
		d := fancyItemDelegate{showDesc: true}
		if d.Height() != 2 {
			t.Errorf("Height() = %d, want 2", d.Height())
		}
	})
}

func TestFancyItemDelegate_Spacing(t *testing.T) {
	d := fancyItemDelegate{}
	if d.Spacing() != 0 {
		t.Errorf("Spacing() = %d, want 0", d.Spacing())
	}
}

func TestFancyListItemWrapper_FilterValue(t *testing.T) {
	wrapper := fancyListItemWrapper{
		item: testFancyItem{filterValue: "test-filter"},
	}
	if wrapper.FilterValue() != "test-filter" {
		t.Errorf("FilterValue() = %q, want 'test-filter'", wrapper.FilterValue())
	}
}
