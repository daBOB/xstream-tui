package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/xc"
)

// Screen model interfaces for type assertions.
// These interfaces avoid circular imports between tui and screens packages.

type loginScreen interface {
	Focus() tea.Cmd
	AutoLogin() tea.Cmd
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
	SetSeriesName(name string)
	SetSeason(season xc.SeasonInfo, episodes []xc.Episode)
	SetSize(width, height int)
	Update(tea.Msg) tea.Cmd
	View() string
}

type seriesBrowserScreen interface {
	SetClient(client *xc.Client)
	SetSeries(series xc.Series) tea.Cmd
	SetSize(width, height int)
	Update(tea.Msg) tea.Cmd
	View() string
}

type globalSearchScreen interface {
	SetClient(client *xc.Client)
	SetContentType(ct ContentType) tea.Cmd
	SetSize(width, height int)
	Update(tea.Msg) tea.Cmd
	View() string
}
