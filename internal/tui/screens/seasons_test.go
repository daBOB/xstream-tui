package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/xc"
)

func TestSeasonItem(t *testing.T) {
	item := SeasonItem{
		SeasonInfo: xc.SeasonInfo{
			SeasonNumber: xc.NewFlexibleID(1),
			Name:         "Season 1",
		},
		ActualEpisodeCount: 10,
	}

	if item.FilterValue() != "Season 1" {
		t.Errorf("FilterValue() = %q, want 'Season 1'", item.FilterValue())
	}
	if item.Title() != "Season 1" {
		t.Errorf("Title() = %q, want 'Season 1'", item.Title())
	}
	if item.Description() != "10 episodes" {
		t.Errorf("Description() = %q, want '10 episodes'", item.Description())
	}
}

func TestNewSeasonsModel(t *testing.T) {
	m := NewSeasonsModel()
	if m == nil {
		t.Fatal("NewSeasonsModel() returned nil")
	}
	if m.list == nil {
		t.Error("list should be initialized")
	}
}

func TestSeasonsModel_SetClient(t *testing.T) {
	m := NewSeasonsModel()
	m.SetClient(nil)
	if m.client != nil {
		t.Error("client should be nil")
	}
}

func TestSeasonsModel_SetSize(t *testing.T) {
	m := NewSeasonsModel()
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestSeasonsModel_SetSeriesInfo(t *testing.T) {
	m := NewSeasonsModel()
	info := &xc.SeriesInfo{
		Seasons: []xc.SeasonInfo{
			{SeasonNumber: xc.NewFlexibleID(1), Name: "Season 1"},
		},
		Episodes: map[string][]xc.Episode{
			"1": {{ID: xc.NewFlexibleID(1), Title: "Ep 1"}},
		},
	}

	m.SetSeriesInfo(info)
	if m.seriesInfo == nil {
		t.Error("seriesInfo should be set")
	}
}

func TestSeasonsModel_Update_SeriesInfoLoaded(t *testing.T) {
	m := NewSeasonsModel()
	info := &xc.SeriesInfo{
		Seasons: []xc.SeasonInfo{
			{SeasonNumber: xc.NewFlexibleID(1), Name: "Season 1"},
		},
		Episodes: map[string][]xc.Episode{
			"1": {},
		},
	}

	m.Update(tui.SeriesInfoLoadedMsg{Info: info})
	if m.seriesInfo == nil {
		t.Error("seriesInfo should be set after update")
	}
}

func TestSeasonsModel_Update_Search(t *testing.T) {
	m := NewSeasonsModel()
	m.SetSize(80, 40)

	// Start search
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.searching {
		t.Error("should be in search mode after '/'")
	}

	// Escape cancels search
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.searching {
		t.Error("should exit search mode on Esc")
	}
}

func TestSeasonsModel_Update_Navigation(t *testing.T) {
	m := NewSeasonsModel()
	m.SetSize(80, 40)

	// Navigate with key
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
}

func TestSeasonsModel_SelectSeason_NoItem(t *testing.T) {
	m := NewSeasonsModel()
	m.SetSize(80, 40)

	// Select with empty list
	cmd := m.selectSeason()
	if cmd != nil {
		t.Error("selectSeason should return nil when list empty")
	}
}

func TestSeasonsModel_View(t *testing.T) {
	m := NewSeasonsModel()
	m.SetSize(80, 40)
	m.series = xc.Series{Name: "Test Series"}

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestSeasonsModel_View_WithRating(t *testing.T) {
	m := NewSeasonsModel()
	m.SetSize(80, 40)
	m.series = xc.Series{Name: "Test Series", Rating: "8.5"}

	view := m.View()
	if view == "" {
		t.Error("View() with rating returned empty string")
	}
}

func TestSeasonsModel_View_Searching(t *testing.T) {
	m := NewSeasonsModel()
	m.SetSize(80, 40)
	m.searching = true

	view := m.View()
	if view == "" {
		t.Error("View() in search mode returned empty string")
	}
}

func TestSeasonsModel_LoadSeriesInfo_NoClient(t *testing.T) {
	m := NewSeasonsModel()

	cmd := m.loadSeriesInfo()
	if cmd == nil {
		t.Fatal("loadSeriesInfo() returned nil")
	}

	msg := cmd()
	if _, ok := msg.(tui.ErrorMsg); !ok {
		t.Errorf("msg type = %T, want ErrorMsg", msg)
	}
}
