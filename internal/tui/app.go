// Package tui implements the terminal user interface using Bubble Tea.
// It follows The Elm Architecture (TEA) with Model-View-Update pattern.
package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/download"
	"github.com/altmueller/xstream-tui/internal/player"
	"github.com/altmueller/xstream-tui/internal/tui/components"
	"github.com/altmueller/xstream-tui/internal/xc"
)

// App is the main application model implementing tea.Model interface.
// It manages screen navigation and coordinates between components.
type App struct {
	screen   Screen
	navStack []Screen
	width    int
	height   int

	// Screen models - using interface{} to avoid circular imports
	login         interface{}
	contentType   interface{}
	categories    interface{}
	streams       interface{}
	seasons       interface{}
	episodes      interface{}
	seriesBrowser interface{}
	globalSearch  interface{}

	// Current series for episodes screen
	currentSeries         xc.Series
	currentSeason         xc.SeasonInfo
	currentSeasonEpisodes []xc.Episode

	// Shared state
	client       *xc.Client
	userInfo     xc.UserInfo
	currentType  ContentType
	currentCat   xc.Category
	errorMsg     string
	loading      bool
	loadingMsg   string
	spinnerFrame int

	// Player
	player     *player.Manager
	playerStop context.CancelFunc

	// Downloader
	downloader     *download.Manager
	downloadDaemon *download.Daemon
	downloadQueue  *components.DownloadQueue
	program        *tea.Program

	// Use split view for series browser
	useSplitView bool
}

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{
		screen:        LoginScreen,
		player:        player.NewManager(),
		downloadQueue: components.NewDownloadQueue(),
		useSplitView:  true,
	}
}

// SetDownloader sets the download manager.
func (a *App) SetDownloader(dm *download.Manager) {
	a.downloader = dm
	a.downloadDaemon = download.NewDaemon(dm)
}

// SetProgram sets the tea.Program for sending messages.
func (a *App) SetProgram(p *tea.Program) {
	a.program = p
}

// SetLoginScreen injects the login screen model.
func (a *App) SetLoginScreen(m interface{}) {
	a.login = m
}

// SetContentTypeScreen injects the content type screen model.
func (a *App) SetContentTypeScreen(m interface{}) {
	a.contentType = m
}

// SetCategoriesScreen injects the categories screen model.
func (a *App) SetCategoriesScreen(m interface{}) {
	a.categories = m
}

// SetStreamsScreen injects the streams screen model.
func (a *App) SetStreamsScreen(m interface{}) {
	a.streams = m
}

// SetSeasonsScreen injects the seasons screen model.
func (a *App) SetSeasonsScreen(m interface{}) {
	a.seasons = m
}

// SetEpisodesScreen injects the episodes screen model.
func (a *App) SetEpisodesScreen(m interface{}) {
	a.episodes = m
}

// SetSeriesBrowserScreen injects the series browser screen model.
func (a *App) SetSeriesBrowserScreen(m interface{}) {
	a.seriesBrowser = m
}

// SetGlobalSearchScreen injects the global search screen model.
func (a *App) SetGlobalSearchScreen(m interface{}) {
	a.globalSearch = m
}

// Init initializes the application.
func (a *App) Init() tea.Cmd {
	var cmds []tea.Cmd

	if s, ok := a.login.(loginScreen); ok {
		cmds = append(cmds, s.Focus())
	}

	if a.downloadDaemon != nil {
		cmds = append(cmds, a.downloadDaemon.Start())
	}

	return tea.Batch(cmds...)
}
