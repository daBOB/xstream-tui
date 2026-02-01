package screens

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/components"
	"github.com/altmueller/xstream-tui/internal/tui/style"
	"github.com/altmueller/xstream-tui/internal/xc"
)

// GlobalSearchItem wraps different content types for unified search results.
type GlobalSearchItem struct {
	name        string
	description string
	itemType    tui.ContentType
	liveStream  *xc.LiveStream
	vodStream   *xc.VODStream
	series      *xc.Series
}

func (g GlobalSearchItem) FilterValue() string { return g.name }
func (g GlobalSearchItem) Title() string       { return g.name }
func (g GlobalSearchItem) Description() string { return g.description }

// GlobalSearchModel handles global search across all content.
type GlobalSearchModel struct {
	list        *components.VirtualList
	search      textinput.Model
	contentType tui.ContentType
	allItems    []GlobalSearchItem // All items loaded from API
	categories  map[string]string  // CategoryID -> CategoryName lookup
	width       int
	height      int
	client      *xc.Client
	loading     bool
	loaded      bool
}

// NewGlobalSearchModel creates a new global search screen.
func NewGlobalSearchModel() *GlobalSearchModel {
	search := textinput.New()
	search.Placeholder = "Search..."
	search.CharLimit = 100
	search.Width = 40

	return &GlobalSearchModel{
		list:   components.NewVirtualList(),
		search: search,
	}
}

// SetClient sets the XC API client.
func (m *GlobalSearchModel) SetClient(client *xc.Client) {
	m.client = client
}

// SetContentType sets the content type and triggers loading all content.
func (m *GlobalSearchModel) SetContentType(ct tui.ContentType) tea.Cmd {
	m.contentType = ct
	m.loading = true
	m.loaded = false
	m.allItems = nil
	m.list.SetItems(nil)
	return tea.Batch(m.search.Focus(), m.loadAllContent())
}

// SetSize sets the screen dimensions.
func (m *GlobalSearchModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	listHeight := height - 10
	if listHeight < 5 {
		listHeight = 5
	}
	m.list.SetSize(width-4, listHeight)
	m.search.Width = width - 10
}

func (m *GlobalSearchModel) loadAllContent() tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return tui.ErrorMsg{Err: xc.ErrInvalidConfig}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var msg tui.GlobalSearchResultsMsg

		switch m.contentType {
		case tui.LiveContent:
			streams, err := m.client.GetAllLiveStreams(ctx)
			if err != nil {
				return tui.ErrorMsg{Err: err}
			}
			msg.LiveStreams = streams
			// Fetch categories for lookup
			cats, _ := m.client.GetLiveCategories(ctx)
			msg.Categories = cats

		case tui.VODContent:
			streams, err := m.client.GetAllVODStreams(ctx)
			if err != nil {
				return tui.ErrorMsg{Err: err}
			}
			msg.VODStreams = streams
			// Fetch categories for lookup
			cats, _ := m.client.GetVODCategories(ctx)
			msg.Categories = cats

		case tui.SeriesContent:
			series, err := m.client.GetAllSeries(ctx)
			if err != nil {
				return tui.ErrorMsg{Err: err}
			}
			msg.Series = series
			// Fetch categories for lookup
			cats, _ := m.client.GetSeriesCategories(ctx)
			msg.Categories = cats
		}

		return msg
	}
}

// SetResults populates items from search results.
func (m *GlobalSearchModel) SetResults(msg tui.GlobalSearchResultsMsg) {
	m.loading = false
	m.loaded = true
	m.allItems = make([]GlobalSearchItem, 0)

	// Build category lookup map
	m.categories = make(map[string]string)
	for _, cat := range msg.Categories {
		m.categories[cat.ID.String()] = cat.Name
	}

	// Add live streams
	for i := range msg.LiveStreams {
		s := &msg.LiveStreams[i]
		desc := m.categories[s.CategoryID.String()]
		m.allItems = append(m.allItems, GlobalSearchItem{
			name:        s.Name,
			description: desc,
			itemType:    tui.LiveContent,
			liveStream:  s,
		})
	}

	// Add VOD streams
	for i := range msg.VODStreams {
		s := &msg.VODStreams[i]
		desc := m.categories[s.CategoryID.String()]
		m.allItems = append(m.allItems, GlobalSearchItem{
			name:        s.Name,
			description: desc,
			itemType:    tui.VODContent,
			vodStream:   s,
		})
	}

	// Add series
	for i := range msg.Series {
		s := &msg.Series[i]
		desc := m.categories[s.CategoryID.String()]
		m.allItems = append(m.allItems, GlobalSearchItem{
			name:        s.Name,
			description: desc,
			itemType:    tui.SeriesContent,
			series:      s,
		})
	}

	// Apply current filter
	m.filterItems()
}

