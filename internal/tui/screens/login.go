// Package screens provides screen-level models for the TUI.
package screens

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/style"
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

	// Pre-populate from environment variables
	m.loadFromEnv()

	return m
}

// loadFromEnv pre-populates fields from environment variables.
func (m *LoginModel) loadFromEnv() {
	if host := os.Getenv("XSTREAM_HOST"); host != "" {
		// Strip protocol prefix if present
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")
		m.inputs[inputHost].SetValue(host)
	}
	if port := os.Getenv("XSTREAM_PORT"); port != "" {
		m.inputs[inputPort].SetValue(port)
	}
	if user := os.Getenv("XSTREAM_USERNAME"); user != "" {
		m.inputs[inputUser].SetValue(user)
	}
	if pass := os.Getenv("XSTREAM_PASSWORD"); pass != "" {
		m.inputs[inputPass].SetValue(pass)
	}
}

// Focus sets focus to the first input.
func (m *LoginModel) Focus() tea.Cmd {
	m.focused = inputHost
	return m.inputs[inputHost].Focus()
}

// HasCompleteCredentials returns true if all required env vars are set.
func (m *LoginModel) HasCompleteCredentials() bool {
	host := strings.TrimSpace(m.inputs[inputHost].Value())
	user := strings.TrimSpace(m.inputs[inputUser].Value())
	pass := strings.TrimSpace(m.inputs[inputPass].Value())
	return host != "" && user != "" && pass != ""
}

// AutoLogin attempts login if credentials are complete from env vars.
func (m *LoginModel) AutoLogin() tea.Cmd {
	if m.HasCompleteCredentials() {
		return m.submit()
	}
	return nil
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
		// Check for debug mode via environment variable
		var opts []xc.ClientOption
		if debugEnv := os.Getenv("XSTREAM_DEBUG"); debugEnv != "" && debugEnv != "0" && debugEnv != "false" {
			opts = append(opts, xc.WithDebug(true))
		}

		client, err := xc.NewClient(url, user, pass, opts...)
		if err != nil {
			return tui.AuthErrorMsg{Err: err}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		auth, err := client.Authenticate(ctx)
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

	// Banner
	banner := `
  ___  ___ _____ ______  _____  ___  ___  ___
  \  \/  //  ___|| ___ \|  _  |/ _ \ |  \/  |
   >    < \ --. | |_/ /| | | / /_\ \| .  . |
  /  /\  \ --. \|    / | | | |  _  || |\/| |
 /  /  \ \/\__/ /| |\ \ \ \_/ /| | | || |  | |
/__/    \_\____/ \_| \_| \___/ \_| |_/\_|  |_/
`
	logo := lipgloss.NewStyle().
		Foreground(style.ColorPrimary).
		Bold(true).
		Render(banner)

	b.WriteString(logo)
	b.WriteString("\n\n")

	// Subtitle
	b.WriteString(style.SubtitleStyle.Align(lipgloss.Center).Width(60).Render("Stream your favorite content in your terminal"))
	b.WriteString("\n\n")

	// Form fields
	labels := []string{"Server", "Port", "Username", "Password"}
	
	// Calculate max label width for alignment
	maxLabelWidth := 0
	for _, l := range labels {
		if len(l) > maxLabelWidth {
			maxLabelWidth = len(l)
		}
	}

	for i, input := range m.inputs {
		labelStyle := style.InputLabelStyle.Width(maxLabelWidth + 2)
		label := labelStyle.Render(labels[i])

		var inputStyle lipgloss.Style
		if i == m.focused {
			inputStyle = style.InputFocusedStyle
		} else {
			inputStyle = style.InputBlurredStyle
		}

		field := inputStyle.Render(input.View())
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, label, field))
		b.WriteString("\n\n") // More spacing between fields
	}

	// Error message
	if m.err != "" {
		b.WriteString(style.ErrorStyle.Render("✗ " + m.err))
		b.WriteString("\n")
	}

	// Help text
	help := style.HelpStyle.Render("[Tab] Next  [Shift+Tab] Prev  [Enter] Submit  [Esc] Quit")
	b.WriteString(help)

	// Center the form
	content := b.String()
	box := style.BoxStyle.
		BorderForeground(style.ColorBorder).
		Padding(2, 4).
		Render(content)
		
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
