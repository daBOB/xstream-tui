package xc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mustNewClient creates a client for testing, panicking on error.
func mustNewClient(host, username, password string, opts ...ClientOption) *Client {
	c, err := NewClient(host, username, password, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

func TestNewClient(t *testing.T) {
	c, err := NewClient("http://example.com", "user", "pass")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	if c.baseURL != "http://example.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "http://example.com")
	}
	if c.username != "user" {
		t.Errorf("username = %q, want %q", c.username, "user")
	}
	if c.password != "pass" {
		t.Errorf("password = %q, want %q", c.password, "pass")
	}
	if c.httpClient.Timeout != 30*time.Second {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, 30*time.Second)
	}
}

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	c, err := NewClient("http://example.com/", "user", "pass")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	if c.baseURL != "http://example.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "http://example.com")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	customTimeout := 60 * time.Second
	c, err := NewClient("http://example.com", "user", "pass", WithTimeout(customTimeout))
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	if c.httpClient.Timeout != customTimeout {
		t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, customTimeout)
	}
}

func TestNewClient_ValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		username string
		password string
		wantErr  bool
	}{
		{"empty host", "", "user", "pass", true},
		{"empty username", "http://example.com", "", "pass", true},
		{"empty password", "http://example.com", "user", "", true},
		{"whitespace host", "   ", "user", "pass", true},
		{"whitespace username", "http://example.com", "  ", "pass", true},
		{"whitespace password", "http://example.com", "user", "  ", true},
		{"valid", "http://example.com", "user", "pass", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.host, tt.username, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidConfig) {
				t.Errorf("error should wrap ErrInvalidConfig, got %v", err)
			}
		})
	}
}

func TestClient_buildURL(t *testing.T) {
	c := mustNewClient("http://example.com:8080", "testuser", "testpass")

	url, err := c.buildURL("get_live_categories", nil)
	if err != nil {
		t.Fatalf("buildURL error: %v", err)
	}

	// Check required params
	if !strings.Contains(url, "username=testuser") {
		t.Error("URL missing username param")
	}
	if !strings.Contains(url, "password=testpass") {
		t.Error("URL missing password param")
	}
	if !strings.Contains(url, "action=get_live_categories") {
		t.Error("URL missing action param")
	}
	if !strings.Contains(url, "/player_api.php") {
		t.Error("URL missing endpoint path")
	}
}

func TestClient_buildURL_WithParams(t *testing.T) {
	c := mustNewClient("http://example.com", "user", "pass")

	params := map[string]string{
		"category_id": "123",
	}
	url, err := c.buildURL("get_live_streams", params)
	if err != nil {
		t.Fatalf("buildURL error: %v", err)
	}

	if !strings.Contains(url, "category_id=123") {
		t.Error("URL missing category_id param")
	}
}

func TestClient_LiveStreamURL(t *testing.T) {
	c := mustNewClient("http://server.com:8080", "myuser", "mypass")

	url := c.LiveStreamURL("12345")
	expected := "http://server.com:8080/live/myuser/mypass/12345.ts"
	if url != expected {
		t.Errorf("LiveStreamURL = %q, want %q", url, expected)
	}
}

func TestClient_LiveStreamURLWithFormat(t *testing.T) {
	c := mustNewClient("http://server.com", "user", "pass")

	url := c.LiveStreamURLWithFormat("999", "m3u8")
	expected := "http://server.com/live/user/pass/999.m3u8"
	if url != expected {
		t.Errorf("LiveStreamURLWithFormat = %q, want %q", url, expected)
	}
}

func TestClient_VODStreamURL(t *testing.T) {
	c := mustNewClient("http://server.com", "user", "pass")

	url := c.VODStreamURL("555", "mkv")
	expected := "http://server.com/movie/user/pass/555.mkv"
	if url != expected {
		t.Errorf("VODStreamURL = %q, want %q", url, expected)
	}
}

