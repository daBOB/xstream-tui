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
		// Let GlobalSearchScreen handle Esc (clear query or navigate back)
		if a.screen == GlobalSearchScreen {
			return a.updateScreen(msg)
		}
		return a.navigateBack()
	}

	// Quit on 'q' only from login/content type screens
	if msg.String() == "q" && (a.screen == LoginScreen || a.screen == ContentTypeScreen) {
		return a, tea.Quit
	}

	// Forward unhandled keys to current screen
	return a.updateScreen(msg)
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
		if err := a.downloader.AddWithPath(msg.Name, msg.URL, subPath); err != nil {
			a.errorMsg = "Download failed: " + err.Error()
		} else {
			a.errorMsg = "Added to download queue - Press 'Q' to view"
		}
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
			_ = a.downloader.AddWithPath(dl.Name, dl.URL, subPath)
		}
		a.errorMsg = fmt.Sprintf("Added %d episodes to queue - Press 'Q' to view", len(msg.Downloads))
	}
	return a, nil
}

// updateScreenSizes updates all screen dimensions.
func (a *App) updateScreenSizes() {
	if a.login != nil {
		a.login.SetSize(a.width, a.height)
	}
	if a.contentType != nil {
		a.contentType.SetSize(a.width, a.height)
	}
	if a.categories != nil {
		a.categories.SetSize(a.width, a.height)
	}
	if a.streams != nil {
		a.streams.SetSize(a.width, a.height)
	}
	if a.seasons != nil {
		a.seasons.SetSize(a.width, a.height)
	}
	if a.episodes != nil {
		a.episodes.SetSize(a.width, a.height)
	}
	if a.seriesBrowser != nil {
		a.seriesBrowser.SetSize(a.width, a.height)
	}
	if a.globalSearch != nil {
		a.globalSearch.SetSize(a.width, a.height)
	}
}

// updateScreen forwards messages to current screen.
func (a *App) updateScreen(msg tea.Msg) (*App, tea.Cmd) {
	var cmd tea.Cmd

	switch a.screen {
	case LoginScreen:
		if a.login != nil {
			cmd = a.login.Update(msg)
		}
	case ContentTypeScreen:
		if a.contentType != nil {
			cmd = a.contentType.Update(msg)
		}
	case CategoriesScreen:
		if a.categories != nil {
			cmd = a.categories.Update(msg)
		}
	case StreamsScreen:
		if a.streams != nil {
			cmd = a.streams.Update(msg)
		}
	case SeasonsScreen:
		if a.seasons != nil {
			cmd = a.seasons.Update(msg)
		}
	case EpisodesScreen:
		if a.episodes != nil {
			cmd = a.episodes.Update(msg)
		}
	case SeriesBrowserScreen:
		if a.seriesBrowser != nil {
			cmd = a.seriesBrowser.Update(msg)
		}
	case GlobalSearchScreen:
		if a.globalSearch != nil {
			cmd = a.globalSearch.Update(msg)
		}
	}

	return a, cmd
}
