package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FriendlyError converts technical errors to user-friendly messages.
func FriendlyError(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()

	switch {
	case strings.Contains(msg, "connection refused"):
		return "Cannot connect to server. Check host and port."
	case strings.Contains(msg, "no such host"):
		return "Server not found. Check the hostname."
	case strings.Contains(msg, "account not active"):
		return "Account is inactive or expired."
	case strings.Contains(msg, "authentication failed"):
		return "Wrong username or password."
	case strings.Contains(msg, "invalid credentials"):
		return "Wrong username or password."
	case strings.Contains(msg, "timeout"):
		return "Connection timed out. Server may be slow or offline."
	case strings.Contains(msg, "context canceled"):
		return "Request was cancelled."
	case strings.Contains(msg, "EOF"):
		return "Server closed connection unexpectedly."
	case strings.Contains(msg, "invalid client configuration"):
		return "Invalid server configuration."
	case strings.Contains(msg, "no player available"):
		return "No video player found. Install mpv or VLC."
	default:
		// Truncate long messages
		if len(msg) > 80 {
			return msg[:77] + "..."
		}
		return msg
	}
}

// ErrorModal renders an error as a centered modal.
type ErrorModal struct {
	message string
	visible bool
}

// SetError sets and shows an error message.
func (e *ErrorModal) SetError(msg string) {
	e.message = msg
	e.visible = msg != ""
}

// Clear hides the error modal.
func (e *ErrorModal) Clear() {
	e.message = ""
	e.visible = false
}

// IsVisible returns whether the modal is showing.
func (e *ErrorModal) IsVisible() bool {
	return e.visible
}

// View renders the error modal.
func (e *ErrorModal) View(width, height int) string {
	if !e.visible || e.message == "" {
		return ""
	}

	maxWidth := 50
	if width < 60 {
		maxWidth = width - 10
	}
	if maxWidth < 20 {
		maxWidth = 20
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(maxWidth)

	title := ErrorStyle.Render("⚠ Error")
	content := title + "\n\n" + wordWrap(e.message, maxWidth-6) + "\n\n" +
		HelpStyle.Render("[Enter] Dismiss")

	box := boxStyle.Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// wordWrap wraps text to a maximum width.
func wordWrap(text string, width int) string {
	if width <= 0 || len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		if i > 0 {
			if lineLen+1+len(word) > width {
				result.WriteString("\n")
				lineLen = 0
			} else {
				result.WriteString(" ")
				lineLen++
			}
		}
		result.WriteString(word)
		lineLen += len(word)
	}

	return result.String()
}
