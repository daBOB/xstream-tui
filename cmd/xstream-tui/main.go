// Package main is the entry point for the xstream-tui application.
// It initializes the Bubble Tea program and runs the TUI.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"github.com/altmueller/xstream-tui/internal/download"
	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/tui/screens"
)

func main() {
	// Load .env file if present (ignore errors if not found)
	_ = godotenv.Load()

	// Create and configure the application
	app := tui.NewApp()

	// Initialize screen models
	login := screens.NewLoginModel()
	contentType := screens.NewContentTypeModel()
	categories := screens.NewCategoriesModel()
	streams := screens.NewStreamsModel()
	seasons := screens.NewSeasonsModel()
	episodes := screens.NewEpisodesModel()
	seriesBrowser := screens.NewSeriesBrowserModel()

	// Inject screens into app
	app.SetLoginScreen(login)
	app.SetContentTypeScreen(contentType)
	app.SetCategoriesScreen(categories)
	app.SetStreamsScreen(streams)
	app.SetSeasonsScreen(seasons)
	app.SetEpisodesScreen(episodes)
	app.SetSeriesBrowserScreen(seriesBrowser)

	// Initialize download manager
	downloadDir := getDownloadDir()
	downloader := download.NewManager(downloadDir)
	app.SetDownloader(downloader)

	// Create Bubble Tea program with alternate screen buffer
	p := tea.NewProgram(app, tea.WithAltScreen())

	// Set program reference for download progress callbacks
	app.SetProgram(p)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Cleanup player on exit
	app.StopPlayback()
}

// getDownloadDir returns the default download directory.
func getDownloadDir() string {
	// Check environment variable first
	if dir := os.Getenv("XSTREAM_DOWNLOAD_DIR"); dir != "" {
		return dir
	}

	// Use ~/Videos/xstream as default
	home, err := os.UserHomeDir()
	if err != nil {
		return "./downloads"
	}
	return filepath.Join(home, "Videos", "xstream")
}
