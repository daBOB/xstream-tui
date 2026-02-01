// Package components provides reusable UI components for the TUI.
package components

import (
	"io"
	"strings"

	"github.com/altmueller/xstream-tui/internal/tui/style"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FancyListItem interface for items that can be displayed in FancyList.
type FancyListItem interface {
	FilterValue() string
	Title() string
	Description() string
}

// fancyItemDelegate renders list items with custom styling.
type fancyItemDelegate struct {
	styles     FancyListStyles
	showDesc   bool
	itemHeight int
}

// FancyListStyles holds styling for the fancy list.
type FancyListStyles struct {
	Title         lipgloss.Style
	TitleSelected lipgloss.Style
	Desc          lipgloss.Style
	DescSelected  lipgloss.Style
	FilterMatch   lipgloss.Style
}

// DefaultFancyListStyles returns default styling.
func DefaultFancyListStyles() FancyListStyles {
	return FancyListStyles{
		Title: lipgloss.NewStyle().
			Foreground(style.ColorText).
			Padding(0, 0, 0, 2),
		TitleSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(style.ColorHighlight).
			Bold(true).
			Padding(0, 0, 0, 1),
		Desc: lipgloss.NewStyle().
			Foreground(style.ColorMuted).
			Padding(0, 0, 0, 4),
		DescSelected: lipgloss.NewStyle().
			Foreground(style.ColorTextDim).
			Padding(0, 0, 0, 3),
		FilterMatch: lipgloss.NewStyle().
			Underline(true),
	}
}

func (d fancyItemDelegate) Height() int {
	if d.showDesc {
		return 2
	}
	return 1
}

func (d fancyItemDelegate) Spacing() int { return 0 }

func (d fancyItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d fancyItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(fancyListItemWrapper)
	if !ok {
		return
	}

	selected := index == m.Index()
	title := i.item.Title()
	desc := i.item.Description()

	// Truncate title if needed
	maxWidth := m.Width() - 6
	if len(title) > maxWidth && maxWidth > 3 {
		title = title[:maxWidth-3] + "..."
	}

	var s strings.Builder

	// Render cursor and title
	if selected {
		s.WriteString("▸ ")
		s.WriteString(d.styles.TitleSelected.Render(title))
	} else {
		s.WriteString(d.styles.Title.Render(title))
	}

	// Render description if enabled and available
	if d.showDesc && desc != "" {
		s.WriteString("\n")
		if selected {
			s.WriteString("  ")
			s.WriteString(d.styles.DescSelected.Render(desc))
		} else {
			s.WriteString(d.styles.Desc.Render(desc))
		}
	}

	io.WriteString(w, s.String())
}

// fancyListItemWrapper wraps FancyListItem to implement list.Item.
type fancyListItemWrapper struct {
	item FancyListItem
}

func (f fancyListItemWrapper) FilterValue() string { return f.item.FilterValue() }

// FancyList is an enhanced list component using bubbles/list.
type FancyList struct {
	list     list.Model
	items    []FancyListItem
	showDesc bool
	styles   FancyListStyles
}

// NewFancyList creates a new fancy list.
func NewFancyList(title string, showDesc bool) *FancyList {
	styles := DefaultFancyListStyles()
	delegate := fancyItemDelegate{
		styles:   styles,
		showDesc: showDesc,
	}

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()

	// Style the list
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(style.ColorPrimary).
		Bold(true).
		MarginBottom(1)

	l.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(style.ColorSecondary)

	l.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(style.ColorSecondary)

	l.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		MarginTop(1)

	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		Italic(true)

	return &FancyList{
		list:     l,
		items:    make([]FancyListItem, 0),
		showDesc: showDesc,
		styles:   styles,
	}
}

// SetItems replaces all items.
func (f *FancyList) SetItems(items []FancyListItem) {
	f.items = items
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = fancyListItemWrapper{item: item}
	}
	f.list.SetItems(listItems)
}

// SetSize sets the dimensions.
func (f *FancyList) SetSize(width, height int) {
	f.list.SetSize(width, height)
}

// Selected returns the currently selected item.
func (f *FancyList) Selected() FancyListItem {
	item := f.list.SelectedItem()
	if item == nil {
		return nil
	}
	if wrapper, ok := item.(fancyListItemWrapper); ok {
		return wrapper.item
	}
	return nil
}

// Update handles input.
func (f *FancyList) Update(msg tea.Msg) (*FancyList, tea.Cmd) {
	var cmd tea.Cmd
	f.list, cmd = f.list.Update(msg)
	return f, cmd
}

// View renders the list.
func (f *FancyList) View() string {
	return f.list.View()
}

// Filtering returns whether filtering is active.
func (f *FancyList) Filtering() bool {
	return f.list.FilterState() == list.Filtering
}

// SetTitle sets the list title.
func (f *FancyList) SetTitle(title string) {
	f.list.Title = title
}

// Len returns the number of items.
func (f *FancyList) Len() int {
	return len(f.items)
}

// Index returns the current cursor index.
func (f *FancyList) Index() int {
	return f.list.Index()
}
