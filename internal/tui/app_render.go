package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current application state.
func (a *App) View() string {
	// Loading overlay
	if a.loading {
		return a.renderLoading()
	}

	// Error overlay (if error msg exists, show it over current screen)
	content := a.renderScreen()
	if a.errorMsg != "" {
		content = a.renderWithError(content)
	} else if a.screen != LoginScreen && a.userInfo.Username != "" {
		content = a.renderWithStatusBar(content)
	}

	// Download queue overlay
	if a.downloadQueue.Visible() {
		content = a.renderWithDownloadQueue(content)
	}

	return content
}

// renderWithDownloadQueue renders the download queue as full-screen view.
func (a *App) renderWithDownloadQueue(_ string) string {
	a.downloadQueue.SetSize(a.width, a.height)
	return a.downloadQueue.View()
}

// renderWithStatusBar renders content with a status bar at the bottom.
func (a *App) renderWithStatusBar(content string) string {
	statusBar := a.renderStatusBar()

	contentHeight := a.height - 1
	if contentHeight < 1 {
		contentHeight = 1
	}
	content = lipgloss.NewStyle().Height(contentHeight).Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

// renderScreen renders the current screen view.
func (a *App) renderScreen() string {
	switch a.screen {
	case LoginScreen:
		if s, ok := a.login.(loginScreen); ok {
			return s.View()
		}
	case ContentTypeScreen:
		if s, ok := a.contentType.(contentTypeScreen); ok {
			return s.View()
		}
	case CategoriesScreen:
		if s, ok := a.categories.(categoriesScreen); ok {
			return s.View()
		}
	case StreamsScreen:
		if s, ok := a.streams.(streamsScreen); ok {
			return s.View()
		}
	case SeasonsScreen:
		if s, ok := a.seasons.(seasonsScreen); ok {
			return s.View()
		}
	case EpisodesScreen:
		if s, ok := a.episodes.(episodesScreen); ok {
			return s.View()
		}
	case SeriesBrowserScreen:
		if s, ok := a.seriesBrowser.(seriesBrowserScreen); ok {
			return s.View()
		}
	case GlobalSearchScreen:
		if s, ok := a.globalSearch.(globalSearchScreen); ok {
			return s.View()
		}
	}

	return a.renderPlaceholder()
}

// renderPlaceholder renders a fallback placeholder view.
func (a *App) renderPlaceholder() string {
	style := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center)

	content := TitleStyle.Render("xstream-tui") + "\n\n" +
		ItemDimStyle.Render("Press 'q' to quit")

	return style.Render(content)
}

// renderLoading renders the loading spinner view.
func (a *App) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center)

	spinner := SpinnerFrames[a.spinnerFrame]
	msg := a.loadingMsg
	if msg == "" {
		msg = "Loading..."
	}

	content := LoadingStyle.Render(spinner + " " + msg)
	return style.Render(content)
}

// renderWithError renders content with an error bar at the bottom.
func (a *App) renderWithError(content string) string {
	errorBar := ErrorStyle.Render("⚠ " + a.errorMsg + "  [Esc] Dismiss")
	errorBar = lipgloss.NewStyle().
		Width(a.width).
		Background(lipgloss.Color("52")).
		Padding(0, 2).
		Render(errorBar)

	contentHeight := a.height - 1
	if contentHeight < 1 {
		contentHeight = 1
	}
	content = lipgloss.NewStyle().Height(contentHeight).Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, content, errorBar)
}

// renderStatusBar renders the status bar at the bottom.
func (a *App) renderStatusBar() string {
	var leftParts []string

	if a.userInfo.Username != "" {
		leftParts = append(leftParts,
			StatusBarLabelStyle.Render("User: ")+a.userInfo.Username)
	}

	if a.screen == CategoriesScreen || a.screen == StreamsScreen {
		leftParts = append(leftParts,
			StatusBarLabelStyle.Render("Type: ")+a.currentType.String())
	}

	left := strings.Join(leftParts, "  │  ")

	var right string
	expDate := a.userInfo.ExpDate.String()
	if expDate != "" && expDate != "0" {
		right = StatusBarLabelStyle.Render("Expires: ") + expDate
	}

	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	availableWidth := a.width - 4

	padding := availableWidth - leftLen - rightLen
	if padding < 1 {
		padding = 1
	}

	content := left + strings.Repeat(" ", padding) + right

	return StatusBarStyle.Width(a.width).Render(content)
}
