package xc

import (
	"context"
	"fmt"
)

// Authenticate validates credentials and returns user/server info.
// This is the primary method to verify credentials are valid.
func (c *Client) Authenticate(ctx context.Context) (*AuthResponse, error) {
	var resp AuthResponse
	if err := c.get(ctx, "", nil, &resp); err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	// Validate account status
	if resp.UserInfo.Status != "Active" {
		return nil, fmt.Errorf("account not active: status=%s", resp.UserInfo.Status)
	}

	return &resp, nil
}

// GetLiveCategories returns all live TV categories.
func (c *Client) GetLiveCategories(ctx context.Context) ([]Category, error) {
	var cats []Category
	if err := c.get(ctx, "get_live_categories", nil, &cats); err != nil {
		return nil, fmt.Errorf("get live categories: %w", err)
	}
	return cats, nil
}

// GetVODCategories returns all video-on-demand categories.
func (c *Client) GetVODCategories(ctx context.Context) ([]Category, error) {
	var cats []Category
	if err := c.get(ctx, "get_vod_categories", nil, &cats); err != nil {
		return nil, fmt.Errorf("get vod categories: %w", err)
	}
	return cats, nil
}

// GetSeriesCategories returns all series categories.
func (c *Client) GetSeriesCategories(ctx context.Context) ([]Category, error) {
	var cats []Category
	if err := c.get(ctx, "get_series_categories", nil, &cats); err != nil {
		return nil, fmt.Errorf("get series categories: %w", err)
	}
	return cats, nil
}

// GetLiveStreams returns live streams, optionally filtered by category.
// Pass empty categoryID to get all streams.
func (c *Client) GetLiveStreams(ctx context.Context, categoryID string) ([]LiveStream, error) {
	params := make(map[string]string)
	if categoryID != "" {
		params["category_id"] = categoryID
	}

	var streams []LiveStream
	if err := c.get(ctx, "get_live_streams", params, &streams); err != nil {
		return nil, fmt.Errorf("get live streams: %w", err)
	}
	return streams, nil
}

// GetVODStreams returns VOD items, optionally filtered by category.
// Pass empty categoryID to get all items.
func (c *Client) GetVODStreams(ctx context.Context, categoryID string) ([]VODStream, error) {
	params := make(map[string]string)
	if categoryID != "" {
		params["category_id"] = categoryID
	}

	var streams []VODStream
	if err := c.get(ctx, "get_vod_streams", params, &streams); err != nil {
		return nil, fmt.Errorf("get vod streams: %w", err)
	}
	return streams, nil
}

// GetSeries returns series list, optionally filtered by category.
// Pass empty categoryID to get all series.
func (c *Client) GetSeries(ctx context.Context, categoryID string) ([]Series, error) {
	params := make(map[string]string)
	if categoryID != "" {
		params["category_id"] = categoryID
	}

	var series []Series
	if err := c.get(ctx, "get_series", params, &series); err != nil {
		return nil, fmt.Errorf("get series: %w", err)
	}
	return series, nil
}

// GetSeriesInfo returns detailed information about a series including episodes.
func (c *Client) GetSeriesInfo(ctx context.Context, seriesID string) (*SeriesInfo, error) {
	params := map[string]string{
		"series_id": seriesID,
	}

	var info SeriesInfo
	if err := c.get(ctx, "get_series_info", params, &info); err != nil {
		return nil, fmt.Errorf("get series info: %w", err)
	}
	return &info, nil
}

// GetVODInfo returns detailed information about a VOD item.
func (c *Client) GetVODInfo(ctx context.Context, vodID string) (*VODStream, error) {
	params := map[string]string{
		"vod_id": vodID,
	}

	var info VODStream
	if err := c.get(ctx, "get_vod_info", params, &info); err != nil {
		return nil, fmt.Errorf("get vod info: %w", err)
	}
	return &info, nil
}

// GetShortEPG returns short EPG data for a stream (current and next).
func (c *Client) GetShortEPG(ctx context.Context, streamID string) (*EPGShort, error) {
	params := map[string]string{
		"stream_id": streamID,
	}

	var epg EPGShort
	if err := c.get(ctx, "get_short_epg", params, &epg); err != nil {
		return nil, fmt.Errorf("get short epg: %w", err)
	}
	return &epg, nil
}

// GetSimpleDataTable returns EPG listings for a stream within date range.
func (c *Client) GetSimpleDataTable(ctx context.Context, streamID string) ([]EPGEntry, error) {
	params := map[string]string{
		"stream_id": streamID,
	}

	var entries []EPGEntry
	if err := c.get(ctx, "get_simple_data_table", params, &entries); err != nil {
		return nil, fmt.Errorf("get simple data table: %w", err)
	}
	return entries, nil
}

// GetAllLiveStreams fetches all live streams (no category filter).
// Use with caution - may return 20,000+ items.
func (c *Client) GetAllLiveStreams(ctx context.Context) ([]LiveStream, error) {
	return c.GetLiveStreams(ctx, "")
}

// GetAllVODStreams fetches all VOD items (no category filter).
// Use with caution - may return thousands of items.
func (c *Client) GetAllVODStreams(ctx context.Context) ([]VODStream, error) {
	return c.GetVODStreams(ctx, "")
}

// GetAllSeries fetches all series (no category filter).
func (c *Client) GetAllSeries(ctx context.Context) ([]Series, error) {
	return c.GetSeries(ctx, "")
}