func (m *GlobalSearchModel) filterItems() {
	query := strings.ToLower(strings.TrimSpace(m.search.Value()))
	var items []components.ListItem

	for i := range m.allItems {
		item := &m.allItems[i]
		if query == "" || strings.Contains(strings.ToLower(item.name), query) {
			items = append(items, item)
		}
	}

	m.list.SetItems(items)
}

// Update handles input events.
func (m *GlobalSearchModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tui.GlobalSearchResultsMsg:
		m.SetResults(msg)
		return nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// If search has text, clear it first
			if m.search.Value() != "" {
				m.search.SetValue("")
				m.filterItems()
				return nil
			}
			// Otherwise, signal to close search (handled by app.go)
			return nil
		case "enter":
			return m.selectItem()
		case "down":
			m.list.Update(msg)
			return nil
		case "up":
			m.list.Update(msg)
			return nil
		case "pgdown", "ctrl+d":
			m.list.Update(msg)
			return nil
		case "pgup", "ctrl+u":
			m.list.Update(msg)
			return nil
		case "d":
			return m.downloadItem()
		default:
			// Update search input (all other keys go to text input)
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			m.filterItems()
			return cmd
		}
	}

	return nil
}

func (m *GlobalSearchModel) selectItem() tea.Cmd {
	item := m.list.Selected()
	if item == nil {
		return nil
	}

	searchItem, ok := item.(*GlobalSearchItem)
	if !ok {
		return nil
	}

	switch searchItem.itemType {
	case tui.LiveContent:
		if searchItem.liveStream != nil {
			url := m.client.LiveStreamURL(searchItem.liveStream.ID.String())
			return func() tea.Msg {
				return tui.StreamSelectedMsg{URL: url, Name: searchItem.liveStream.Name}
			}
		}
	case tui.VODContent:
		if searchItem.vodStream != nil {
			container := searchItem.vodStream.Container
			if container == "" {
				container = "mp4"
			}
			url := m.client.VODStreamURL(searchItem.vodStream.ID.String(), container)
			return func() tea.Msg {
				return tui.StreamSelectedMsg{URL: url, Name: searchItem.vodStream.Name}
			}
		}
	case tui.SeriesContent:
		if searchItem.series != nil {
			return func() tea.Msg {
				return tui.SeriesSelectedMsg{Series: *searchItem.series}
			}
		}
	}

	return nil
}

func (m *GlobalSearchModel) downloadItem() tea.Cmd {
	item := m.list.Selected()
	if item == nil {
		return nil
	}

	searchItem, ok := item.(*GlobalSearchItem)
	if !ok {
		return nil
	}

	// Only VOD content can be downloaded
	if searchItem.itemType == tui.VODContent && searchItem.vodStream != nil {
		container := searchItem.vodStream.Container
		if container == "" {
			container = "mp4"
		}
		url := m.client.VODStreamURL(searchItem.vodStream.ID.String(), container)
		name := searchItem.vodStream.Name + "." + container
		return func() tea.Msg {
			return tui.DownloadRequestMsg{Name: name, URL: url}
		}
	}

	return nil
}

// View renders the global search screen.
func (m *GlobalSearchModel) View() string {
	var b strings.Builder

	// Header
	icon := ">"
	switch m.contentType {
	case tui.LiveContent:
		icon = ">"
	case tui.VODContent:
		icon = ">"
	case tui.SeriesContent:
		icon = ">"
	}
	title := style.TitleStyle.Render(icon + " Global Search: " + m.contentType.String())
	b.WriteString(title)
	b.WriteString("\n")

	// Search input (always focused)
	searchBox := style.InputFocusedStyle.Render("> " + m.search.View())
	b.WriteString(searchBox)
	b.WriteString("\n\n")

	// Loading state
	if m.loading {
		loadingText := style.LoadingStyle.Render("Loading content...")
		b.WriteString(loadingText)
	} else if !m.loaded {
		b.WriteString(style.ItemDimStyle.Render("Initializing..."))
	} else {
		// Results list
		b.WriteString(m.list.View())
		b.WriteString("\n")

		// Scroll info
		b.WriteString(m.list.ScrollInfo())
	}

	// Help text
	helpText := "[Enter] Play/Select  [d] Download  [Esc] Close/Clear"
	help := style.HelpStyle.Render("\n" + helpText)
	b.WriteString(help)

	return style.AppStyle.Render(b.String())
}
