// Package xc provides a client for the Xtream Codes API.
// It handles JSON type inconsistencies common across providers.
package xc

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexibleID handles IDs that may be returned as either int or string.
// Many XC providers return IDs inconsistently across endpoints.
type FlexibleID struct {
	intVal    int
	stringVal string
	isInt     bool
}

// UnmarshalJSON implements json.Unmarshaler for FlexibleID.
// Handles numeric (123), float (8.1 -> truncated to 8), and string ("123") JSON values.
func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	// Try integer first
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		f.intVal = i
		f.isInt = true
		return nil
	}

	// Try float (truncate to int) - some APIs return floats for IDs
	var fl float64
	if err := json.Unmarshal(data, &fl); err == nil {
		f.intVal = int(fl)
		f.isInt = true
		return nil
	}

	// Try string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		f.stringVal = s
		f.isInt = false
		return nil
	}

	return fmt.Errorf("FlexibleID: cannot unmarshal %s", string(data))
}

// MarshalJSON implements json.Marshaler for FlexibleID.
func (f FlexibleID) MarshalJSON() ([]byte, error) {
	if f.isInt {
		return json.Marshal(f.intVal)
	}
	return json.Marshal(f.stringVal)
}

// String returns the ID as a string regardless of original type.
func (f FlexibleID) String() string {
	if f.isInt {
		return strconv.Itoa(f.intVal)
	}
	return f.stringVal
}

// Int returns the ID as an integer. If stored as string, attempts conversion.
func (f FlexibleID) Int() (int, error) {
	if f.isInt {
		return f.intVal, nil
	}
	return strconv.Atoi(f.stringVal)
}

// IsZero returns true if the ID is unset or zero.
func (f FlexibleID) IsZero() bool {
	if f.isInt {
		return f.intVal == 0
	}
	return f.stringVal == ""
}

// NewFlexibleID creates a FlexibleID from an integer.
func NewFlexibleID(id int) FlexibleID {
	return FlexibleID{intVal: id, isInt: true}
}

// NewFlexibleIDFromString creates a FlexibleID from a string.
func NewFlexibleIDFromString(id string) FlexibleID {
	return FlexibleID{stringVal: id, isInt: false}
}

// FlexibleFloat handles floats that may be returned as either float or string.
// XC providers often return ratings as strings like "4.5" instead of 4.5.
type FlexibleFloat float64

// UnmarshalJSON implements json.Unmarshaler for FlexibleFloat.
func (f *FlexibleFloat) UnmarshalJSON(data []byte) error {
	// Try float first
	var fl float64
	if err := json.Unmarshal(data, &fl); err == nil {
		*f = FlexibleFloat(fl)
		return nil
	}

	// Try string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*f = 0
			return nil
		}
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			*f = 0
			return nil // Don't fail on unparseable strings
		}
		*f = FlexibleFloat(val)
		return nil
	}

	return fmt.Errorf("FlexibleFloat: cannot unmarshal %s", string(data))
}

// Float64 returns the value as float64.
func (f FlexibleFloat) Float64() float64 {
	return float64(f)
}

// UserInfo represents authenticated user account information.
type UserInfo struct {
	Username       string     `json:"username"`
	Password       string     `json:"password"`
	Status         string     `json:"status"`
	ActiveCons     FlexibleID `json:"active_cons"`
	MaxConnections FlexibleID `json:"max_connections"`
	ExpDate        FlexibleID `json:"exp_date"` // Unix timestamp or "Unlimited"
	CreatedAt      FlexibleID `json:"created_at"`
	IsTrial        FlexibleID `json:"is_trial"`
}

// ServerInfo contains XC server metadata.
type ServerInfo struct {
	URL          string     `json:"url"`
	Port         FlexibleID `json:"port"`
	HTTPSPort    FlexibleID `json:"https_port"`
	Protocol     string     `json:"server_protocol"`
	RTMPPort     FlexibleID `json:"rtmp_port"`
	Timezone     string     `json:"timezone"`
	TimestampNow FlexibleID `json:"timestamp_now"`
	TimeNow      string     `json:"time_now"`
}

// AuthResponse is returned from the authentication endpoint.
type AuthResponse struct {
	UserInfo   UserInfo   `json:"user_info"`
	ServerInfo ServerInfo `json:"server_info"`
}

// Category represents a content category (live, VOD, or series).
type Category struct {
	ID       FlexibleID `json:"category_id"`
	Name     string     `json:"category_name"`
	ParentID FlexibleID `json:"parent_id"`
}

