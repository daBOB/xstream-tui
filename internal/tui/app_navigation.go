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
		if s, ok := a.contentType.(contentTypeScreen); ok {
			s.SetUserInfo(a.userInfo.Username)
		}
	case CategoriesScreen:
		if s, ok := a.categories.(categoriesScreen); ok {
			s.SetClient(a.client)
			cmd = s.SetContentType(a.currentType)
		}
	case StreamsScreen:
		if s, ok := a.streams.(streamsScreen); ok {
			s.SetClient(a.client)
			cmd = s.SetCategory(a.currentCat, a.currentType)
		}
	case SeasonsScreen:
		if s, ok := a.seasons.(seasonsScreen); ok {
			s.SetClient(a.client)
			cmd = s.SetSeries(a.currentSeries)
		}
	case EpisodesScreen:
		if s, ok := a.episodes.(episodesScreen); ok {
			s.SetClient(a.client)
			s.SetSeriesName(a.currentSeries.Name)
			s.SetSeason(a.currentSeason, a.currentSeasonEpisodes)
		}
	case SeriesBrowserScreen:
		if s, ok := a.seriesBrowser.(seriesBrowserScreen); ok {
			s.SetClient(a.client)
			cmd = s.SetSeries(a.currentSeries)
		}
	case GlobalSearchScreen:
		if s, ok := a.globalSearch.(globalSearchScreen); ok {
			s.SetClient(a.client)
			a.loading = true
			a.loadingMsg = "Loading all " + a.currentType.String() + "..."
			cmd = s.SetContentType(a.currentType)
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
	return func() tea.Msg {
		if a.playerStop != nil {
			a.playerStop()
		}

		ctx, cancel := context.WithCancel(context.Background())
		a.playerStop = cancel

		if err := a.player.Play(ctx, url, title); err != nil {
			return ErrorMsg{Err: err}
		}

		return PlayerStartedMsg{PlayerType: string(a.player.Type())}
	}
}

// StopPlayback stops current playback.
func (a *App) StopPlayback() {
	if a.playerStop != nil {
		a.playerStop()
	}
	a.player.Stop()
}
