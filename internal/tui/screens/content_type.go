package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/tui"
)

// ContentTypeModel handles content type selection (Live/VOD/Series).
type ContentTypeModel struct {
	options  []tui.ContentType
	cursor   int
	width    int
	height   int
	userInfo string // Display user info from login
}

// NewContentTypeModel creates a new content type selector.
func NewContentTypeModel() *ContentTypeModel {
	return &ContentTypeModel{
		options: []tui.ContentType{
			tui.LiveContent,
			tui.VODContent,
			tui.SeriesContent,
		},
	}
}

// SetSize sets the screen dimensions.
func (m *ContentTypeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetUserInfo sets the user info to display.
func (m *ContentTypeModel) SetUserInfo(info string) {
	m.userInfo = info
}

// Update handles input events.
func (m *ContentTypeModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j", "right", "l":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter", " ":
			return func() tea.Msg {
				return tui.ContentTypeSelectedMsg{Type: m.options[m.cursor]}
			}
		case "1":
			m.cursor = 0
			return func() tea.Msg {
				return tui.ContentTypeSelectedMsg{Type: tui.LiveContent}
			}
		case "2":
			m.cursor = 1
			return func() tea.Msg {
				return tui.ContentTypeSelectedMsg{Type: tui.VODContent}
			}
		case "3":
			m.cursor = 2
			return func() tea.Msg {
				return tui.ContentTypeSelectedMsg{Type: tui.SeriesContent}
			}
		}
	}
	return nil
}

// Selected returns the currently selected content type.
func (m *ContentTypeModel) Selected() tui.ContentType {
	return m.options[m.cursor]
}

// View renders the content type selector.
func (m *ContentTypeModel) View() string {
	var b strings.Builder

	// Title
	title := tui.TitleStyle.Render("📺 Select Content Type")
	b.WriteString(title)
	b.WriteString("\n\n")

	// User info if available
	if m.userInfo != "" {
		info := tui.ItemDimStyle.Render("Logged in: " + m.userInfo)
		b.WriteString(info)
		b.WriteString("\n\n")
	}

	// Content type options as large buttons
	icons := []string{"📡", "🎬", "📺"}
	descs := []string{
		"Live TV channels",
		"Movies & VOD",
		"TV Series",
	}

	for i, opt := range m.options {
		var style lipgloss.Style
		if i == m.cursor {
			style = lipgloss.NewStyle().
				Background(tui.ColorHighlight).
				Foreground(lipgloss.Color("230")).
				Bold(true).
				Padding(1, 4).
				Margin(0, 0, 1, 0).
				Width(30)
		} else {
			style = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(tui.ColorMuted).
				Padding(1, 4).
				Margin(0, 0, 1, 0).
				Width(30)
		}

		content := icons[i] + "  " + opt.String() + "\n" +
			tui.ItemDimStyle.Render(descs[i])

		b.WriteString(style.Render(content))
		b.WriteString("\n")
	}

	// Help text
	help := tui.HelpStyle.Render("\n[↑↓] Navigate  [Enter] Select  [1-3] Quick select  [Esc] Back")
	b.WriteString(help)

	// Center the content
	box := tui.BoxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
