// Package tui implements the terminal user interface using Bubble Tea.
// It follows The Elm Architecture (TEA) with Model-View-Update pattern.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// App is the main application model implementing tea.Model interface.
// It manages application state and coordinates between screens.
type App struct {
	width  int
	height int
}

// NewApp creates a new application instance with default state.
func NewApp() App {
	return App{}
}

// Init initializes the application. Called once when program starts.
func (a App) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates application state.
// Returns updated model and optional command to run.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Quit on q or ctrl+c
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		// Track terminal dimensions for responsive layout
		a.width = msg.Width
		a.height = msg.Height
	}

	return a, nil
}

// View renders the current application state as a string.
func (a App) View() string {
	// Placeholder view - will be replaced with actual screens
	style := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center)

	content := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Render("xstream-tui") +
		"\n\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render("Press 'q' to quit")

	return style.Render(content)
}
