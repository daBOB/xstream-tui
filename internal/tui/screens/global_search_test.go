package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/xc"
)

func TestGlobalSearchItem(t *testing.T) {
	item := GlobalSearchItem{
		name:        "Test Movie",
		description: "Action",
		itemType:    tui.VODContent,
	}

	if item.FilterValue() != "Test Movie" {
		t.Errorf("FilterValue() = %q, want 'Test Movie'", item.FilterValue())
	}
	if item.Title() != "Test Movie" {
		t.Errorf("Title() = %q, want 'Test Movie'", item.Title())
	}
	if item.Description() != "Action" {
		t.Errorf("Description() = %q, want 'Action'", item.Description())
	}
}

func TestNewGlobalSearchModel(t *testing.T) {
	m := NewGlobalSearchModel()
	if m == nil {
		t.Fatal("NewGlobalSearchModel() returned nil")
	}
	if m.list == nil {
		t.Error("list should be initialized")
	}
}

func TestGlobalSearchModel_SetClient(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetClient(nil)
	if m.client != nil {
		t.Error("client should be nil")
	}
}

func TestGlobalSearchModel_SetSize(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestGlobalSearchModel_SetResults_Live(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)

	msg := tui.GlobalSearchResultsMsg{
		LiveStreams: []xc.LiveStream{
			{ID: xc.NewFlexibleID(1), Name: "CNN"},
		},
		Categories: []xc.Category{
			{ID: xc.NewFlexibleID(1), Name: "News"},
		},
	}

	m.SetResults(msg)
	if len(m.allItems) != 1 {
		t.Errorf("allItems = %d, want 1", len(m.allItems))
	}
	if m.loading {
		t.Error("loading should be false after SetResults")
	}
	if !m.loaded {
		t.Error("loaded should be true after SetResults")
	}
}

func TestGlobalSearchModel_SetResults_VOD(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)

	msg := tui.GlobalSearchResultsMsg{
		VODStreams: []xc.VODStream{
			{ID: xc.NewFlexibleID(1), Name: "Movie"},
		},
	}

	m.SetResults(msg)
	if len(m.allItems) != 1 {
		t.Errorf("allItems = %d, want 1", len(m.allItems))
	}
}

func TestGlobalSearchModel_SetResults_Series(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)

	msg := tui.GlobalSearchResultsMsg{
		Series: []xc.Series{
			{ID: xc.NewFlexibleID(1), Name: "Show"},
		},
	}

	m.SetResults(msg)
	if len(m.allItems) != 1 {
		t.Errorf("allItems = %d, want 1", len(m.allItems))
	}
}

func TestGlobalSearchModel_Update_GlobalSearchResults(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)

	msg := tui.GlobalSearchResultsMsg{
		LiveStreams: []xc.LiveStream{{ID: xc.NewFlexibleID(1), Name: "Test"}},
	}

	m.Update(msg)
	if len(m.allItems) != 1 {
		t.Errorf("allItems = %d after update, want 1", len(m.allItems))
	}
}

func TestGlobalSearchModel_Update_Navigation(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)

	// Navigation keys
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
}

func TestGlobalSearchModel_Update_Escape(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)
	m.search.SetValue("test")

	// Escape clears search first
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.search.Value() != "" {
		t.Error("Esc should clear search text")
	}
}

func TestGlobalSearchModel_FilterItems(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)
	m.allItems = []GlobalSearchItem{
		{name: "Apple"},
		{name: "Banana"},
	}

	m.search.SetValue("ban")
	m.filterItems()
	// Filter should reduce visible items
}

func TestGlobalSearchModel_SelectItem_Empty(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)

	cmd := m.selectItem()
	if cmd != nil {
		t.Error("selectItem should return nil when list empty")
	}
}

func TestGlobalSearchModel_View(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)
	m.contentType = tui.LiveContent

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestGlobalSearchModel_View_Loading(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)
	m.loading = true

	view := m.View()
	if view == "" {
		t.Error("View() in loading state returned empty string")
	}
}

func TestGlobalSearchModel_View_Loaded(t *testing.T) {
	m := NewGlobalSearchModel()
	m.SetSize(80, 40)
	m.loaded = true

	view := m.View()
	if view == "" {
		t.Error("View() in loaded state returned empty string")
	}
}

func TestGlobalSearchModel_LoadAllContent_NoClient(t *testing.T) {
	m := NewGlobalSearchModel()
	m.contentType = tui.LiveContent

	cmd := m.loadAllContent()
	if cmd == nil {
		t.Fatal("loadAllContent() returned nil")
	}

	msg := cmd()
	if _, ok := msg.(tui.ErrorMsg); !ok {
		t.Errorf("msg type = %T, want ErrorMsg", msg)
	}
}
