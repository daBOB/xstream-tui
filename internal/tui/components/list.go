// Package components provides reusable UI components for the TUI.
package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem interface for items that can be displayed in VirtualList.
type ListItem interface {
	// FilterValue returns the string used for filtering.
	FilterValue() string
	// Title returns the main display text.
	Title() string
	// Description returns optional secondary text.
	Description() string
}

// VirtualList is a virtualized list component for large datasets.
// It only renders visible items for performance.
type VirtualList struct {
	items    []ListItem
	filtered []int // indices into items
	cursor   int
	offset   int
	height   int
	width    int

	// Styles
	selectedStyle lipgloss.Style
	normalStyle   lipgloss.Style
	dimStyle      lipgloss.Style
}

// NewVirtualList creates a new virtualized list.
func NewVirtualList() *VirtualList {
	return &VirtualList{
		items:    make([]ListItem, 0),
		filtered: make([]int, 0),
		selectedStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true),
		normalStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		dimStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
	}
}

// SetItems replaces all items and resets the view.
func (l *VirtualList) SetItems(items []ListItem) {
	l.items = items
	l.filtered = make([]int, len(items))
	for i := range items {
		l.filtered[i] = i
	}
	l.cursor = 0
	l.offset = 0
}

// SetSize sets the visible dimensions.
func (l *VirtualList) SetSize(width, height int) {
	l.width = width
	l.height = height
	// Ensure offset is valid
	if l.offset > l.cursor {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+l.height {
		l.offset = l.cursor - l.height + 1
	}
}

// Update handles keyboard input.
func (l *VirtualList) Update(msg tea.Msg) (*VirtualList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			l.moveUp(1)
		case "down", "j":
			l.moveDown(1)
		case "pgup", "ctrl+u":
			l.moveUp(l.height / 2)
		case "pgdown", "ctrl+d":
			l.moveDown(l.height / 2)
		case "home", "g":
			l.cursor = 0
			l.offset = 0
		case "end", "G":
			if len(l.filtered) > 0 {
				l.cursor = len(l.filtered) - 1
				l.offset = max(0, l.cursor-l.height+1)
			}
		}
	}
	return l, nil
}

func (l *VirtualList) moveUp(n int) {
	l.cursor = max(0, l.cursor-n)
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
}

func (l *VirtualList) moveDown(n int) {
	if len(l.filtered) == 0 {
		return
	}
	l.cursor = min(len(l.filtered)-1, l.cursor+n)
	if l.cursor >= l.offset+l.height {
		l.offset = l.cursor - l.height + 1
	}
}

// Filter filters items by query string (case-insensitive).
func (l *VirtualList) Filter(query string) {
	query = strings.ToLower(strings.TrimSpace(query))
	l.filtered = l.filtered[:0]

	for i, item := range l.items {
		if query == "" || strings.Contains(strings.ToLower(item.FilterValue()), query) {
			l.filtered = append(l.filtered, i)
		}
	}

	// Reset cursor and offset
	l.cursor = 0
	l.offset = 0
}

// ClearFilter shows all items.
func (l *VirtualList) ClearFilter() {
	l.filtered = make([]int, len(l.items))
	for i := range l.items {
		l.filtered[i] = i
	}
	l.cursor = 0
	l.offset = 0
}

// Selected returns the currently selected item, or nil if none.
func (l *VirtualList) Selected() ListItem {
	if len(l.filtered) == 0 || l.cursor >= len(l.filtered) {
		return nil
	}
	return l.items[l.filtered[l.cursor]]
}

// SelectedIndex returns the cursor index in filtered list.
func (l *VirtualList) SelectedIndex() int {
	return l.cursor
}

// Len returns the number of filtered items.
func (l *VirtualList) Len() int {
	return len(l.filtered)
}

// TotalLen returns the total number of items.
func (l *VirtualList) TotalLen() int {
	return len(l.items)
}

// View renders the visible portion of the list.
func (l *VirtualList) View() string {
	if len(l.filtered) == 0 {
		return l.dimStyle.Render("No items")
	}

	var b strings.Builder
	end := min(l.offset+l.height, len(l.filtered))

	for i := l.offset; i < end; i++ {
		idx := l.filtered[i]
		item := l.items[idx]
		selected := i == l.cursor

		line := l.renderItem(item, selected)
		b.WriteString(line)
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (l *VirtualList) renderItem(item ListItem, selected bool) string {
	title := item.Title()
	desc := item.Description()

	// Truncate to fit width
	maxWidth := l.width - 4 // Account for cursor and padding
	if len(title) > maxWidth {
		title = title[:maxWidth-3] + "..."
	}

	var line string
	if selected {
		cursor := "▸ "
		if desc != "" {
			line = cursor + l.selectedStyle.Render(title) + " " + l.dimStyle.Render(desc)
		} else {
			line = cursor + l.selectedStyle.Render(title)
		}
	} else {
		cursor := "  "
		if desc != "" {
			line = cursor + l.normalStyle.Render(title) + " " + l.dimStyle.Render(desc)
		} else {
			line = cursor + l.normalStyle.Render(title)
		}
	}

	return line
}

// ScrollInfo returns scroll position info (e.g., "1-20 of 100").
func (l *VirtualList) ScrollInfo() string {
	if len(l.filtered) == 0 {
		return "0 items"
	}
	start := l.offset + 1
	end := min(l.offset+l.height, len(l.filtered))
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).
		Render(strings.Join([]string{
			itoa(start), "-", itoa(end), " of ", itoa(len(l.filtered)),
		}, ""))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
