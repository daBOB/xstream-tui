package screens

import (
	"testing"

	"github.com/altmueller/xstream-tui/internal/xc"
)

func TestSeriesBrowserModel_PopulateSeasons_Empty(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.seriesInfo = nil

	m.populateSeasons()
	// Should not panic
}

func TestSeriesBrowserModel_PopulateSeasons_WithSeasons(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.seriesInfo = &xc.SeriesInfo{
		Seasons: []xc.SeasonInfo{
			{SeasonNumber: xc.NewFlexibleID(1), Name: "Season 1"},
			{SeasonNumber: xc.NewFlexibleID(2), Name: "Season 2"},
		},
		Episodes: map[string][]xc.Episode{
			"1": {{ID: xc.NewFlexibleID(1), Title: "Ep 1"}},
			"2": {{ID: xc.NewFlexibleID(2), Title: "Ep 2"}},
		},
	}

	m.populateSeasons()
	if m.seasons.Len() != 2 {
		t.Errorf("seasons.Len() = %d, want 2", m.seasons.Len())
	}
}

func TestSeriesBrowserModel_GenerateSeasonsFromEpisodes(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.seriesInfo = &xc.SeriesInfo{
		Seasons:  []xc.SeasonInfo{}, // Empty seasons
		Episodes: map[string][]xc.Episode{
			"1": {{ID: xc.NewFlexibleID(1), Title: "S1E1"}},
			"2": {{ID: xc.NewFlexibleID(2), Title: "S2E1"}},
		},
	}

	m.generateSeasonsFromEpisodes()
	if m.seasons.Len() != 2 {
		t.Errorf("seasons.Len() = %d, want 2", m.seasons.Len())
	}
}

func TestSeriesBrowserModel_PopulateSeasons_GeneratesFromEpisodes(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.seriesInfo = &xc.SeriesInfo{
		Seasons: []xc.SeasonInfo{}, // No seasons, should generate
		Episodes: map[string][]xc.Episode{
			"1": {{ID: xc.NewFlexibleID(1), Title: "Ep 1"}},
		},
	}

	m.populateSeasons()
	if m.seasons.Len() != 1 {
		t.Errorf("seasons.Len() = %d, want 1", m.seasons.Len())
	}
}

func TestSeriesBrowserModel_SelectFirstSeason(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.seriesInfo = &xc.SeriesInfo{
		Seasons: []xc.SeasonInfo{
			{SeasonNumber: xc.NewFlexibleID(1), Name: "Season 1"},
		},
		Episodes: map[string][]xc.Episode{
			"1": {{ID: xc.NewFlexibleID(1), Title: "Ep 1"}},
		},
	}
	m.allEpisodes = m.seriesInfo.Episodes
	m.populateSeasons()

	m.selectFirstSeason()
	// Should populate episodes list
}

func TestSeriesBrowserModel_UpdateEpisodesForSelectedSeason_Empty(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)

	m.updateEpisodesForSelectedSeason()
	// Should not panic with empty seasons
}

func TestSeriesBrowserModel_LoadSeriesInfo_NoClient(t *testing.T) {
	m := NewSeriesBrowserModel()

	cmd := m.loadSeriesInfo()
	if cmd == nil {
		t.Fatal("loadSeriesInfo() returned nil")
	}

	msg := cmd()
	// Should return an ErrorMsg when client is nil
	t.Logf("msg type = %T", msg)
}
