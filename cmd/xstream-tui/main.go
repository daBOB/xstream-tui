// Package main is the entry point for the xstream-tui application.
// It initializes the Bubble Tea program and runs the TUI.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
)

func main() {
	// Create new Bubble Tea program with alternate screen buffer
	p := tea.NewProgram(tui.NewApp(), tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