func TestClient_SeriesEpisodeURL(t *testing.T) {
	c := mustNewClient("http://server.com", "user", "pass")

	url := c.SeriesEpisodeURL("777", "mp4")
	expected := "http://server.com/series/user/pass/777.mp4"
	if url != expected {
		t.Errorf("SeriesEpisodeURL = %q, want %q", url, expected)
	}
}

func TestClient_Authenticate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request params
		q := r.URL.Query()
		if q.Get("username") != "testuser" {
			t.Error("missing username")
		}
		if q.Get("password") != "testpass" {
			t.Error("missing password")
		}

		resp := AuthResponse{
			UserInfo: UserInfo{
				Username: "testuser",
				Status:   "Active",
			},
			ServerInfo: ServerInfo{
				URL: "example.com",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "testuser", "testpass")
	ctx := context.Background()

	auth, err := c.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if auth.UserInfo.Status != "Active" {
		t.Errorf("Status = %q, want %q", auth.UserInfo.Status, "Active")
	}
}

func TestClient_Authenticate_Inactive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AuthResponse{
			UserInfo: UserInfo{
				Status: "Disabled",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	_, err := c.Authenticate(ctx)
	if err == nil {
		t.Error("expected error for disabled account")
	}
	if !strings.Contains(err.Error(), "not active") {
		t.Errorf("error = %q, want to contain 'not active'", err.Error())
	}
}

func TestClient_GetLiveCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("action") != "get_live_categories" {
			t.Errorf("action = %q, want get_live_categories", q.Get("action"))
		}

		cats := []Category{
			{ID: NewFlexibleID(1), Name: "Sports"},
			{ID: NewFlexibleIDFromString("2"), Name: "News"},
		}
		json.NewEncoder(w).Encode(cats)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	cats, err := c.GetLiveCategories(ctx)
	if err != nil {
		t.Fatalf("GetLiveCategories error: %v", err)
	}
	if len(cats) != 2 {
		t.Errorf("len(cats) = %d, want 2", len(cats))
	}
	if cats[0].Name != "Sports" {
		t.Errorf("cats[0].Name = %q, want Sports", cats[0].Name)
	}
}

func TestClient_GetVODCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("action") != "get_vod_categories" {
			t.Errorf("action = %q, want get_vod_categories", q.Get("action"))
		}

		cats := []Category{
			{ID: NewFlexibleID(10), Name: "Action"},
		}
		json.NewEncoder(w).Encode(cats)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	cats, err := c.GetVODCategories(ctx)
	if err != nil {
		t.Fatalf("GetVODCategories error: %v", err)
	}
	if len(cats) != 1 {
		t.Errorf("len(cats) = %d, want 1", len(cats))
	}
}

func TestClient_GetSeriesCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("action") != "get_series_categories" {
			t.Errorf("action = %q, want get_series_categories", q.Get("action"))
		}

		cats := []Category{
			{ID: NewFlexibleID(20), Name: "Drama"},
		}
		json.NewEncoder(w).Encode(cats)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	cats, err := c.GetSeriesCategories(ctx)
	if err != nil {
		t.Fatalf("GetSeriesCategories error: %v", err)
	}
	if len(cats) != 1 {
		t.Errorf("len(cats) = %d, want 1", len(cats))
	}
}

func TestClient_GetLiveStreams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("action") != "get_live_streams" {
			t.Errorf("action = %q, want get_live_streams", q.Get("action"))
		}

		streams := []LiveStream{
			{ID: NewFlexibleID(100), Name: "ESPN", CategoryID: NewFlexibleID(1)},
			{ID: NewFlexibleID(101), Name: "CNN", CategoryID: NewFlexibleID(2)},
		}
		json.NewEncoder(w).Encode(streams)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	streams, err := c.GetLiveStreams(ctx, "")
	if err != nil {
		t.Fatalf("GetLiveStreams error: %v", err)
	}
	if len(streams) != 2 {
		t.Errorf("len(streams) = %d, want 2", len(streams))
	}
}