// LiveStream represents a live TV channel.
type LiveStream struct {
	ID           FlexibleID `json:"stream_id"`
	Num          FlexibleID `json:"num"`
	Name         string     `json:"name"`
	Icon         string     `json:"stream_icon"`
	CategoryID   FlexibleID `json:"category_id"`
	EPGChannelID string     `json:"epg_channel_id"`
	Added        FlexibleID `json:"added"`
	IsAdult      FlexibleID `json:"is_adult"`
	CustomSid    string     `json:"custom_sid"`
	TVArchive    FlexibleID `json:"tv_archive"`
}

// VODStream represents a video-on-demand item (movie).
type VODStream struct {
	ID           FlexibleID `json:"stream_id"`
	Num          FlexibleID `json:"num"`
	Name         string     `json:"name"`
	Icon         string     `json:"stream_icon"`
	CategoryID   FlexibleID `json:"category_id"`
	Container    string     `json:"container_extension"`
	Added        FlexibleID `json:"added"`
	Rating       string     `json:"rating"`
	Rating5Based FlexibleFloat `json:"rating_5based"`
	DirectSource string     `json:"direct_source"`
}

// Series represents a TV series.
type Series struct {
	ID           FlexibleID `json:"series_id"`
	Num          FlexibleID `json:"num"`
	Name         string     `json:"name"`
	Cover        string     `json:"cover"`
	Plot         string     `json:"plot"`
	Cast         string     `json:"cast"`
	Director     string     `json:"director"`
	Genre        string     `json:"genre"`
	ReleaseDate  string     `json:"releaseDate"`
	Rating       string     `json:"rating"`
	Rating5Based FlexibleFloat `json:"rating_5based"`
	CategoryID   FlexibleID `json:"category_id"`
}

// SeriesInfo contains detailed series information with episodes.
type SeriesInfo struct {
	Seasons  []SeasonInfo         `json:"seasons"`
	Episodes map[string][]Episode `json:"episodes"` // Keyed by season number
	Info     SeriesDetails        `json:"info"`
}

// SeasonInfo represents a season within a series.
type SeasonInfo struct {
	AirDate      string     `json:"air_date"`
	EpisodeCount FlexibleID `json:"episode_count"`
	ID           FlexibleID `json:"id"`
	Name         string     `json:"name"`
	Overview     string     `json:"overview"`
	SeasonNumber FlexibleID `json:"season_number"`
	Cover        string     `json:"cover"`
}

// Episode represents a single episode.
type Episode struct {
	ID           FlexibleID  `json:"id"`
	EpisodeNum   int         `json:"episode_num"`
	Title        string      `json:"title"`
	ContainerExt string      `json:"container_extension"`
	Info         EpisodeInfo `json:"info"`
	Added        FlexibleID  `json:"added"`
	Season       int         `json:"season"`
	DirectSource string      `json:"direct_source"`
}

// EpisodeInfo contains metadata about an episode.
type EpisodeInfo struct {
	Plot       string        `json:"plot"`
	Duration   string        `json:"duration"`
	Rating     FlexibleFloat `json:"rating"`
	MovieImage string        `json:"movie_image"`
	Bitrate    FlexibleID    `json:"bitrate"`
}

// SeriesDetails contains extended series metadata.
type SeriesDetails struct {
	Name           string     `json:"name"`
	Cover          string     `json:"cover"`
	Plot           string     `json:"plot"`
	Cast           string     `json:"cast"`
	Director       string     `json:"director"`
	Genre          string     `json:"genre"`
	ReleaseDate    string     `json:"releaseDate"`
	LastModified   FlexibleID `json:"last_modified"`
	Rating         string     `json:"rating"`
	Rating5Based   FlexibleFloat `json:"rating_5based"`
	BackdropPath   []string   `json:"backdrop_path"`
	YoutubeTrailer string     `json:"youtube_trailer"`
	TMDbID         FlexibleID `json:"tmdb_id"`
	CategoryID     FlexibleID `json:"category_id"`
}

// EPGShort represents short EPG (Electronic Program Guide) data.
type EPGShort struct {
	ID          FlexibleID `json:"id"`
	EPGListings []EPGEntry `json:"epg_listings"`
}

// EPGEntry represents a single EPG program entry.
type EPGEntry struct {
	ID             FlexibleID `json:"id"`
	EPGId          FlexibleID `json:"epg_id"`
	Title          string     `json:"title"`
	Lang           string     `json:"lang"`
	Start          string     `json:"start"`
	End            string     `json:"end"`
	Description    string     `json:"description"`
	ChannelID      string     `json:"channel_id"`
	StartTimestamp FlexibleID `json:"start_timestamp"`
	StopTimestamp  FlexibleID `json:"stop_timestamp"`
}
