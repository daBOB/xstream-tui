package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyMsg processes keyboard input.
func (a *App) handleKeyMsg(msg tea.KeyMsg) (*App, tea.Cmd) {
	// Global quit
	if msg.String() == "ctrl+c" {
		return a, tea.Quit
	}

	// Handle download queue overlay
	if a.downloadQueue.Visible() {
		return a.handleDownloadQueueKey(msg)
	}

	// Toggle download queue with 'Q'
	if msg.String() == "Q" && a.screen != LoginScreen {
		a.downloadQueue.Toggle()
		if a.downloadQueue.Visible() && a.downloader != nil {
			a.downloadQueue.SetItems(a.downloader.Queue())
		}
		return a, nil
	}

	// Global search with Ctrl+F
	if msg.String() == "ctrl+f" && a.screen != LoginScreen && a.screen != ContentTypeScreen {
		return a.navigateTo(GlobalSearchScreen)
	}

	// Global back navigation
	if msg.String() == "esc" && a.screen != LoginScreen {
		if a.errorMsg != "" {
			a.errorMsg = ""
			return a, nil
		}
		return a.navigateBack()
	}

	// Quit on 'q' only from login/content type screens
	if msg.String() == "q" && (a.screen == LoginScreen || a.screen == ContentTypeScreen) {
		return a, tea.Quit
	}

	return a, nil
}

// handleDownloadQueueKey handles keys when download queue is visible.
func (a *App) handleDownloadQueueKey(msg tea.KeyMsg) (*App, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.downloadQueue.Hide()
		return a, nil
	case "d":
		if id := a.downloadQueue.SelectedID(); id != "" && a.downloader != nil {
			a.downloader.Cancel(id)
			a.downloadQueue.SetItems(a.downloader.Queue())
		}
		return a, nil
	case "x":
		if id := a.downloadQueue.SelectedID(); id != "" && a.downloader != nil {
			a.downloader.Remove(id)
			a.downloadQueue.SetItems(a.downloader.Queue())
		}
		return a, nil
	default:
		a.downloadQueue.Update(msg)
		return a, nil
	}
}

// handleDownloadRequest adds a single download to the queue.
func (a *App) handleDownloadRequest(msg DownloadRequestMsg) (*App, tea.Cmd) {
	if a.downloader != nil {
		var subPath string
		if msg.SeriesName != "" {
			subPath = msg.SeriesName
			if msg.SeasonName != "" {
				subPath = subPath + "/" + msg.SeasonName
			}
		}
		a.downloader.AddWithPath(msg.Name, msg.URL, subPath)
		a.errorMsg = "Added to download queue - Press 'D' to view"
	}
	return a, nil
}

// handleBatchDownload adds multiple downloads to the queue.
func (a *App) handleBatchDownload(msg BatchDownloadMsg) (*App, tea.Cmd) {
	if a.downloader != nil && len(msg.Downloads) > 0 {
		for _, dl := range msg.Downloads {
			var subPath string
			if dl.SeriesName != "" {
				subPath = dl.SeriesName
				if dl.SeasonName != "" {
					subPath = subPath + "/" + dl.SeasonName
				}
			}
			a.downloader.AddWithPath(dl.Name, dl.URL, subPath)
		}
		a.errorMsg = fmt.Sprintf("Added %d episodes to queue - Press 'Q' to view", len(msg.Downloads))
	}
	return a, nil
}

// updateScreenSizes updates all screen dimensions.
func (a *App) updateScreenSizes() {
	if s, ok := a.login.(loginScreen); ok {
		s.SetSize(a.width, a.height)
	}
	if s, ok := a.contentType.(contentTypeScreen); ok {
		s.SetSize(a.width, a.height)
	}
	if s, ok := a.categories.(categoriesScreen); ok {
		s.SetSize(a.width, a.height)
	}
	if s, ok := a.streams.(streamsScreen); ok {
		s.SetSize(a.width, a.height)
	}
	if s, ok := a.seasons.(seasonsScreen); ok {
		s.SetSize(a.width, a.height)
	}
	if s, ok := a.episodes.(episodesScreen); ok {
		s.SetSize(a.width, a.height)
	}
	if s, ok := a.seriesBrowser.(seriesBrowserScreen); ok {
		s.SetSize(a.width, a.height)
	}
	if s, ok := a.globalSearch.(globalSearchScreen); ok {
		s.SetSize(a.width, a.height)
	}
}

// updateScreen forwards messages to current screen.
func (a *App) updateScreen(msg tea.Msg) (*App, tea.Cmd) {
	var cmd tea.Cmd

	switch a.screen {
	case LoginScreen:
		if s, ok := a.login.(loginScreen); ok {
			cmd = s.Update(msg)
		}
	case ContentTypeScreen:
		if s, ok := a.contentType.(contentTypeScreen); ok {
			cmd = s.Update(msg)
		}
	case CategoriesScreen:
		if s, ok := a.categories.(categoriesScreen); ok {
			cmd = s.Update(msg)
		}
	case StreamsScreen:
		if s, ok := a.streams.(streamsScreen); ok {
			cmd = s.Update(msg)
		}
	case SeasonsScreen:
		if s, ok := a.seasons.(seasonsScreen); ok {
			cmd = s.Update(msg)
		}
	case EpisodesScreen:
		if s, ok := a.episodes.(episodesScreen); ok {
			cmd = s.Update(msg)
		}
	case SeriesBrowserScreen:
		if s, ok := a.seriesBrowser.(seriesBrowserScreen); ok {
			cmd = s.Update(msg)
		}
	case GlobalSearchScreen:
		if s, ok := a.globalSearch.(globalSearchScreen); ok {
			cmd = s.Update(msg)
		}
	}

	return a, cmd
}
