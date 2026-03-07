package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/components"
	"github.com/altmueller/xstream-tui/internal/tui/style"
	"github.com/altmueller/xstream-tui/internal/xc"
)

// SeasonItem wraps SeasonInfo for list display.
type SeasonItem struct {
	xc.SeasonInfo
	ActualEpisodeCount int // actual count from episodes map
}

func (s SeasonItem) FilterValue() string { return s.Name }
func (s SeasonItem) Title() string       { return s.Name }
func (s SeasonItem) Description() string {
	return fmt.Sprintf("%d episodes", s.ActualEpisodeCount)
}

// SeasonsModel handles season browsing for a series.
type SeasonsModel struct {
	list       *components.VirtualList
	search     textinput.Model
	searching  bool
	series     xc.Series
	seriesInfo *xc.SeriesInfo
	width      int
	height     int
	client     *xc.Client
}

// NewSeasonsModel creates a new seasons screen.
func NewSeasonsModel() *SeasonsModel {
	search := textinput.New()
	search.Placeholder = "Type to search..."
	search.CharLimit = 50

	return &SeasonsModel{
		list:   components.NewVirtualList(),
		search: search,
	}
}

// SetClient sets the XC API client.
func (m *SeasonsModel) SetClient(client *xc.Client) {
	m.client = client
}

// SetSeries sets the series and loads its info.
func (m *SeasonsModel) SetSeries(series xc.Series) tea.Cmd {
	m.series = series
	return m.loadSeriesInfo()
}

// SetSize sets the screen dimensions.
func (m *SeasonsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	listHeight := height - 10
	if listHeight < 5 {
		listHeight = 5
	}
	m.list.SetSize(width-4, listHeight)
}

func (m *SeasonsModel) loadSeriesInfo() tea.Cmd {
	client := m.client
	seriesID := m.series.ID.String()

	return func() tea.Msg {
		if client == nil {
			return tui.ErrorMsg{Err: xc.ErrInvalidConfig}
		}

		ctx, cancel := context.WithTimeout(context.Background(), tui.DefaultTimeout)
		defer cancel()

		info, err := client.GetSeriesInfo(ctx, seriesID)
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		return tui.SeriesInfoLoadedMsg{Info: info}
	}
}

// SetSeriesInfo populates the list with seasons.
func (m *SeasonsModel) SetSeriesInfo(info *xc.SeriesInfo) {
	m.seriesInfo = info
	items := make([]components.ListItem, len(info.Seasons))
	for i, s := range info.Seasons {
		// Get actual episode count from episodes map
		seasonNum := s.SeasonNumber.String()
		actualCount := len(info.Episodes[seasonNum])
		items[i] = SeasonItem{SeasonInfo: s, ActualEpisodeCount: actualCount}
	}
	m.list.SetItems(items)
}

// Update handles input events.
func (m *SeasonsModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tui.SeriesInfoLoadedMsg:
		m.SetSeriesInfo(msg.Info)
		return nil

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
			return m.selectSeason()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return cmd
}

func (m *SeasonsModel) selectSeason() tea.Cmd {
	item := m.list.Selected()
	if item == nil || m.seriesInfo == nil {
		return nil
	}

	seasonItem, ok := item.(SeasonItem)
	if !ok {
		return nil
	}

	// Get episodes for this season
	seasonNum := seasonItem.SeasonNumber.String()
	episodes := m.seriesInfo.Episodes[seasonNum]

	return func() tea.Msg {
		return tui.SeasonSelectedMsg{
			Season:   seasonItem.SeasonInfo,
			Episodes: episodes,
		}
	}
}

// View renders the seasons screen.
func (m *SeasonsModel) View() string {
	var b strings.Builder

	// Header with series name
	title := style.TitleStyle.Render("📺 " + m.series.Name)
	b.WriteString(title)
	b.WriteString("\n")

	// Series info (plot/rating if available)
	if m.series.Rating != "" {
		info := style.ItemDimStyle.Render("★ " + m.series.Rating)
		b.WriteString(info)
		b.WriteString("\n")
	}

	// Search bar
	if m.searching {
		searchBox := style.InputFocusedStyle.Render("🔍 " + m.search.View())
		b.WriteString(searchBox)
	} else {
		hint := style.ItemDimStyle.Render("[/] Search")
		b.WriteString(hint)
	}
	b.WriteString("\n\n")

	// Seasons header
	b.WriteString(style.SubtitleStyle.Render("Seasons"))
	b.WriteString("\n")

	// Season list
	b.WriteString(m.list.View())
	b.WriteString("\n")

	// Scroll info
	b.WriteString(m.list.ScrollInfo())

	// Help text
	help := style.HelpStyle.Render("\n[↑↓jk] Navigate  [Enter] View Episodes  [/] Search  [Esc] Back")
	b.WriteString(help)

	return style.AppStyle.Render(b.String())
}
