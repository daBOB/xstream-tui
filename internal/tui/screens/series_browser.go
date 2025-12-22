// Package screens provides TUI screen models.
package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/components"
	"github.com/altmueller/xstream-tui/internal/xc"
)

// SeriesBrowserModel provides a split view for browsing series seasons/episodes.
type SeriesBrowserModel struct {
	splitView *components.SplitView
	seasons   *components.FancyList
	episodes  *components.FancyList
	spinner   spinner.Model

	series        xc.Series
	seriesInfo    *xc.SeriesInfo
	allEpisodes   map[string][]xc.Episode
	currentSeason string
	loading       bool

	width  int
	height int
	client *xc.Client
}

// NewSeriesBrowserModel creates a new series browser.
func NewSeriesBrowserModel() *SeriesBrowserModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))

	return &SeriesBrowserModel{
		splitView:   components.NewSplitView(0.35),
		seasons:     components.NewFancyList("Seasons", true),
		episodes:    components.NewFancyList("Episodes", true),
		spinner:     s,
		allEpisodes: make(map[string][]xc.Episode),
	}
}

// SetClient sets the XC API client.
func (m *SeriesBrowserModel) SetClient(client *xc.Client) {
	m.client = client
}

// SetSeries sets the series and starts loading.
func (m *SeriesBrowserModel) SetSeries(series xc.Series) tea.Cmd {
	m.series = series
	m.seriesInfo = nil
	m.allEpisodes = make(map[string][]xc.Episode)
	m.currentSeason = ""
	m.loading = true
	m.splitView.SetActive(components.LeftPane)

	m.seasons.SetItems(nil)
	m.episodes.SetItems(nil)

	return tea.Batch(
		m.spinner.Tick,
		m.loadSeriesInfo(),
	)
}

// SetSize sets the screen dimensions.
func (m *SeriesBrowserModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.splitView.SetSize(width-4, height-8)

	leftW, leftH := m.splitView.GetLeftDimensions()
	rightW, rightH := m.splitView.GetRightDimensions()

	m.seasons.SetSize(leftW-2, leftH-2)
	m.episodes.SetSize(rightW-2, rightH-2)
}

// Update handles input events.
func (m *SeriesBrowserModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tui.SeriesInfoLoadedMsg:
		m.loading = false
		m.seriesInfo = msg.Info
		m.allEpisodes = msg.Info.Episodes
		m.populateSeasons()
		m.selectFirstSeason()
		return nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		if m.loading {
			return nil
		}

		switch msg.String() {
		case "tab", "l":
			if m.splitView.Active() == components.LeftPane && m.episodes.Len() > 0 {
				m.splitView.SetActive(components.RightPane)
			}
			return nil
		case "shift+tab", "h":
			if m.splitView.Active() == components.RightPane {
				m.splitView.SetActive(components.LeftPane)
			}
			return nil
		case "enter":
			if m.splitView.Active() == components.LeftPane {
				m.updateEpisodesForSelectedSeason()
				m.splitView.SetActive(components.RightPane)
			} else {
				return m.playSelectedEpisode()
			}
		case "d":
			if m.splitView.Active() == components.RightPane {
				return m.downloadSelectedEpisode()
			}
		case "D":
			return m.downloadSelectedSeason()
		case "Q":
			return func() tea.Msg {
				return tui.DownloadQueueToggleMsg{}
			}
		}

		// Route input to active pane
		if m.splitView.Active() == components.LeftPane {
			var cmd tea.Cmd
			m.seasons, cmd = m.seasons.Update(msg)
			cmds = append(cmds, cmd)

			if msg.String() == "up" || msg.String() == "down" ||
				msg.String() == "j" || msg.String() == "k" {
				m.updateEpisodesForSelectedSeason()
			}
		} else {
			var cmd tea.Cmd
			m.episodes, cmd = m.episodes.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

// View renders the series browser.
func (m *SeriesBrowserModel) View() string {
	var b strings.Builder

	title := tui.TitleStyle.Render("📺 " + m.series.Name)
	b.WriteString(title)
	b.WriteString("\n")

	if m.series.Rating != "" {
		info := tui.ItemDimStyle.Render("★ " + m.series.Rating)
		b.WriteString(info)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " Loading series info...")
		b.WriteString("\n")
	} else if m.seriesInfo == nil {
		b.WriteString(tui.ItemDimStyle.Render("No series data"))
	} else {
		seasonTitle := fmt.Sprintf("Seasons (%d)", m.seasons.Len())
		episodeTitle := fmt.Sprintf("Episodes (%d)", m.episodes.Len())
		m.splitView.SetTitles(seasonTitle, episodeTitle)

		styles := components.DefaultSplitViewStyles()
		splitContent := m.splitView.Render(
			m.seasons.View(),
			m.episodes.View(),
			styles,
		)
		b.WriteString(splitContent)
	}

	b.WriteString("\n")

	help := "[Tab/hl] Switch  [↑↓jk] Nav  [Enter] Play  [d] Download  [D] DL Season  [Ctrl+F] Search  [Q] Queue  [Esc] Back"
	b.WriteString(tui.HelpStyle.Render(help))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
