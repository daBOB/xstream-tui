// Package xc — stream playback URL construction.
//
// These methods build provider playback URLs from the client's base URL and
// credentials. They are pure string builders: no HTTP, no retry, no context.
//
// # Security Note
//
// The XC API embeds credentials directly in playback URL paths and query
// parameters. This is an API limitation. Never log full URLs from this file.
package xc

import (
	"fmt"
	"time"
)

// LiveStreamURL returns the playback URL for a live stream.
func (c *Client) LiveStreamURL(streamID string) string {
	return fmt.Sprintf("%s/live/%s/%s/%s.ts", c.baseURL, c.username, c.password, streamID)
}

// LiveStreamURLWithFormat returns the playback URL with custom extension.
func (c *Client) LiveStreamURLWithFormat(streamID, ext string) string {
	return fmt.Sprintf("%s/live/%s/%s/%s.%s", c.baseURL, c.username, c.password, streamID, ext)
}

// VODStreamURL returns the playback URL for a VOD item.
func (c *Client) VODStreamURL(streamID, container string) string {
	return fmt.Sprintf("%s/movie/%s/%s/%s.%s", c.baseURL, c.username, c.password, streamID, container)
}

// SeriesEpisodeURL returns the playback URL for a series episode.
func (c *Client) SeriesEpisodeURL(episodeID, container string) string {
	return fmt.Sprintf("%s/series/%s/%s/%s.%s", c.baseURL, c.username, c.password, episodeID, container)
}

// TimeShiftURL returns the playback URL for time-shifted content.
func (c *Client) TimeShiftURL(streamID string, start time.Time, duration time.Duration) string {
	return fmt.Sprintf("%s/timeshift/%s/%s/%d/%d/%s.ts",
		c.baseURL, c.username, c.password,
		int(duration.Minutes()), start.Unix(), streamID)
}

// CatchUpURL returns the playback URL for catch-up content.
func (c *Client) CatchUpURL(streamID string, start, end time.Time) string {
	return fmt.Sprintf("%s/streaming/timeshift.php?username=%s&password=%s&stream=%s&start=%s&end=%s",
		c.baseURL, c.username, c.password, streamID,
		start.Format("2006-01-02:15-04"), end.Format("2006-01-02:15-04"))
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}
