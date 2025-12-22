package screens

import (
	"testing"
)

func TestSeriesBrowserModel_PlaySelectedEpisode_Empty(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)

	cmd := m.playSelectedEpisode()
	if cmd != nil {
		t.Error("playSelectedEpisode should return nil with no selection")
	}
}

func TestSeriesBrowserModel_DownloadSelectedEpisode_Empty(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)

	cmd := m.downloadSelectedEpisode()
	if cmd != nil {
		t.Error("downloadSelectedEpisode should return nil with no selection")
	}
}

func TestSeriesBrowserModel_DownloadSelectedSeason_Empty(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)

	cmd := m.downloadSelectedSeason()
	if cmd != nil {
		t.Error("downloadSelectedSeason should return nil with no selection")
	}
}

func TestSeriesBrowserModel_DownloadSelectedSeason_NoClient(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.client = nil

	cmd := m.downloadSelectedSeason()
	if cmd != nil {
		t.Error("downloadSelectedSeason should return nil when client is nil")
	}
}

func TestSeriesBrowserModel_DownloadSelectedEpisode_NoClient(t *testing.T) {
	m := NewSeriesBrowserModel()
	m.SetSize(80, 40)
	m.client = nil

	cmd := m.downloadSelectedEpisode()
	if cmd != nil {
		t.Error("downloadSelectedEpisode should return nil when client is nil")
	}
}
