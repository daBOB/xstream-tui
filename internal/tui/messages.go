package tui

import (
	"github.com/altmueller/xstream-tui/internal/xc"
)

// Screen represents the current application screen.
type Screen int

const (
	LoginScreen Screen = iota
	ContentTypeScreen
	CategoriesScreen
	StreamsScreen
	SeasonsScreen
	EpisodesScreen
	SeriesBrowserScreen
	PlayerScreen
)

// ContentType represents the type of content (Live, VOD, Series).
type ContentType int

const (
	LiveContent ContentType = iota
	VODContent
	SeriesContent
)

func (c ContentType) String() string {
	switch c {
	case LiveContent:
		return "Live TV"
	case VODContent:
		return "Movies"
	case SeriesContent:
		return "Series"
	default:
		return "Unknown"
	}
}

// Custom messages for inter-screen communication.

// AuthSuccessMsg indicates successful authentication.
type AuthSuccessMsg struct {
	Client   *xc.Client
	UserInfo xc.UserInfo
}

// AuthErrorMsg indicates authentication failure.
type AuthErrorMsg struct {
	Err error
}

func (e AuthErrorMsg) Error() string {
	return e.Err.Error()
}

// ContentTypeSelectedMsg indicates a content type was chosen.
type ContentTypeSelectedMsg struct {
	Type ContentType
}

// CategoriesLoadedMsg contains loaded categories.
type CategoriesLoadedMsg struct {
	Categories []xc.Category
}

// CategorySelectedMsg indicates a category was chosen.
type CategorySelectedMsg struct {
	Category xc.Category
}

// StreamsLoadedMsg contains loaded streams.
type StreamsLoadedMsg struct {
	LiveStreams []xc.LiveStream
	VODStreams  []xc.VODStream
	Series      []xc.Series
}

// StreamSelectedMsg indicates a stream was chosen for playback.
type StreamSelectedMsg struct {
	URL  string
	Name string
}

// ErrorMsg is a generic error message.
type ErrorMsg struct {
	Err error
}

func (e ErrorMsg) Error() string {
	return e.Err.Error()
}

// ClearErrorMsg clears the current error.
type ClearErrorMsg struct{}

// LoadingMsg indicates a loading state change.
type LoadingMsg struct {
	Loading bool
	Message string
}

// SpinnerTickMsg is sent to advance the spinner animation.
type SpinnerTickMsg struct{}

// NavigateBackMsg requests navigation to the previous screen.
type NavigateBackMsg struct{}

// PlayerStartedMsg indicates playback has started.
type PlayerStartedMsg struct {
	PlayerType string // "mpv" or "vlc"
}

// PlayerStoppedMsg indicates playback has ended.
type PlayerStoppedMsg struct {
	Err error // nil if ended normally
}

// SeriesSelectedMsg indicates a series was chosen.
type SeriesSelectedMsg struct {
	Series xc.Series
}

// SeriesInfoLoadedMsg contains loaded series info with seasons/episodes.
type SeriesInfoLoadedMsg struct {
	Info *xc.SeriesInfo
}

// SeasonSelectedMsg indicates a season was chosen.
type SeasonSelectedMsg struct {
	Season   xc.SeasonInfo
	Episodes []xc.Episode
}

// EpisodeSelectedMsg indicates an episode was chosen for playback.
type EpisodeSelectedMsg struct {
	Episode xc.Episode
}

// DownloadRequestMsg requests adding a download to the queue.
type DownloadRequestMsg struct {
	Name       string
	URL        string
	SeriesName string // Optional: for series episodes, creates subfolder
	SeasonName string // Optional: for series episodes, creates subfolder
}

// DownloadProgressMsg reports download progress updates.
type DownloadProgressMsg struct {
	ID         string
	Progress   float64
	Downloaded int64
	Size       int64
	Status     string
	Error      error
}

// DownloadQueueToggleMsg toggles the download queue visibility.
type DownloadQueueToggleMsg struct{}

// DownloadCancelMsg cancels a download by ID.
type DownloadCancelMsg struct {
	ID string
}

// DownloadRemoveMsg removes a download from the queue by ID.
type DownloadRemoveMsg struct {
	ID string
}

// BatchDownloadMsg requests adding multiple downloads to the queue.
type BatchDownloadMsg struct {
	Downloads []DownloadRequestMsg
}
