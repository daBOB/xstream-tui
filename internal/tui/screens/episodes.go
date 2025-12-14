package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/components"
	"github.com/altmueller/xstream-tui/internal/xc"
)

// EpisodeItem wraps Episode for list display.
type EpisodeItem struct {
	xc.Episode
}

func (e EpisodeItem) FilterValue() string { return e.Episode.Title }
func (e EpisodeItem) Title() string       { return e.Episode.Title }
func (e EpisodeItem) Description() string {
	if e.Info.Duration != "" {
		return "⏱ " + e.Info.Duration
	}
	return ""
}

// EpisodesModel handles episode browsing for a season.
type EpisodesModel struct {
	list      *components.VirtualList
	search    textinput.Model
	searching bool
	season    xc.SeasonInfo
	episodes  []xc.Episode
	width     int
	height    int
	client    *xc.Client
}

// NewEpisodesModel creates a new episodes screen.
func NewEpisodesModel() *EpisodesModel {
	search := textinput.New()
	search.Placeholder = "Type to search..."
	search.CharLimit = 50

	return &EpisodesModel{
		list:   components.NewVirtualList(),
		search: search,
	}
}

// SetClient sets the XC API client.
func (m *EpisodesModel) SetClient(client *xc.Client) {
	m.client = client
}

// SetSeason sets the season and its episodes.
func (m *EpisodesModel) SetSeason(season xc.SeasonInfo, episodes []xc.Episode) {
	m.season = season
	m.episodes = episodes
	m.populateList()
}

// SetSize sets the screen dimensions.
func (m *EpisodesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	listHeight := height - 10
	if listHeight < 5 {
		listHeight = 5
	}
	m.list.SetSize(width-4, listHeight)
}

func (m *EpisodesModel) populateList() {
	items := make([]components.ListItem, len(m.episodes))
	for i, ep := range m.episodes {
		items[i] = EpisodeItem{ep}
	}
	m.list.SetItems(items)
}

// Update handles input events.
func (m *EpisodesModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.search.Blur()
				m.list.ClearFilter()
				return nil
			case "enter":
				m.searching = false
				m.search.Blur()
				return nil
			default:
				var cmd tea.Cmd
				m.search, cmd = m.search.Update(msg)
				m.list.Filter(m.search.Value())
				return cmd
			}
		}

		switch msg.String() {
		case "/":
			m.searching = true
			m.search.SetValue("")
			return m.search.Focus()
		case "enter":
			return m.selectEpisode()
		case "d":
			return m.downloadEpisode()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return cmd
}

func (m *EpisodesModel) selectEpisode() tea.Cmd {
	item := m.list.Selected()
	if item == nil {
		return nil
	}

	episodeItem, ok := item.(EpisodeItem)
	if !ok {
		return nil
	}

	return func() tea.Msg {
		return tui.EpisodeSelectedMsg{Episode: episodeItem.Episode}
	}
}

func (m *EpisodesModel) downloadEpisode() tea.Cmd {
	item := m.list.Selected()
	if item == nil || m.client == nil {
		return nil
	}

	episodeItem, ok := item.(EpisodeItem)
	if !ok {
		return nil
	}

	ep := episodeItem.Episode
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

// View renders the episodes screen.
func (m *EpisodesModel) View() string {
	var b strings.Builder

	// Header with season name
	title := tui.TitleStyle.Render("📺 " + m.season.Name)
	b.WriteString(title)
	b.WriteString("\n")

	// Season info
	info := tui.ItemDimStyle.Render(fmt.Sprintf("%d episodes", len(m.episodes)))
	b.WriteString(info)
	b.WriteString("\n")

	// Search bar
	if m.searching {
		searchBox := tui.InputFocusedStyle.Render("🔍 " + m.search.View())
		b.WriteString(searchBox)
	} else {
		hint := tui.ItemDimStyle.Render("[/] Search")
		b.WriteString(hint)
	}
	b.WriteString("\n\n")

	// Episodes header
	b.WriteString(tui.SubtitleStyle.Render("Episodes"))
	b.WriteString("\n")

	// Episode list
	b.WriteString(m.list.View())
	b.WriteString("\n")

	// Scroll info
	b.WriteString(m.list.ScrollInfo())

	// Help text
	help := tui.HelpStyle.Render("\n[↑↓jk] Navigate  [Enter] Play  [d] Download  [/] Search  [D] Queue  [Esc] Back")
	b.WriteString(help)

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
