// Package tui implements the terminal user interface using Bubble Tea.
// It follows The Elm Architecture (TEA) with Model-View-Update pattern.
package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	// Screens are set via dependency injection from main
	login       interface{}
	contentType interface{}
	categories  interface{}
	streams     interface{}
	seasons     interface{}
	episodes    interface{}

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
	downloader    *download.Manager
	downloadQueue *components.DownloadQueue
	program       *tea.Program // needed for progress callbacks
}

// Screen model interfaces for type assertions.
type loginScreen interface {
	Focus() tea.Cmd
	SetSize(width, height int)
	SetError(err string)
	Update(tea.Msg) tea.Cmd
	View() string
}

type contentTypeScreen interface {
	SetSize(width, height int)
	SetUserInfo(info string)
	Update(tea.Msg) tea.Cmd
	View() string
}

type categoriesScreen interface {
	SetClient(client *xc.Client)
	SetContentType(ct ContentType) tea.Cmd
	SetSize(width, height int)
	Update(tea.Msg) tea.Cmd
	View() string
}

type streamsScreen interface {
	SetClient(client *xc.Client)
	SetCategory(cat xc.Category, ct ContentType) tea.Cmd
	SetSize(width, height int)
	Update(tea.Msg) tea.Cmd
	View() string
}

type seasonsScreen interface {
	SetClient(client *xc.Client)
	SetSeries(series xc.Series) tea.Cmd
	SetSize(width, height int)
	Update(tea.Msg) tea.Cmd
	View() string
}

type episodesScreen interface {
	SetClient(client *xc.Client)
	SetSeason(season xc.SeasonInfo, episodes []xc.Episode)
	SetSize(width, height int)
	Update(tea.Msg) tea.Cmd
	View() string
}

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{
		screen:        LoginScreen,
		player:        player.NewManager(),
		downloadQueue: components.NewDownloadQueue(),
	}
}

// SetDownloader sets the download manager.
func (a *App) SetDownloader(dm *download.Manager) {
	a.downloader = dm
	// Set progress callback to send messages to TUI
	dm.SetProgressCallback(func(update download.ProgressUpdate) {
		if a.program != nil {
			a.program.Send(DownloadProgressMsg{
				ID:         update.ID,
				Progress:   update.Progress,
				Downloaded: update.Downloaded,
				Size:       update.Size,
				Status:     update.Status.String(),
				Error:      update.Error,
			})
		}
	})
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

// Init initializes the application.
func (a *App) Init() tea.Cmd {
	if s, ok := a.login.(loginScreen); ok {
		return s.Focus()
	}
	return nil
}

// Update handles messages and updates application state.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateScreenSizes()
		return a, nil

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

		// Handle download queue overlay
		if a.downloadQueue.Visible() {
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

		// Toggle download queue with 'D'
		if msg.String() == "D" && a.screen != LoginScreen {
			a.downloadQueue.Toggle()
			if a.downloadQueue.Visible() && a.downloader != nil {
				a.downloadQueue.SetItems(a.downloader.Queue())
			}
			return a, nil
		}

		// Global back navigation (except on login)
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

	case AuthSuccessMsg:
		a.client = msg.Client
		// Clear password from memory immediately for security
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
		return a.navigateTo(SeasonsScreen)

	case SeriesInfoLoadedMsg:
		a.loading = false
		if s, ok := a.seasons.(seasonsScreen); ok {
			return a, s.Update(msg)
		}
		return a, nil

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
			a.spinnerFrame = (a.spinnerFrame + 1) % len(SpinnerFrames)
			return a, a.spinnerTick()
		}
		return a, nil

	case DownloadRequestMsg:
		if a.downloader != nil {
			a.downloader.Add(msg.Name, msg.URL)
			a.errorMsg = "Added to download queue - Press 'D' to view"
		}
		return a, nil

	case DownloadProgressMsg:
		if a.downloader != nil {
			a.downloadQueue.SetItems(a.downloader.Queue())
		}
		return a, nil
	}

	// Forward to current screen
	return a.updateScreen(msg)
}

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
}

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
	}

	return a, cmd
}

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
			s.SetSeason(a.currentSeason, a.currentSeasonEpisodes)
		}
	}

	// Start spinner if loading
	if a.loading {
		return a, tea.Batch(cmd, a.spinnerTick())
	}
	return a, cmd
}

func (a *App) navigateBack() (*App, tea.Cmd) {
	if len(a.navStack) == 0 {
		return a, nil
	}
	a.screen = a.navStack[len(a.navStack)-1]
	a.navStack = a.navStack[:len(a.navStack)-1]
	a.loading = false
	return a, nil
}

func (a *App) spinnerTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}

