package style

import "github.com/charmbracelet/lipgloss"

// Color palette for consistent theming.
var (
	// Neon Night Palette
	ColorBackground = lipgloss.Color("#1a1b26") // Deep Blue/Black
	ColorPrimary    = lipgloss.Color("#7aa2f7") // Neon Blue
	ColorSecondary  = lipgloss.Color("#bb9af7") // Neon Purple
	ColorSuccess    = lipgloss.Color("#9ece6a") // Neon Green
	ColorError      = lipgloss.Color("#f7768e") // Neon Red
	ColorWarning    = lipgloss.Color("#e0af68") // Neon Orange
	ColorMuted      = lipgloss.Color("#565f89") // Muted Blue/Gray
	ColorHighlight  = lipgloss.Color("#2f3549") // Highlight Background
	ColorText       = lipgloss.Color("#c0caf5") // Main Text (White-ish)
	ColorTextDim    = lipgloss.Color("#9aa5ce") // Dim Text
	ColorBorder     = lipgloss.Color("#414868") // Border Color
)

// Base styles for layout and structure.
var (
	// AppStyle wraps the entire application with padding.
	AppStyle = lipgloss.NewStyle().Padding(1, 2)

	// TitleStyle for screen headers.
	// TitleStyle for screen headers.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorHighlight).
			Padding(0, 1). // Add some breathing room
			MarginBottom(1)

	// SubtitleStyle for secondary headers.
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			MarginBottom(1)
)

// Navigation and selection styles.
var (
	// SelectedStyle highlights the currently selected item.
	// SelectedStyle highlights the currently selected item.
	SelectedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), false, false, false, true). // Left border indicator
			BorderForeground(ColorSecondary).
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1)

	// CursorStyle for the selection indicator.
	// CursorStyle for the selection indicator.
	CursorStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true).
			MarginRight(1)

	// ItemStyle for normal list items.
	ItemStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// ItemDimStyle for secondary info on items.
	ItemDimStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)

// Layout styles for screen structure.
var (
	// SidebarStyle for the left panel (30% width).
	SidebarStyle = lipgloss.NewStyle().
			BorderRight(true).
			BorderStyle(lipgloss.RoundedBorder()). // Rounded border
			BorderForeground(ColorBorder).
			Padding(0, 1)

	// ContentStyle for the main panel (70% width).
	ContentStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// BoxStyle for bordered containers.
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)
)

// Status and feedback styles.
var (
	// ErrorStyle for error messages.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// SuccessStyle for success messages.
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// WarningStyle for warning messages.
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	// HelpStyle for keyboard hints.
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	// LoadingStyle for spinner/loading text.
	LoadingStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Italic(true)
)

// Input styles for forms.
var (
	// InputLabelStyle for form labels.
	InputLabelStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Width(12)

	// InputFocusedStyle for focused input fields.
	InputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	// InputBlurredStyle for unfocused input fields.
	InputBlurredStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorMuted).
				Padding(0, 1)
)

// Content type styles for Live/VOD/Series tabs.
var (
	// TabActiveStyle for the selected content type.
	TabActiveStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 2)

	// TabInactiveStyle for unselected content types.
	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Padding(0, 2)
)

// Modal styles for overlays.
var (
	// ModalStyle for modal dialog boxes.
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSecondary).
			Padding(1, 3).
			Align(lipgloss.Center)

	// ModalTitleStyle for modal headers.
	ModalTitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true).
			MarginBottom(1)
)

// Status bar styles.
var (
	// StatusBarStyle for the bottom status bar.
	StatusBarStyle = lipgloss.NewStyle().
			Background(ColorHighlight).
			Foreground(ColorText).
			Padding(0, 1)

	// StatusBarLabelStyle for labels in the status bar.
	StatusBarLabelStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary). // Use pink for status labels
			Bold(true)
)

	// Spinner frames for loading animation - dots are good, but let's make them consistent
	var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
