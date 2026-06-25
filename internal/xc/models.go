// Package xc provides a client for the Xtream Codes API.
//
// This file holds the plain domain DTOs for XC API responses. The custom
// JSON adapter types they reference (FlexibleID, FlexibleFloat, etc.) live
// in flexible_types.go.
package xc

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
	ID           FlexibleID    `json:"stream_id"`
	Num          FlexibleID    `json:"num"`
	Name         string        `json:"name"`
	Icon         string        `json:"stream_icon"`
	CategoryID   FlexibleID    `json:"category_id"`
	Container    string        `json:"container_extension"`
	Added        FlexibleID    `json:"added"`
	Rating       string        `json:"rating"`
	Rating5Based FlexibleFloat `json:"rating_5based"`
	DirectSource string        `json:"direct_source"`
}

// Series represents a TV series.
type Series struct {
	ID           FlexibleID    `json:"series_id"`
	Num          FlexibleID    `json:"num"`
	Name         string        `json:"name"`
	Cover        string        `json:"cover"`
	Plot         string        `json:"plot"`
	Cast         string        `json:"cast"`
	Director     string        `json:"director"`
	Genre        string        `json:"genre"`
	ReleaseDate  string        `json:"releaseDate"`
	Rating       string        `json:"rating"`
	Rating5Based FlexibleFloat `json:"rating_5based"`
	CategoryID   FlexibleID    `json:"category_id"`
}

// SeriesInfo contains detailed series information with episodes.
type SeriesInfo struct {
	Seasons  FlexibleSeasons    `json:"seasons"`
	Episodes FlexibleEpisodes   `json:"episodes"` // Keyed by season number
	Info     FlexibleSeriesInfo `json:"info"`
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
	Name           string            `json:"name"`
	Cover          string            `json:"cover"`
	Plot           string            `json:"plot"`
	Cast           string            `json:"cast"`
	Director       string            `json:"director"`
	Genre          string            `json:"genre"`
	ReleaseDate    string            `json:"releaseDate"`
	LastModified   FlexibleID        `json:"last_modified"`
	Rating         string            `json:"rating"`
	Rating5Based   FlexibleFloat     `json:"rating_5based"`
	BackdropPath   FlexibleStringArr `json:"backdrop_path"`
	YoutubeTrailer string            `json:"youtube_trailer"`
	TMDbID         FlexibleID        `json:"tmdb_id"`
	CategoryID     FlexibleID        `json:"category_id"`
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
