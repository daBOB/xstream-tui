package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/components"
	"github.com/altmueller/xstream-tui/internal/xc"
)

func TestNewSeriesBrowserModel(t *testing.T) {
	m := NewSeriesBrowserModel()
	if m == nil {
		t.Fatal("NewSeriesBrowserModel() returned nil")
	}
	if m.splitView == nil {
		t.Error("splitView should be initialized")
	}
	if m.seasons == nil {
		t.Error("seasons should be initialized")
	}
	if m.episodes == nil {
		t.Error("episodes should be initialized")
	}
}

func TestSeriesBrowserModel_SetClient(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetClient(nil)
	if m.client != nil {
		t.Error("client should be nil")
	}
}

func TestSeriesBrowserModel_SetSize(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestSeriesBrowserModel_SetSeries(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)

	series := xc.Series{ID: xc.NewFlexibleID(1), Name: "Test"}
	cmd := m.SetSeries(series)

	if !m.loading {
		t.Error("loading should be true after SetSeries")
	}
	if m.series.Name != "Test" {
		t.Errorf("series.Name = %q, want 'Test'", m.series.Name)
	}
	if cmd == nil {
		t.Error("SetSeries should return command")
	}
}

func TestSeriesBrowserModel_Update_SeriesInfoLoaded(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.loading = true

	info := &xc.SeriesInfo{
		Seasons: []xc.SeasonInfo{
			{SeasonNumber: xc.NewFlexibleID(1), Name: "Season 1"},
		},
		Episodes: map[string][]xc.Episode{
			"1": {{ID: xc.NewFlexibleID(1), Title: "Ep 1"}},
		},
	}

	m.Update(tui.SeriesInfoLoadedMsg{Info: info})
	if m.loading {
		t.Error("loading should be false after SeriesInfoLoaded")
	}
	if m.seriesInfo == nil {
		t.Error("seriesInfo should be set")
	}
}

func TestSeriesBrowserModel_Update_LoadingIgnoresKeys(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.loading = true

	// Keys should be ignored while loading
	cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("should ignore keys while loading")
	}
}

func TestSeriesBrowserModel_Update_TabSwitchPane(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.loading = false
	m.seriesInfo = &xc.SeriesInfo{}

	// Set up some episodes to allow switch
	m.episodes.SetItems([]components.FancyListItem{
		testFancyItem{title: "Episode 1"},
	})

	// Start at left pane, switch to right
	m.splitView.SetActive(components.LeftPane)
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.splitView.Active() != components.RightPane {
		t.Error("tab should switch to right pane")
	}
}

func TestSeriesBrowserModel_Update_ShiftTabSwitchPane(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.loading = false

	m.splitView.SetActive(components.RightPane)
	m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.splitView.Active() != components.LeftPane {
		t.Error("shift+tab should switch to left pane")
	}
}

func TestSeriesBrowserModel_Update_QueueToggle(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.loading = false

	cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Q'}})
	if cmd == nil {
		t.Error("Q should return command for queue toggle")
	}
}

func TestSeriesBrowserModel_View(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.series = xc.Series{Name: "Test Series"}

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestSeriesBrowserModel_View_Loading(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.series = xc.Series{Name: "Test"}
	m.loading = true

	view := m.View()
	if view == "" {
		t.Error("View() in loading state returned empty string")
	}
}

func TestSeriesBrowserModel_View_WithSeriesInfo(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.series = xc.Series{Name: "Test", Rating: "8.5"}
	m.seriesInfo = &xc.SeriesInfo{
		Seasons: []xc.SeasonInfo{
			{Name: "Season 1"},
		},
	}

	view := m.View()
	if view == "" {
		t.Error("View() with series info returned empty string")
	}
}

func TestSeriesBrowserModel_View_NoSeriesInfo(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.series = xc.Series{Name: "Test"}
	m.seriesInfo = nil
	m.loading = false

	view := m.View()
	if view == "" {
		t.Error("View() without series info returned empty string")
	}
}

// Helper type for FancyListItem
type testFancyItem struct {
	title       string
	description string
}

func (t testFancyItem) Title() string       { return t.title }
func (t testFancyItem) Description() string { return t.description }
func (t testFancyItem) FilterValue() string { return t.title }
