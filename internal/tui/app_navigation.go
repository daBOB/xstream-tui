package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// navigateTo pushes current screen and navigates to new screen.
func (a *App) navigateTo(screen Screen) (*App, tea.Cmd) {
	a.navStack = append(a.navStack, a.screen)
	a.screen = screen

	var cmd tea.Cmd

	switch screen {
	case ContentTypeScreen:
		if a.contentType != nil {
			a.contentType.SetUserInfo(a.userInfo.Username)
		}
	case CategoriesScreen:
		if a.categories != nil {
			a.categories.SetClient(a.client)
			cmd = a.categories.SetContentType(a.currentType)
		}
	case StreamsScreen:
		if a.streams != nil {
			a.streams.SetClient(a.client)
			cmd = a.streams.SetCategory(a.currentCat, a.currentType)
		}
	case SeasonsScreen:
		if a.seasons != nil {
			a.seasons.SetClient(a.client)
			cmd = a.seasons.SetSeries(a.currentSeries)
		}
	case EpisodesScreen:
		if a.episodes != nil {
			a.episodes.SetClient(a.client)
			a.episodes.SetSeriesName(a.currentSeries.Name)
			a.episodes.SetSeason(a.currentSeason, a.currentSeasonEpisodes)
		}
	case SeriesBrowserScreen:
		if a.seriesBrowser != nil {
			a.seriesBrowser.SetClient(a.client)
			cmd = a.seriesBrowser.SetSeries(a.currentSeries)
		}
	case GlobalSearchScreen:
		if a.globalSearch != nil {
			a.globalSearch.SetClient(a.client)
			a.loading = true
			a.loadingMsg = "Loading all " + a.currentType.String() + "..."
			cmd = a.globalSearch.SetContentType(a.currentType)
		}
	}

	// Start spinner if loading
	if a.loading {
		return a, tea.Batch(cmd, a.spinnerTick())
	}
	return a, cmd
}

// navigateBack pops from navigation stack.
func (a *App) navigateBack() (*App, tea.Cmd) {
	if len(a.navStack) == 0 {
		return a, nil
	}
	a.screen = a.navStack[len(a.navStack)-1]
	a.navStack = a.navStack[:len(a.navStack)-1]
	a.loading = false
	return a, nil
}

// spinnerTick returns a command that triggers spinner animation.
func (a *App) spinnerTick() tea.Cmd {
	return tea.Tick(SpinnerTickRate, func(t time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}

// startPlayback launches the media player.
func (a *App) startPlayback(url, title string) tea.Cmd {
	// Cancel previous playback and create new context in Update (single-threaded)
	if a.playerStop != nil {
		a.playerStop()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.playerStop = cancel

	// Capture player reference locally to avoid accessing App fields from goroutine
	p := a.player

	return func() tea.Msg {
		if err := p.Play(ctx, url, title); err != nil {
			return ErrorMsg{Err: err}
		}

		return PlayerStartedMsg{PlayerType: string(p.Type())}
	}
}

// StopPlayback stops current playback.
func (a *App) StopPlayback() {
	if a.playerStop != nil {
		a.playerStop()
	}
	a.player.Stop()
}
