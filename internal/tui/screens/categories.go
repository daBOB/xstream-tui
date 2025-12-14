package screens

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/components"
	"github.com/altmueller/xstream-tui/internal/xc"
)

// CategoryItem wraps xc.Category to implement ListItem.
type CategoryItem struct {
	xc.Category
}

func (c CategoryItem) FilterValue() string { return c.Name }
func (c CategoryItem) Title() string       { return c.Name }
func (c CategoryItem) Description() string { return "" }

// CategoriesModel handles category browsing.
type CategoriesModel struct {
	list        *components.VirtualList
	search      textinput.Model
	searching   bool
	contentType tui.ContentType
	categories  []xc.Category
	width       int
	height      int
	client      *xc.Client
}

// NewCategoriesModel creates a new categories screen.
func NewCategoriesModel() *CategoriesModel {
	search := textinput.New()
	search.Placeholder = "Type to search..."
	search.CharLimit = 50

	return &CategoriesModel{
		list:   components.NewVirtualList(),
		search: search,
	}
}

// SetClient sets the XC API client.
func (m *CategoriesModel) SetClient(client *xc.Client) {
	m.client = client
}

// SetContentType sets the content type and triggers loading.
func (m *CategoriesModel) SetContentType(ct tui.ContentType) tea.Cmd {
	m.contentType = ct
	return m.loadCategories()
}

// SetSize sets the screen dimensions.
func (m *CategoriesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// List gets most of the height, minus header and search
	listHeight := height - 8
	if listHeight < 5 {
		listHeight = 5
	}
	m.list.SetSize(width-4, listHeight)
}

func (m *CategoriesModel) loadCategories() tea.Cmd {
	return func() tea.Msg {
		if m.client == nil {
			return tui.ErrorMsg{Err: xc.ErrInvalidConfig}
		}

		ctx := context.Background()
		var cats []xc.Category
		var err error

		switch m.contentType {
		case tui.LiveContent:
			cats, err = m.client.GetLiveCategories(ctx)
		case tui.VODContent:
			cats, err = m.client.GetVODCategories(ctx)
		case tui.SeriesContent:
			cats, err = m.client.GetSeriesCategories(ctx)
		}

		if err != nil {
			return tui.ErrorMsg{Err: err}
		}

		return tui.CategoriesLoadedMsg{Categories: cats}
	}
}

// SetCategories populates the list with categories.
func (m *CategoriesModel) SetCategories(cats []xc.Category) {
	m.categories = cats
	items := make([]components.ListItem, len(cats))
	for i, c := range cats {
		items[i] = CategoryItem{c}
	}
	m.list.SetItems(items)
}

// Update handles input events.
func (m *CategoriesModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tui.CategoriesLoadedMsg:
		m.SetCategories(msg.Categories)
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
			if item := m.list.Selected(); item != nil {
				cat := item.(CategoryItem).Category
				return func() tea.Msg {
					return tui.CategorySelectedMsg{Category: cat}
				}
			}
		}
	}

	// Update list navigation
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return cmd
}

// View renders the categories screen.
func (m *CategoriesModel) View() string {
	var b strings.Builder

	// Header with content type
	icon := "📡"
	switch m.contentType {
	case tui.VODContent:
		icon = "🎬"
	case tui.SeriesContent:
		icon = "📺"
	}
	title := tui.TitleStyle.Render(icon + " " + m.contentType.String() + " Categories")
	b.WriteString(title)
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

	// Category list
	b.WriteString(m.list.View())
	b.WriteString("\n")

	// Scroll info
	b.WriteString(m.list.ScrollInfo())

	// Help text
	help := tui.HelpStyle.Render("\n[↑↓jk] Navigate  [Enter] Select  [/] Search  [Esc] Back")
	b.WriteString(help)

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
