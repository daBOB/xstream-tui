package style

import "github.com/charmbracelet/lipgloss"

// Color palette for consistent theming.
var (
	ColorPrimary   = lipgloss.Color("86")  // Cyan - brand color
	ColorSecondary = lipgloss.Color("205") // Pink - accent
	ColorSuccess   = lipgloss.Color("42")  // Green
	ColorError     = lipgloss.Color("196") // Red
	ColorWarning   = lipgloss.Color("214") // Orange
	ColorMuted     = lipgloss.Color("241") // Gray
	ColorHighlight = lipgloss.Color("62")  // Purple - selection bg
	ColorText      = lipgloss.Color("252") // Light gray - text
	ColorTextDim   = lipgloss.Color("244") // Dimmer text
)

// Base styles for layout and structure.
var (
	// AppStyle wraps the entire application with padding.
	AppStyle = lipgloss.NewStyle().Padding(1, 2)

	// TitleStyle for screen headers.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	// SubtitleStyle for secondary headers.
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			MarginBottom(1)
)

// Navigation and selection styles.
var (
	// SelectedStyle highlights the currently selected item.
	SelectedStyle = lipgloss.NewStyle().
			Background(ColorHighlight).
			Foreground(lipgloss.Color("230")).
			Bold(true)

	// CursorStyle for the selection indicator.
	CursorStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

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
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1)

	// ContentStyle for the main panel (70% width).
	ContentStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// BoxStyle for bordered containers.
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
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
			Border(lipgloss.DoubleBorder()).
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
			Background(lipgloss.Color("236")).
			Foreground(ColorText).
			Padding(0, 1)

	// StatusBarLabelStyle for labels in the status bar.
	StatusBarLabelStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)
)

// Spinner frames for loading animation.
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
