package screens

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/components"
	"github.com/altmueller/xstream-tui/internal/xc"
)

// loadSeriesInfo fetches series info from the API.
func (m *SeriesBrowserModel) loadSeriesInfo() tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return tui.ErrorMsg{Err: xc.ErrInvalidConfig}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		info, err := m.client.GetSeriesInfo(ctx, m.series.ID.String())
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		return tui.SeriesInfoLoadedMsg{Info: info}
	}
}

// populateSeasons fills the seasons list from series info.
func (m *SeriesBrowserModel) populateSeasons() {
	if m.seriesInfo == nil {
		return
	}

	// If seasons array is empty but we have episodes, generate from episodes
	if len(m.seriesInfo.Seasons) == 0 && len(m.seriesInfo.Episodes) > 0 {
		m.generateSeasonsFromEpisodes()
		return
	}

	items := make([]components.FancyListItem, len(m.seriesInfo.Seasons))
	for i, s := range m.seriesInfo.Seasons {
		seasonNum := s.SeasonNumber.String()
		actualCount := len(m.seriesInfo.Episodes[seasonNum])
		items[i] = seasonBrowserItem{
			SeasonInfo:         s,
			ActualEpisodeCount: actualCount,
		}
	}
	m.seasons.SetItems(items)
}

// generateSeasonsFromEpisodes creates season entries when seasons array is empty.
func (m *SeriesBrowserModel) generateSeasonsFromEpisodes() {
	seasonNums := make([]string, 0, len(m.seriesInfo.Episodes))
	for seasonNum := range m.seriesInfo.Episodes {
		seasonNums = append(seasonNums, seasonNum)
	}
	sort.Slice(seasonNums, func(i, j int) bool {
		ni, _ := strconv.Atoi(seasonNums[i])
		nj, _ := strconv.Atoi(seasonNums[j])
		return ni < nj
	})

	items := make([]components.FancyListItem, len(seasonNums))
	for i, seasonNum := range seasonNums {
		episodes := m.seriesInfo.Episodes[seasonNum]
		items[i] = seasonBrowserItem{
			SeasonInfo: xc.SeasonInfo{
				SeasonNumber: xc.NewFlexibleIDFromString(seasonNum),
				Name:         fmt.Sprintf("Season %s", seasonNum),
				EpisodeCount: xc.NewFlexibleID(len(episodes)),
			},
			ActualEpisodeCount: len(episodes),
		}
	}
	m.seasons.SetItems(items)
}

// selectFirstSeason selects the first season after loading.
func (m *SeriesBrowserModel) selectFirstSeason() {
	if m.seasons.Len() > 0 {
		m.updateEpisodesForSelectedSeason()
	}
}

// updateEpisodesForSelectedSeason populates episodes list for current season.
func (m *SeriesBrowserModel) updateEpisodesForSelectedSeason() {
	item := m.seasons.Selected()
	if item == nil {
		return
	}

	seasonItem, ok := item.(seasonBrowserItem)
	if !ok {
		return
	}

	m.currentSeason = seasonItem.Name

	seasonNum := seasonItem.SeasonNumber.String()
	episodes := m.allEpisodes[seasonNum]

	items := make([]components.FancyListItem, len(episodes))
	for i, ep := range episodes {
		items[i] = episodeBrowserItem{Episode: ep}
	}
	m.episodes.SetItems(items)
}
