// Package screens provides screen-level models for the TUI.
package screens

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/xc"
)

const (
	inputHost = iota
	inputPort
	inputUser
	inputPass
)

// LoginModel handles user credential input.
type LoginModel struct {
	inputs  []textinput.Model
	focused int
	width   int
	height  int
	err     string
}

// NewLoginModel creates a new login screen model.
func NewLoginModel() *LoginModel {
	m := &LoginModel{
		inputs: make([]textinput.Model, 4),
	}

	// Host input
	m.inputs[inputHost] = textinput.New()
	m.inputs[inputHost].Placeholder = "example.com"
	m.inputs[inputHost].CharLimit = 100
	m.inputs[inputHost].Width = 40

	// Port input
	m.inputs[inputPort] = textinput.New()
	m.inputs[inputPort].Placeholder = "8080"
	m.inputs[inputPort].CharLimit = 5
	m.inputs[inputPort].Width = 10

	// Username input
	m.inputs[inputUser] = textinput.New()
	m.inputs[inputUser].Placeholder = "username"
	m.inputs[inputUser].CharLimit = 50
	m.inputs[inputUser].Width = 30

	// Password input
	m.inputs[inputPass] = textinput.New()
	m.inputs[inputPass].Placeholder = "password"
	m.inputs[inputPass].EchoMode = textinput.EchoPassword
	m.inputs[inputPass].EchoCharacter = '•'
	m.inputs[inputPass].CharLimit = 50
	m.inputs[inputPass].Width = 30

	return m
}

// Focus sets focus to the first input.
func (m *LoginModel) Focus() tea.Cmd {
	m.focused = inputHost
	return m.inputs[inputHost].Focus()
}

// SetSize sets the screen dimensions.
func (m *LoginModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetError sets an error message to display.
func (m *LoginModel) SetError(err string) {
	m.err = err
}

// ClearError clears the error message.
func (m *LoginModel) ClearError() {
	m.err = ""
}

// Update handles input events.
func (m *LoginModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear error on any keypress
		m.err = ""

		switch msg.String() {
		case "tab", "down":
			m.focusNext()
			return nil
		case "shift+tab", "up":
			m.focusPrev()
			return nil
		case "enter":
			if m.focused == inputPass {
				// Submit form
				return m.submit()
			}
			m.focusNext()
			return nil
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return cmd
}

func (m *LoginModel) focusNext() {
	m.inputs[m.focused].Blur()
	m.focused = (m.focused + 1) % len(m.inputs)
	m.inputs[m.focused].Focus()
}

func (m *LoginModel) focusPrev() {
	m.inputs[m.focused].Blur()
	m.focused = (m.focused - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focused].Focus()
}

func (m *LoginModel) submit() tea.Cmd {
	host := strings.TrimSpace(m.inputs[inputHost].Value())
	port := strings.TrimSpace(m.inputs[inputPort].Value())
	user := strings.TrimSpace(m.inputs[inputUser].Value())
	pass := strings.TrimSpace(m.inputs[inputPass].Value())

	// Validation
	if host == "" {
		m.err = "Server host is required"
		return nil
	}
	if user == "" {
		m.err = "Username is required"
		return nil
	}
	if pass == "" {
		m.err = "Password is required"
		return nil
	}

	// Clear password from input field for security
	m.inputs[inputPass].SetValue("")

	// Default port
	if port == "" {
		port = "8080"
	}

	// Build URL
	url := "http://" + host + ":" + port

	return func() tea.Msg {
		client, err := xc.NewClient(url, user, pass)
		if err != nil {
			return tui.AuthErrorMsg{Err: err}
		}

		auth, err := client.Authenticate(context.Background())
		if err != nil {
			return tui.AuthErrorMsg{Err: err}
		}

		return tui.AuthSuccessMsg{
			Client:   client,
			UserInfo: auth.UserInfo,
		}
	}
}

// View renders the login screen.
func (m *LoginModel) View() string {
	var b strings.Builder

	// Title
	title := tui.TitleStyle.Render("🔐 IPTV Login")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Form fields
	labels := []string{"Server:", "Port:", "Username:", "Password:"}
	for i, input := range m.inputs {
		label := tui.InputLabelStyle.Render(labels[i])

		var inputStyle lipgloss.Style
		if i == m.focused {
			inputStyle = tui.InputFocusedStyle
		} else {
			inputStyle = tui.InputBlurredStyle
		}

		field := inputStyle.Render(input.View())
		b.WriteString(label + " " + field + "\n")
	}

	// Error message
	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(tui.ErrorStyle.Render("✗ " + m.err))
	}

	// Help text
	help := tui.HelpStyle.Render("\n[Tab] Next field  [Enter] Submit  [Ctrl+C] Quit")
	b.WriteString(help)

	// Center the form
	box := tui.BoxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