func TestClient_GetLiveStreams_WithCategory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("category_id") != "5" {
			t.Errorf("category_id = %q, want 5", q.Get("category_id"))
		}

		streams := []LiveStream{
			{ID: NewFlexibleID(200), Name: "Fox Sports"},
		}
		json.NewEncoder(w).Encode(streams)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	streams, err := c.GetLiveStreams(ctx, "5")
	if err != nil {
		t.Fatalf("GetLiveStreams error: %v", err)
	}
	if len(streams) != 1 {
		t.Errorf("len(streams) = %d, want 1", len(streams))
	}
}

func TestClient_GetVODStreams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("action") != "get_vod_streams" {
			t.Errorf("action = %q, want get_vod_streams", q.Get("action"))
		}

		vods := []VODStream{
			{ID: NewFlexibleID(500), Name: "Movie 1", Container: "mp4"},
		}
		json.NewEncoder(w).Encode(vods)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	vods, err := c.GetVODStreams(ctx, "")
	if err != nil {
		t.Fatalf("GetVODStreams error: %v", err)
	}
	if len(vods) != 1 {
		t.Errorf("len(vods) = %d, want 1", len(vods))
	}
}

func TestClient_GetSeries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("action") != "get_series" {
			t.Errorf("action = %q, want get_series", q.Get("action"))
		}

		series := []Series{
			{ID: NewFlexibleID(800), Name: "Breaking Bad"},
		}
		json.NewEncoder(w).Encode(series)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	series, err := c.GetSeries(ctx, "")
	if err != nil {
		t.Fatalf("GetSeries error: %v", err)
	}
	if len(series) != 1 {
		t.Errorf("len(series) = %d, want 1", len(series))
	}
}

func TestClient_GetSeriesInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("action") != "get_series_info" {
			t.Errorf("action = %q, want get_series_info", q.Get("action"))
		}
		if q.Get("series_id") != "123" {
			t.Errorf("series_id = %q, want 123", q.Get("series_id"))
		}

		info := SeriesInfo{
			Info: FlexibleSeriesInfo{
				SeriesDetails: SeriesDetails{
					Name: "The Office",
				},
			},
			Seasons: FlexibleSeasons{
				{SeasonNumber: NewFlexibleID(1), EpisodeCount: NewFlexibleID(6)},
			},
		}
		json.NewEncoder(w).Encode(info)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	info, err := c.GetSeriesInfo(ctx, "123")
	if err != nil {
		t.Fatalf("GetSeriesInfo error: %v", err)
	}
	if info.Info.Name != "The Office" {
		t.Errorf("Info.Name = %q, want The Office", info.Info.Name)
	}
}

func TestClient_GetShortEPG(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("action") != "get_short_epg" {
			t.Errorf("action = %q, want get_short_epg", q.Get("action"))
		}
		if q.Get("stream_id") != "999" {
			t.Errorf("stream_id = %q, want 999", q.Get("stream_id"))
		}

		epg := EPGShort{
			ID: NewFlexibleID(999),
			EPGListings: []EPGEntry{
				{Title: "News at 6"},
			},
		}
		json.NewEncoder(w).Encode(epg)
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	epg, err := c.GetShortEPG(ctx, "999")
	if err != nil {
		t.Fatalf("GetShortEPG error: %v", err)
	}
	if len(epg.EPGListings) != 1 {
		t.Errorf("len(EPGListings) = %d, want 1", len(epg.EPGListings))
	}
}

func TestClient_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	_, err := c.Authenticate(ctx)
	if err == nil {
		t.Error("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want to contain '500'", err.Error())
	}
}

func TestClient_JSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx := context.Background()

	_, err := c.Authenticate(ctx)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(AuthResponse{})
	}))
	defer server.Close()

	c := mustNewClient(server.URL, "user", "pass")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := c.Authenticate(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
