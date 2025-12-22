package tui

import "time"

// Layout constants for consistent UI dimensions.
const (
	// Padding values
	PaddingSmall  = 1
	PaddingMedium = 2
	PaddingLarge  = 4

	// Header/footer reserved heights
	HeaderHeight       = 1
	FooterHeight       = 1
	StatusBarHeight    = 1
	SearchBarHeight    = 3
	ScreenPaddingTotal = 8 // Total vertical padding for screens (header + footer + chrome)

	// List layout
	ListHorizontalPadding = 4
	ListMinHeight         = 5

	// Spinner/animation timing
	SpinnerTickRate = 100 * time.Millisecond

	// Network timeouts
	DefaultTimeout       = 30 * time.Second
	GlobalSearchTimeout  = 60 * time.Second
	PlayerStartupTimeout = 30 * time.Second
)
