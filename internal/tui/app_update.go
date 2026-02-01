package tui

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/download"
	"github.com/altmueller/xstream-tui/internal/tui/style"
)

// Update handles messages and updates application state.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle download daemon tick messages
	if a.downloadDaemon != nil {
		if statusMsg, cmd := a.downloadDaemon.Update(msg); statusMsg != nil {
			if status, ok := statusMsg.(download.DaemonStatusMsg); ok {
				a.downloadQueue.SetItems(status.Items)
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	// Handle progress bar animation frames
	if _, ok := msg.(progress.FrameMsg); ok {
		cmd := a.downloadQueue.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateScreenSizes()
		return a, nil

	case tea.KeyMsg:
		return a.handleKeyMsg(msg)

	case AuthSuccessMsg:
		a.client = msg.Client
		msg.UserInfo.Password = ""
		a.userInfo = msg.UserInfo
		a.loading = false
		return a.navigateTo(ContentTypeScreen)

	case AuthErrorMsg:
		a.loading = false
		if s, ok := a.login.(loginScreen); ok {
			s.SetError(msg.Error())
		}
		return a, nil

	case ContentTypeSelectedMsg:
		a.currentType = msg.Type
		a.loading = true
		a.loadingMsg = "Loading categories..."
		return a.navigateTo(CategoriesScreen)

	case CategoriesLoadedMsg:
		a.loading = false
		if s, ok := a.categories.(categoriesScreen); ok {
			return a, s.Update(msg)
		}
		return a, nil

	case CategorySelectedMsg:
		a.currentCat = msg.Category
		a.loading = true
		a.loadingMsg = "Loading streams..."
		return a.navigateTo(StreamsScreen)

	case StreamsLoadedMsg:
		a.loading = false
		if s, ok := a.streams.(streamsScreen); ok {
			return a, s.Update(msg)
		}
		return a, nil

	case StreamSelectedMsg:
		return a, a.startPlayback(msg.URL, msg.Name)

	case SeriesSelectedMsg:
		a.currentSeries = msg.Series
		a.loading = true
		a.loadingMsg = "Loading series info..."
		if a.useSplitView && a.seriesBrowser != nil {
			return a.navigateTo(SeriesBrowserScreen)
		}
		return a.navigateTo(SeasonsScreen)

	case SeriesInfoLoadedMsg:
		a.loading = false
		if a.screen == SeriesBrowserScreen {
			if s, ok := a.seriesBrowser.(seriesBrowserScreen); ok {
				return a, s.Update(msg)
			}
		} else if s, ok := a.seasons.(seasonsScreen); ok {
			return a, s.Update(msg)
		}
		return a, tea.Batch(cmds...)

	case SeasonSelectedMsg:
		a.currentSeason = msg.Season
		a.currentSeasonEpisodes = msg.Episodes
		return a.navigateTo(EpisodesScreen)

	case EpisodeSelectedMsg:
		container := msg.Episode.ContainerExt
		if container == "" {
			container = "mp4"
		}
		url := a.client.SeriesEpisodeURL(msg.Episode.ID.String(), container)
		return a, a.startPlayback(url, msg.Episode.Title)

	case PlayerStartedMsg:
		a.loading = false
		a.errorMsg = "Playing via " + msg.PlayerType + " - Press 'q' to stop"
		return a, nil

	case PlayerStoppedMsg:
		if msg.Err != nil {
			a.errorMsg = "Playback ended: " + msg.Err.Error()
		} else {
			a.errorMsg = ""
		}
		return a, nil

	case ErrorMsg:
		a.loading = false
		a.errorMsg = msg.Error()
		return a, nil

	case ClearErrorMsg:
		a.errorMsg = ""
		return a, nil

	case LoadingMsg:
		a.loading = msg.Loading
		a.loadingMsg = msg.Message
		return a, nil

	case SpinnerTickMsg:
		if a.loading {
			a.spinnerFrame = (a.spinnerFrame + 1) % len(style.SpinnerFrames)
			return a, a.spinnerTick()
		}
		return a, nil

	case DownloadRequestMsg:
		return a.handleDownloadRequest(msg)

	case BatchDownloadMsg:
		return a.handleBatchDownload(msg)

	case DownloadProgressMsg:
		return a, tea.Batch(cmds...)

	case DownloadQueueToggleMsg:
		a.downloadQueue.Toggle()
		if a.downloadQueue.Visible() && a.downloader != nil {
			a.downloadQueue.SetItems(a.downloader.Queue())
		}
		return a, nil

	case GlobalSearchResultsMsg:
		a.loading = false
		if s, ok := a.globalSearch.(globalSearchScreen); ok {
			return a, s.Update(msg)
		}
		return a, nil
	}

	// Forward to current screen
	model, cmd := a.updateScreen(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	return model, tea.Batch(cmds...)
}
