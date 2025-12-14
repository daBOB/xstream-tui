// Package main is the entry point for the xstream-tui application.
// It initializes the Bubble Tea program and runs the TUI.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/screens"
)

func main() {
	// Create and configure the application
	app := tui.NewApp()

	// Initialize screen models
	login := screens.NewLoginModel()
	contentType := screens.NewContentTypeModel()
	categories := screens.NewCategoriesModel()
	streams := screens.NewStreamsModel()

	// Inject screens into app
	app.SetLoginScreen(login)
	app.SetContentTypeScreen(contentType)
	app.SetCategoriesScreen(categories)
	app.SetStreamsScreen(streams)

	// Create Bubble Tea program with alternate screen buffer
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Cleanup player on exit
	app.StopPlayback()
}
