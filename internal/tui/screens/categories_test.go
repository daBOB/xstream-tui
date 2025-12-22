package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/xc"
)

func TestCategoryItem(t *testing.T) {
	cat := xc.Category{
		ID:   xc.NewFlexibleID(123),
		Name: "Movies",
	}
	item := CategoryItem{cat}

	if item.FilterValue() != "Movies" {
		t.Errorf("FilterValue() = %q, want 'Movies'", item.FilterValue())
	}
	if item.Title() != "Movies" {
		t.Errorf("Title() = %q, want 'Movies'", item.Title())
	}
	if item.Description() != "" {
		t.Errorf("Description() = %q, want ''", item.Description())
	}
}

func TestNewCategoriesModel(t *testing.T) {
	m := NewCategoriesModel()
	if m == nil {
		t.Fatal("NewCategoriesModel() returned nil")
	}
	if m.list == nil {
		t.Error("list should be initialized")
	}
}

func TestCategoriesModel_SetClient(t *testing.T) {
	m := NewCategoriesModel()
	// We can't create a real client without server, just test nil handling
	m.SetClient(nil)
	if m.client != nil {
		t.Error("client should be nil")
	}
}

func TestCategoriesModel_SetSize(t *testing.T) {
	m := NewCategoriesModel()
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestCategoriesModel_SetCategories(t *testing.T) {
	m := NewCategoriesModel()
	cats := []xc.Category{
		{ID: xc.NewFlexibleID(1), Name: "Action"},
		{ID: xc.NewFlexibleID(2), Name: "Comedy"},
	}
	m.SetCategories(cats)

	if len(m.categories) != 2 {
		t.Errorf("categories count = %d, want 2", len(m.categories))
	}
}

func TestCategoriesModel_Update_CategoriesLoaded(t *testing.T) {
	m := NewCategoriesModel()
	cats := []xc.Category{{ID: xc.NewFlexibleID(1), Name: "Test"}}

	m.Update(tui.CategoriesLoadedMsg{Categories: cats})

	if len(m.categories) != 1 {
		t.Errorf("categories = %d after load, want 1", len(m.categories))
	}
}

func TestCategoriesModel_Update_Search(t *testing.T) {
	m := NewCategoriesModel()
	m.SetSize(80, 40)
	m.SetCategories([]xc.Category{
		{ID: xc.NewFlexibleID(1), Name: "Action"},
		{ID: xc.NewFlexibleID(2), Name: "Comedy"},
	})

	// Start search
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.searching {
		t.Error("should be in search mode after '/'")
	}

	// Type search query
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Escape cancels search
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.searching {
		t.Error("should exit search mode on Esc")
	}
}

func TestCategoriesModel_Update_Navigation(t *testing.T) {
	m := NewCategoriesModel()
	m.SetSize(80, 40)
	m.SetCategories([]xc.Category{
		{ID: xc.NewFlexibleID(1), Name: "A"},
		{ID: xc.NewFlexibleID(2), Name: "B"},
	})

	// Navigate down
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
}

func TestCategoriesModel_Update_Enter(t *testing.T) {
	m := NewCategoriesModel()
	m.SetSize(80, 40)
	m.SetCategories([]xc.Category{
		{ID: xc.NewFlexibleID(1), Name: "Test"},
	})

	cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("enter should return command with selection")
	}
}

func TestCategoriesModel_View(t *testing.T) {
	m := NewCategoriesModel()
	m.SetSize(80, 40)
	m.contentType = tui.LiveContent

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// Test different content types
	m.contentType = tui.VODContent
	vodView := m.View()
	if vodView == "" {
		t.Error("View() for VOD returned empty string")
	}

	m.contentType = tui.SeriesContent
	seriesView := m.View()
	if seriesView == "" {
		t.Error("View() for Series returned empty string")
	}
}

func TestCategoriesModel_View_Searching(t *testing.T) {
	m := NewCategoriesModel()
	m.SetSize(80, 40)
	m.searching = true

	view := m.View()
	if view == "" {
		t.Error("View() in search mode returned empty string")
	}
}

func TestCategoriesModel_LoadCategories_NoClient(t *testing.T) {
	m := NewCategoriesModel()
	m.contentType = tui.LiveContent

	cmd := m.loadCategories()
	if cmd == nil {
		t.Fatal("loadCategories() returned nil")
	}

	msg := cmd()
	if errMsg, ok := msg.(tui.ErrorMsg); !ok {
		t.Errorf("msg type = %T, want ErrorMsg", msg)
	} else if errMsg.Err == nil {
		t.Error("ErrorMsg.Err should not be nil")
	}
}

func TestCategoriesModel_SetContentType_NoClient(t *testing.T) {
	m := NewCategoriesModel()
	// No client set

	cmd := m.SetContentType(tui.VODContent)
	if cmd == nil {
		t.Error("SetContentType should return command")
	}
}

func TestCategoriesModel_Update_SearchEnter(t *testing.T) {
	m := NewCategoriesModel()
	m.SetSize(80, 40)
	m.searching = true

	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.searching {
		t.Error("Enter should exit search mode")
	}
}