// startPlayback launches the media player.
// Note: Player runs in background, mpv handles its own window.
// TUI remains active while player is running.
func (a *App) startPlayback(url, title string) tea.Cmd {
	return func() tea.Msg {
		// Cancel any previous playback context
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

// StopPlayback stops current playback (call from main on quit).
func (a *App) StopPlayback() {
	if a.playerStop != nil {
		a.playerStop()
	}
	a.player.Stop()
}

// View renders the current application state.
func (a *App) View() string {
	// Loading overlay
	if a.loading {
		return a.renderLoading()
	}

	// Error overlay (if error msg exists, show it over current screen)
	content := a.renderScreen()
	if a.errorMsg != "" {
		content = a.renderWithError(content)
	} else if a.screen != LoginScreen && a.userInfo.Username != "" {
		// Add status bar when logged in (not on login screen)
		content = a.renderWithStatusBar(content)
	}

	// Download queue overlay
	if a.downloadQueue.Visible() {
		content = a.renderWithDownloadQueue(content)
	}

	return content
}

// renderWithDownloadQueue renders content with download queue overlay.
func (a *App) renderWithDownloadQueue(content string) string {
	a.downloadQueue.SetSize(a.width, a.height)
	queueView := a.downloadQueue.View()
	if queueView == "" {
		return content
	}

	// Position queue panel on the right side using Place
	overlay := lipgloss.Place(
		a.width, a.height,
		lipgloss.Right, lipgloss.Top,
		queueView,
	)

	return overlay
}

// renderWithStatusBar renders content with a status bar at the bottom.
func (a *App) renderWithStatusBar(content string) string {
	statusBar := a.renderStatusBar()

	// Adjust content height to fit status bar
	contentHeight := a.height - 1
	if contentHeight < 1 {
		contentHeight = 1
	}
	content = lipgloss.NewStyle().Height(contentHeight).Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

func (a *App) renderScreen() string {
	switch a.screen {
	case LoginScreen:
		if s, ok := a.login.(loginScreen); ok {
			return s.View()
		}
	case ContentTypeScreen:
		if s, ok := a.contentType.(contentTypeScreen); ok {
			return s.View()
		}
	case CategoriesScreen:
		if s, ok := a.categories.(categoriesScreen); ok {
			return s.View()
		}
	case StreamsScreen:
		if s, ok := a.streams.(streamsScreen); ok {
			return s.View()
		}
	case SeasonsScreen:
		if s, ok := a.seasons.(seasonsScreen); ok {
			return s.View()
		}
	case EpisodesScreen:
		if s, ok := a.episodes.(episodesScreen); ok {
			return s.View()
		}
	}

	// Fallback placeholder
	return a.renderPlaceholder()
}

func (a *App) renderPlaceholder() string {
	style := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center)

	content := TitleStyle.Render("xstream-tui") + "\n\n" +
		ItemDimStyle.Render("Press 'q' to quit")

	return style.Render(content)
}

func (a *App) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center)

	spinner := SpinnerFrames[a.spinnerFrame]
	msg := a.loadingMsg
	if msg == "" {
		msg = "Loading..."
	}

	content := LoadingStyle.Render(spinner + " " + msg)
	return style.Render(content)
}

func (a *App) renderWithError(content string) string {
	// Show error as a bottom bar
	errorBar := ErrorStyle.Render("⚠ " + a.errorMsg + "  [Esc] Dismiss")
	errorBar = lipgloss.NewStyle().
		Width(a.width).
		Background(lipgloss.Color("52")).
		Padding(0, 2).
		Render(errorBar)

	// Adjust content height and combine
	contentHeight := a.height - 1
	if contentHeight < 1 {
		contentHeight = 1
	}
	content = lipgloss.NewStyle().Height(contentHeight).Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, content, errorBar)
}

// renderStatusBar renders the status bar at the bottom of the screen.
func (a *App) renderStatusBar() string {
	// Build left side: username and content type
	var leftParts []string

	if a.userInfo.Username != "" {
		leftParts = append(leftParts,
			StatusBarLabelStyle.Render("User: ")+a.userInfo.Username)
	}

	// Show content type if on categories or streams screen
	if a.screen == CategoriesScreen || a.screen == StreamsScreen {
		leftParts = append(leftParts,
			StatusBarLabelStyle.Render("Type: ")+a.currentType.String())
	}

	left := strings.Join(leftParts, "  │  ")

	// Build right side: expiration date
	var right string
	expDate := a.userInfo.ExpDate.String()
	if expDate != "" && expDate != "0" {
		right = StatusBarLabelStyle.Render("Expires: ") + expDate
	}

	// Calculate padding for right alignment
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	availableWidth := a.width - 4 // Account for padding

	padding := availableWidth - leftLen - rightLen
	if padding < 1 {
		padding = 1
	}

	content := left + strings.Repeat(" ", padding) + right

	return StatusBarStyle.Width(a.width).Render(content)
}
