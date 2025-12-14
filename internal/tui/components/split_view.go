// Package components provides reusable UI components for the TUI.
package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SplitPane represents which pane is active.
type SplitPane int

const (
	LeftPane SplitPane = iota
	RightPane
)

// SplitView provides a side-by-side view of two components.
type SplitView struct {
	leftWidth  float64 // 0.0 to 1.0 proportion
	width      int
	height     int
	active     SplitPane
	leftTitle  string
	rightTitle string
}

// NewSplitView creates a new split view with specified left proportion.
func NewSplitView(leftProportion float64) *SplitView {
	if leftProportion < 0.2 {
		leftProportion = 0.2
	}
	if leftProportion > 0.8 {
		leftProportion = 0.8
	}

	return &SplitView{
		leftWidth: leftProportion,
		active:    LeftPane,
	}
}

// SetSize sets the total dimensions.
func (s *SplitView) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetTitles sets the pane titles.
func (s *SplitView) SetTitles(left, right string) {
	s.leftTitle = left
	s.rightTitle = right
}

// GetLeftDimensions returns the left pane dimensions.
func (s *SplitView) GetLeftDimensions() (width, height int) {
	dividerWidth := 1
	leftW := int(float64(s.width) * s.leftWidth)
	if leftW < 10 {
		leftW = 10
	}
	return leftW - dividerWidth, s.height - 4 // Account for title bar
}

// GetRightDimensions returns the right pane dimensions.
func (s *SplitView) GetRightDimensions() (width, height int) {
	dividerWidth := 1
	leftW := int(float64(s.width) * s.leftWidth)
	rightW := s.width - leftW - dividerWidth
	if rightW < 10 {
		rightW = 10
	}
	return rightW, s.height - 4 // Account for title bar
}

// Active returns which pane is active.
func (s *SplitView) Active() SplitPane {
	return s.active
}

// SetActive sets the active pane.
func (s *SplitView) SetActive(pane SplitPane) {
	s.active = pane
}

// ToggleActive switches the active pane.
func (s *SplitView) ToggleActive() {
	if s.active == LeftPane {
		s.active = RightPane
	} else {
		s.active = LeftPane
	}
}

// Update handles pane switching.
func (s *SplitView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "l":
			if s.active == LeftPane {
				s.active = RightPane
			}
		case "shift+tab", "h":
			if s.active == RightPane {
				s.active = LeftPane
			}
		}
	}
	return nil
}

// SplitViewStyles holds styling for split view.
type SplitViewStyles struct {
	ActiveBorder   lipgloss.Style
	InactiveBorder lipgloss.Style
	Title          lipgloss.Style
	ActiveTitle    lipgloss.Style
	Divider        lipgloss.Style
}

// DefaultSplitViewStyles returns default styling.
func DefaultSplitViewStyles() SplitViewStyles {
	return SplitViewStyles{
		ActiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")),
		InactiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1),
		ActiveTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Padding(0, 1),
		Divider: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")),
	}
}

// Render renders both panes with a divider.
func (s *SplitView) Render(leftContent, rightContent string, styles SplitViewStyles) string {
	leftW, leftH := s.GetLeftDimensions()
	rightW, rightH := s.GetRightDimensions()

	// Style for left pane
	leftStyle := styles.InactiveBorder
	leftTitleStyle := styles.Title
	if s.active == LeftPane {
		leftStyle = styles.ActiveBorder
		leftTitleStyle = styles.ActiveTitle
	}
	leftStyle = leftStyle.Width(leftW).Height(leftH)

	// Style for right pane
	rightStyle := styles.InactiveBorder
	rightTitleStyle := styles.Title
	if s.active == RightPane {
		rightStyle = styles.ActiveBorder
		rightTitleStyle = styles.ActiveTitle
	}
	rightStyle = rightStyle.Width(rightW).Height(rightH)

	// Build left pane
	leftTitle := leftTitleStyle.Render(s.leftTitle)
	leftPane := lipgloss.JoinVertical(lipgloss.Left,
		leftTitle,
		leftStyle.Render(leftContent),
	)

	// Build right pane
	rightTitle := rightTitleStyle.Render(s.rightTitle)
	rightPane := lipgloss.JoinVertical(lipgloss.Left,
		rightTitle,
		rightStyle.Render(rightContent),
	)

	// Divider
	divider := styles.Divider.Render(strings.Repeat("│\n", s.height-2))

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, divider, rightPane)
}

// RenderSimple renders without borders for a cleaner look.
func (s *SplitView) RenderSimple(leftContent, rightContent string) string {
	leftW, _ := s.GetLeftDimensions()
	rightW, _ := s.GetRightDimensions()

	activeIndicator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	inactiveIndicator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	dividerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("238")).
		Padding(0, 1)

	// Build titles with indicators
	var leftTitle, rightTitle string
	if s.active == LeftPane {
		leftTitle = activeIndicator.Render("▸ " + s.leftTitle)
		rightTitle = inactiveIndicator.Render("  " + s.rightTitle)
	} else {
		leftTitle = inactiveIndicator.Render("  " + s.leftTitle)
		rightTitle = activeIndicator.Render("▸ " + s.rightTitle)
	}

	// Apply width constraints
	leftPane := lipgloss.NewStyle().
		Width(leftW).
		Render(lipgloss.JoinVertical(lipgloss.Left, leftTitle, leftContent))

	rightPane := lipgloss.NewStyle().
		Width(rightW).
		Render(lipgloss.JoinVertical(lipgloss.Left, rightTitle, rightContent))

	divider := dividerStyle.Render("│")

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, divider, rightPane)
}
