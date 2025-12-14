// Package tui implements the terminal user interface using Bubble Tea.
// It follows The Elm Architecture (TEA) with Model-View-Update pattern.
package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/player"
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
	playerCtx  context.Context
	playerStop context.CancelFunc
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

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{
		screen: LoginScreen,
		player: player.NewManager(),
	}
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
func (a *App) startPlayback(url, title string) tea.Cmd {
	return func() tea.Msg {
		// Cancel any previous playback context
		if a.playerStop != nil {
			a.playerStop()
		}

		ctx, cancel := context.WithCancel(context.Background())
		a.playerCtx = ctx
		a.playerStop = cancel

		// Register exit callback
		a.player.OnExit(func(err error) {
			// Note: This runs in a goroutine, cannot send tea.Msg directly
			// The UI will check IsPlaying() state
		})

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
		return a.renderWithError(content)
	}

	return content
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
