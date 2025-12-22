package screens

import (
	"testing"

	"github.com/altmueller/xstream-tui/internal/xc"
)

func TestSeasonBrowserItem(t *testing.T) {
	item := seasonBrowserItem{
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

func TestEpisodeBrowserItem(t *testing.T) {
	t.Run("with duration", func(t *testing.T) {
		item := episodeBrowserItem{
			Episode: xc.Episode{
				ID:    xc.NewFlexibleID(123),
				Title: "Pilot",
				Info: xc.EpisodeInfo{
					Duration: "45:00",
				},
			},
		}

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
		item := episodeBrowserItem{
			Episode: xc.Episode{
				ID:    xc.NewFlexibleID(124),
				Title: "Episode 2",
				Info: xc.EpisodeInfo{
					Duration: "", // no duration
				},
			},
		}

		if item.Description() != "" {
			t.Errorf("Description() = %q, want ''", item.Description())
		}
	})
}
