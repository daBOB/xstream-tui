package screens

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/components"
	"github.com/altmueller/xstream-tui/internal/tui/style"
	"github.com/altmueller/xstream-tui/internal/xc"
)

// Stream item wrappers for different content types.

type LiveStreamItem struct{ xc.LiveStream }

func (s LiveStreamItem) FilterValue() string { return s.Name }
func (s LiveStreamItem) Title() string       { return s.Name }
func (s LiveStreamItem) Description() string { return s.EPGChannelID }

type VODStreamItem struct{ xc.VODStream }

func (s VODStreamItem) FilterValue() string { return s.Name }
func (s VODStreamItem) Title() string       { return s.Name }
func (s VODStreamItem) Description() string {
	if s.Rating != "" {
		return "★ " + s.Rating
	}
	return ""
}

type SeriesItem struct{ xc.Series }

func (s SeriesItem) FilterValue() string { return s.Name }
func (s SeriesItem) Title() string       { return s.Name }
func (s SeriesItem) Description() string {
	if s.Rating != "" {
		return "★ " + s.Rating
	}
	return ""
}

// StreamsModel handles stream browsing within a category.
type StreamsModel struct {
	list        *components.VirtualList
	search      textinput.Model
	searching   bool
	contentType tui.ContentType
	category    xc.Category
	liveStreams []xc.LiveStream
	vodStreams  []xc.VODStream
	series      []xc.Series
	width       int
	height      int
	client      *xc.Client
}

// NewStreamsModel creates a new streams screen.
func NewStreamsModel() *StreamsModel {
	search := textinput.New()
	search.Placeholder = "Type to search..."
	search.CharLimit = 50

	return &StreamsModel{
		list:   components.NewVirtualList(),
		search: search,
	}
}

// SetClient sets the XC API client.
func (m *StreamsModel) SetClient(client *xc.Client) {
	m.client = client
}

// SetCategory sets the category and content type, then loads streams.
func (m *StreamsModel) SetCategory(cat xc.Category, ct tui.ContentType) tea.Cmd {
	m.category = cat
	m.contentType = ct
	return m.loadStreams()
}

// SetSize sets the screen dimensions.
func (m *StreamsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	listHeight := height - 8
	if listHeight < 5 {
		listHeight = 5
	}
	m.list.SetSize(width-4, listHeight)
}

func (m *StreamsModel) loadStreams() tea.Cmd {
	// Capture fields locally to avoid accessing model from goroutine
	client := m.client
	contentType := m.contentType
	catID := m.category.ID.String()

	return func() tea.Msg {
		if client == nil {
			return tui.ErrorMsg{Err: xc.ErrInvalidConfig}
		}

		ctx, cancel := context.WithTimeout(context.Background(), tui.DefaultTimeout)
		defer cancel()

		switch contentType {
		case tui.LiveContent:
			streams, err := client.GetLiveStreams(ctx, catID)
			if err != nil {
				return tui.ErrorMsg{Err: err}
			}
			return tui.StreamsLoadedMsg{LiveStreams: streams}

		case tui.VODContent:
			streams, err := client.GetVODStreams(ctx, catID)
			if err != nil {
				return tui.ErrorMsg{Err: err}
			}
			return tui.StreamsLoadedMsg{VODStreams: streams}

		case tui.SeriesContent:
			series, err := client.GetSeries(ctx, catID)
			if err != nil {
				return tui.ErrorMsg{Err: err}
			}
			return tui.StreamsLoadedMsg{Series: series}
		}

		return nil
	}
}

// SetStreams populates the list with streams.
func (m *StreamsModel) SetStreams(msg tui.StreamsLoadedMsg) {
	var items []components.ListItem

	switch {
	case len(msg.LiveStreams) > 0:
		m.liveStreams = msg.LiveStreams
		items = make([]components.ListItem, len(m.liveStreams))
		for i, s := range m.liveStreams {
			items[i] = LiveStreamItem{s}
		}

	case len(msg.VODStreams) > 0:
		m.vodStreams = msg.VODStreams
		items = make([]components.ListItem, len(m.vodStreams))
		for i, s := range m.vodStreams {
			items[i] = VODStreamItem{s}
		}

	case len(msg.Series) > 0:
		m.series = msg.Series
		items = make([]components.ListItem, len(m.series))
		for i, s := range m.series {
			items[i] = SeriesItem{s}
		}
	}

	m.list.SetItems(items)
}

// Update handles input events.
func (m *StreamsModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tui.StreamsLoadedMsg:
		m.SetStreams(msg)
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
			return m.selectStream()
		case "d":
			return m.downloadStream()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return cmd
}

func (m *StreamsModel) selectStream() tea.Cmd {
	item := m.list.Selected()
	if item == nil {
		return nil
	}

	switch v := item.(type) {
	case LiveStreamItem:
		url := m.client.LiveStreamURL(v.ID.String())
		return func() tea.Msg {
			return tui.StreamSelectedMsg{URL: url, Name: v.Name}
		}
	case VODStreamItem:
		container := v.Container
		if container == "" {
			container = "mp4"
		}
		url := m.client.VODStreamURL(v.ID.String(), container)
		return func() tea.Msg {
			return tui.StreamSelectedMsg{URL: url, Name: v.Name}
		}
	case SeriesItem:
		return func() tea.Msg {
			return tui.SeriesSelectedMsg{Series: v.Series}
		}
	}

	return nil
}

func (m *StreamsModel) downloadStream() tea.Cmd {
	item := m.list.Selected()
	if item == nil {
		return nil
	}

	// Only VOD content can be downloaded
	switch v := item.(type) {
	case VODStreamItem:
		container := v.Container
		if container == "" {
			container = "mp4"
		}
		url := m.client.VODStreamURL(v.ID.String(), container)
		name := v.Name + "." + container
		return func() tea.Msg {
			return tui.DownloadRequestMsg{Name: name, URL: url}
		}
	}

	return nil
}

// View renders the streams screen.
func (m *StreamsModel) View() string {
	var b strings.Builder

	// Header with category name
	icon := "📡"
	switch m.contentType {
	case tui.VODContent:
		icon = "🎬"
	case tui.SeriesContent:
		icon = "📺"
	}
	title := style.TitleStyle.Render(icon + " " + m.category.Name)
	b.WriteString(title)
	b.WriteString("\n")

	// Search bar
	if m.searching {
		searchBox := style.InputFocusedStyle.Render("🔍 " + m.search.View())
		b.WriteString(searchBox)
	} else {
		hint := style.ItemDimStyle.Render("[/] Search")
		b.WriteString(hint)
	}
	b.WriteString("\n\n")

	// Stream list
	b.WriteString(m.list.View())
	b.WriteString("\n")

	// Scroll info
	b.WriteString(m.list.ScrollInfo())

	// Help text
	help := style.HelpStyle.Render("\n[↑↓jk] Navigate  [Enter] Play  [d] Download  [/] Search  [Ctrl+F] Global  [Q] Queue  [Esc] Back")
	b.WriteString(help)

	return style.AppStyle.Render(b.String())
}
