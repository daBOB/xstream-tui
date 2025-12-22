package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/xc"
)

func TestEpisodeItem(t *testing.T) {
	t.Run("with duration", func(t *testing.T) {
		item := EpisodeItem{xc.Episode{
			ID:    xc.NewFlexibleID(1),
			Title: "Pilot",
			Info: xc.EpisodeInfo{
				Duration: "45:00",
			},
		}}

		if item.FilterValue() != "Pilot" {
			t.Errorf("FilterValue() = %q, want 'Pilot'", item.FilterValue())
		}
		if item.Title() != "Pilot" {
			t.Errorf("Title() = %q, want 'Pilot'", item.Title())
		}
		if item.Description() != "⏱ 45:00" {
			t.Errorf("Description() = %q, want '⏱ 45:00'", item.Description())
		}
	})

	t.Run("without duration", func(t *testing.T) {
		item := EpisodeItem{xc.Episode{
			ID:    xc.NewFlexibleID(2),
			Title: "Episode 2",
		}}

		if item.Description() != "" {
			t.Errorf("Description() = %q, want ''", item.Description())
		}
	})
}

func TestNewEpisodesModel(t *testing.T) {
	m := NewEpisodesModel()
	if m == nil {
		t.Fatal("NewEpisodesModel() returned nil")
	}
	if m.list == nil {
		t.Error("list should be initialized")
	}
}

func TestEpisodesModel_SetClient(t *testing.T) {
	m := NewEpisodesModel()
	m.SetClient(nil)
	if m.client != nil {
		t.Error("client should be nil")
	}
}

func TestEpisodesModel_SetSeriesName(t *testing.T) {
	m := NewEpisodesModel()
	m.SetSeriesName("Breaking Bad")
	if m.seriesName != "Breaking Bad" {
		t.Errorf("seriesName = %q, want 'Breaking Bad'", m.seriesName)
	}
}

func TestEpisodesModel_SetSeason(t *testing.T) {
	m := NewEpisodesModel()
	season := xc.SeasonInfo{Name: "Season 1"}
	episodes := []xc.Episode{
		{ID: xc.NewFlexibleID(1), Title: "Episode 1"},
		{ID: xc.NewFlexibleID(2), Title: "Episode 2"},
	}

	m.SetSeason(season, episodes)

	if m.season.Name != "Season 1" {
		t.Errorf("season.Name = %q, want 'Season 1'", m.season.Name)
	}
	if len(m.episodes) != 2 {
		t.Errorf("episodes = %d, want 2", len(m.episodes))
	}
}

func TestEpisodesModel_SetSize(t *testing.T) {
	m := NewEpisodesModel()
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestEpisodesModel_Update_Search(t *testing.T) {
	m := NewEpisodesModel()
	m.SetSize(80, 40)

	// Start search
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.searching {
		t.Error("should be in search mode after '/'")
	}

	// Enter confirms search
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.searching {
		t.Error("should exit search mode on Enter")
	}

	// Start again and escape
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.searching {
		t.Error("should exit search mode on Esc")
	}
}

func TestEpisodesModel_SelectEpisode_Empty(t *testing.T) {
	m := NewEpisodesModel()
	m.SetSize(80, 40)

	cmd := m.selectEpisode()
	if cmd != nil {
		t.Error("selectEpisode should return nil when list empty")
	}
}

func TestEpisodesModel_SelectEpisode_WithItem(t *testing.T) {
	m := NewEpisodesModel()
	m.SetSize(80, 40)
	m.SetSeason(xc.SeasonInfo{Name: "S1"}, []xc.Episode{
		{ID: xc.NewFlexibleID(1), Title: "Pilot"},
	})

	cmd := m.selectEpisode()
	if cmd == nil {
		t.Error("selectEpisode should return command")
	}
}

func TestEpisodesModel_DownloadEpisode_Empty(t *testing.T) {
	m := NewEpisodesModel()
	m.SetSize(80, 40)

	cmd := m.downloadEpisode()
	if cmd != nil {
		t.Error("downloadEpisode should return nil when list empty")
	}
}

func TestEpisodesModel_View(t *testing.T) {
	m := NewEpisodesModel()
	m.SetSize(80, 40)
	m.season = xc.SeasonInfo{Name: "Season 1"}

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestEpisodesModel_View_Searching(t *testing.T) {
	m := NewEpisodesModel()
	m.SetSize(80, 40)
	m.searching = true

	view := m.View()
	if view == "" {
		t.Error("View() in search mode returned empty string")
	}
}

func TestEpisodesModel_PopulateList(t *testing.T) {
	m := NewEpisodesModel()
	m.episodes = []xc.Episode{
		{ID: xc.NewFlexibleID(1), Title: "Ep1"},
		{ID: xc.NewFlexibleID(2), Title: "Ep2"},
	}

	m.populateList()
	// No panic = success
}
