package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/tui/style"
)

// HelpModel manages the help overlay state.
type HelpModel struct {
	visible bool
}

// Toggle toggles help visibility.
func (h *HelpModel) Toggle() {
	h.visible = !h.visible
}

// Show makes help visible.
func (h *HelpModel) Show() {
	h.visible = true
}

// Hide makes help invisible.
func (h *HelpModel) Hide() {
	h.visible = false
}

// IsVisible returns whether help is showing.
func (h *HelpModel) IsVisible() bool {
	return h.visible
}

// View renders the help overlay.
func (h *HelpModel) View(width, height int) string {
	if !h.visible {
		return ""
	}

	content := `Keyboard Shortcuts
══════════════════

Navigation
  ↑/k      Move up
  ↓/j      Move down
  PgUp     Page up
  PgDn     Page down
  Enter    Select
  Esc      Back
  q        Quit

Search
  /        Start search
  Esc      Cancel search

Player (mpv)
  Space    Toggle pause
  ←        Seek back 10s
  →        Seek forward 10s
  m        Mute

Press ? to close`

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorHighlight).
		Padding(1, 2).
		Width(32)

	box := boxStyle.Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
