// Package screens provides TUI screen models.
package screens

import (
	"context"
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

	series      xc.Series
	seriesInfo  *xc.SeriesInfo
	allEpisodes map[string][]xc.Episode
	loading     bool

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
	m.loading = true
	m.splitView.SetActive(components.LeftPane)

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

func (m *SeriesBrowserModel) loadSeriesInfo() tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return tui.ErrorMsg{Err: xc.ErrInvalidConfig}
		}

		ctx := context.Background()
		info, err := m.client.GetSeriesInfo(ctx, m.series.ID.String())
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		return tui.SeriesInfoLoadedMsg{Info: info}
	}
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
				// Selecting a season populates episodes
				m.updateEpisodesForSelectedSeason()
				m.splitView.SetActive(components.RightPane)
			} else {
				// Play episode
				return m.playSelectedEpisode()
			}
		case "d":
			if m.splitView.Active() == components.RightPane {
				return m.downloadSelectedEpisode()
			}
		}

		// Route input to active pane
		if m.splitView.Active() == components.LeftPane {
			var cmd tea.Cmd
			m.seasons, cmd = m.seasons.Update(msg)
			cmds = append(cmds, cmd)

			// Update episodes when season selection changes
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

func (m *SeriesBrowserModel) populateSeasons() {
	if m.seriesInfo == nil {
		return
	}

	items := make([]components.FancyListItem, len(m.seriesInfo.Seasons))
	for i, s := range m.seriesInfo.Seasons {
		seasonNum := s.SeasonNumber.String()
		actualCount := len(m.seriesInfo.Episodes[seasonNum])
		items[i] = seasonBrowserItem{
			SeasonInfo:         s,
			ActualEpisodeCount: actualCount,
		}
	}
	m.seasons.SetItems(items)
}

func (m *SeriesBrowserModel) selectFirstSeason() {
	if m.seasons.Len() > 0 {
		m.updateEpisodesForSelectedSeason()
	}
}

func (m *SeriesBrowserModel) updateEpisodesForSelectedSeason() {
	item := m.seasons.Selected()
	if item == nil {
		return
	}

	seasonItem, ok := item.(seasonBrowserItem)
	if !ok {
		return
	}

	seasonNum := seasonItem.SeasonNumber.String()
	episodes := m.allEpisodes[seasonNum]

	items := make([]components.FancyListItem, len(episodes))
	for i, ep := range episodes {
		items[i] = episodeBrowserItem{Episode: ep}
	}
	m.episodes.SetItems(items)
}

func (m *SeriesBrowserModel) playSelectedEpisode() tea.Cmd {
	item := m.episodes.Selected()
	if item == nil {
		return nil
	}

	epItem, ok := item.(episodeBrowserItem)
	if !ok {
		return nil
	}

	return func() tea.Msg {
		return tui.EpisodeSelectedMsg{Episode: epItem.Episode}
	}
}

func (m *SeriesBrowserModel) downloadSelectedEpisode() tea.Cmd {
	item := m.episodes.Selected()
	if item == nil || m.client == nil {
		return nil
	}

	epItem, ok := item.(episodeBrowserItem)
	if !ok {
		return nil
	}

	ep := epItem.Episode
	container := ep.ContainerExt
	if container == "" {
		container = "mp4"
	}
	url := m.client.SeriesEpisodeURL(ep.ID.String(), container)
	name := ep.Title + "." + container

	return func() tea.Msg {
		return tui.DownloadRequestMsg{Name: name, URL: url}
	}
}

// View renders the series browser.
func (m *SeriesBrowserModel) View() string {
	var b strings.Builder

	// Header
	title := tui.TitleStyle.Render("📺 " + m.series.Name)
	b.WriteString(title)
	b.WriteString("\n")

	// Series info
	if m.series.Rating != "" {
		info := tui.ItemDimStyle.Render("★ " + m.series.Rating)
		b.WriteString(info)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Loading state
	if m.loading {
		b.WriteString(m.spinner.View() + " Loading series info...")
		b.WriteString("\n")
	} else if m.seriesInfo == nil {
		b.WriteString(tui.ItemDimStyle.Render("No series data"))
	} else {
		// Set titles with counts
		seasonTitle := fmt.Sprintf("Seasons (%d)", m.seasons.Len())
		episodeTitle := fmt.Sprintf("Episodes (%d)", m.episodes.Len())
		m.splitView.SetTitles(seasonTitle, episodeTitle)

		// Render split view
		styles := components.DefaultSplitViewStyles()
		splitContent := m.splitView.Render(
			m.seasons.View(),
			m.episodes.View(),
			styles,
		)
		b.WriteString(splitContent)
	}

	b.WriteString("\n")

	// Help text
	help := "[Tab/hl] Switch pane  [↑↓jk] Navigate  [Enter] Select/Play  [d] Download  [D] Queue  [Esc] Back"
	b.WriteString(tui.HelpStyle.Render(help))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

// seasonBrowserItem wraps SeasonInfo for FancyList.
type seasonBrowserItem struct {
	xc.SeasonInfo
	ActualEpisodeCount int
}

func (s seasonBrowserItem) FilterValue() string { return s.Name }
func (s seasonBrowserItem) Title() string       { return s.Name }
func (s seasonBrowserItem) Description() string {
	return fmt.Sprintf("%d episodes", s.ActualEpisodeCount)
}

// episodeBrowserItem wraps Episode for FancyList.
type episodeBrowserItem struct {
	xc.Episode
}

func (e episodeBrowserItem) FilterValue() string { return e.Episode.Title }
func (e episodeBrowserItem) Title() string       { return e.Episode.Title }
func (e episodeBrowserItem) Description() string {
	if e.Info.Duration != "" {
		return "⏱ " + e.Info.Duration
	}
	return ""
}
