package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
)

// playSelectedEpisode plays the currently selected episode.
func (m *SeriesBrowserModel) playSelectedEpisode() tea.Cmd {
	item := m.episodes.Selected()
	if item == nil {
		return nil
	}

	epItem, ok := item.(episodeBrowserItem)
	if !ok {
		return nil
	}

	return func() tea.Msg {
		return tui.EpisodeSelectedMsg{Episode: epItem.Episode}
	}
}

// downloadSelectedEpisode downloads the currently selected episode.
func (m *SeriesBrowserModel) downloadSelectedEpisode() tea.Cmd {
	item := m.episodes.Selected()
	if item == nil || m.client == nil {
		return nil
	}

	epItem, ok := item.(episodeBrowserItem)
	if !ok {
		return nil
	}

	ep := epItem.Episode
	container := ep.ContainerExt
	if container == "" {
		container = "mp4"
	}
	url := m.client.SeriesEpisodeURL(ep.ID.String(), container)
	name := ep.Title + "." + container

	return func() tea.Msg {
		return tui.DownloadRequestMsg{
			Name:       name,
			URL:        url,
			SeriesName: m.series.Name,
			SeasonName: m.currentSeason,
		}
	}
}

// downloadSelectedSeason downloads all episodes in the selected season.
func (m *SeriesBrowserModel) downloadSelectedSeason() tea.Cmd {
	item := m.seasons.Selected()
	if item == nil || m.client == nil {
		return nil
	}

	seasonItem, ok := item.(seasonBrowserItem)
	if !ok {
		return nil
	}

	seasonNum := seasonItem.SeasonNumber.String()
	episodes := m.allEpisodes[seasonNum]
	if len(episodes) == 0 {
		return nil
	}

	var msgs []tui.DownloadRequestMsg
	for _, ep := range episodes {
		container := ep.ContainerExt
		if container == "" {
			container = "mp4"
		}
		url := m.client.SeriesEpisodeURL(ep.ID.String(), container)
		name := ep.Title + "." + container

		msgs = append(msgs, tui.DownloadRequestMsg{
			Name:       name,
			URL:        url,
			SeriesName: m.series.Name,
			SeasonName: seasonItem.Name,
		})
	}

	return func() tea.Msg {
		return tui.BatchDownloadMsg{Downloads: msgs}
	}
}
