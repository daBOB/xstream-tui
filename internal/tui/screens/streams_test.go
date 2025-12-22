package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/tui"
	"github.com/altmueller/xstream-tui/internal/xc"
)

func TestLiveStreamItem(t *testing.T) {
	item := LiveStreamItem{xc.LiveStream{
		ID:           xc.NewFlexibleID(1),
		Name:         "CNN",
		EPGChannelID: "cnn.us",
	}}

	if item.FilterValue() != "CNN" {
		t.Errorf("FilterValue() = %q, want 'CNN'", item.FilterValue())
	}
	if item.Title() != "CNN" {
		t.Errorf("Title() = %q, want 'CNN'", item.Title())
	}
	if item.Description() != "cnn.us" {
		t.Errorf("Description() = %q, want 'cnn.us'", item.Description())
	}
}

func TestVODStreamItem(t *testing.T) {
	t.Run("with rating", func(t *testing.T) {
		item := VODStreamItem{xc.VODStream{
			ID:     xc.NewFlexibleID(1),
			Name:   "Movie",
			Rating: "8.5",
		}}

		if item.FilterValue() != "Movie" {
			t.Errorf("FilterValue() = %q, want 'Movie'", item.FilterValue())
		}
		if item.Title() != "Movie" {
			t.Errorf("Title() = %q, want 'Movie'", item.Title())
		}
		if item.Description() != "★ 8.5" {
			t.Errorf("Description() = %q, want '★ 8.5'", item.Description())
		}
	})

	t.Run("without rating", func(t *testing.T) {
		item := VODStreamItem{xc.VODStream{
			ID:   xc.NewFlexibleID(2),
			Name: "Another Movie",
		}}

		if item.Description() != "" {
			t.Errorf("Description() = %q, want ''", item.Description())
		}
	})
}

func TestSeriesItem(t *testing.T) {
	t.Run("with rating", func(t *testing.T) {
		item := SeriesItem{xc.Series{
			ID:     xc.NewFlexibleID(1),
			Name:   "Breaking Bad",
			Rating: "9.5",
		}}

		if item.FilterValue() != "Breaking Bad" {
			t.Errorf("FilterValue() = %q, want 'Breaking Bad'", item.FilterValue())
		}
		if item.Title() != "Breaking Bad" {
			t.Errorf("Title() = %q, want 'Breaking Bad'", item.Title())
		}
		if item.Description() != "★ 9.5" {
			t.Errorf("Description() = %q, want '★ 9.5'", item.Description())
		}
	})

	t.Run("without rating", func(t *testing.T) {
		item := SeriesItem{xc.Series{
			ID:   xc.NewFlexibleID(2),
			Name: "Show",
		}}

		if item.Description() != "" {
			t.Errorf("Description() = %q, want ''", item.Description())
		}
	})
}

func TestNewStreamsModel(t *testing.T) {
	m := NewStreamsModel()
	if m == nil {
		t.Fatal("NewStreamsModel() returned nil")
	}
	if m.list == nil {
		t.Error("list should be initialized")
	}
}

func TestStreamsModel_SetClient(t *testing.T) {
	m := NewStreamsModel()
	m.SetClient(nil)
	if m.client != nil {
		t.Error("client should be nil")
	}
}

func TestStreamsModel_SetSize(t *testing.T) {
	m := NewStreamsModel()
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestStreamsModel_SetStreams(t *testing.T) {
	m := NewStreamsModel()

	t.Run("live streams", func(t *testing.T) {
		msg := tui.StreamsLoadedMsg{
			LiveStreams: []xc.LiveStream{{ID: xc.NewFlexibleID(1), Name: "Test"}},
		}
		m.SetStreams(msg)
		if len(m.liveStreams) != 1 {
			t.Errorf("liveStreams = %d, want 1", len(m.liveStreams))
		}
	})

	t.Run("vod streams", func(t *testing.T) {
		msg := tui.StreamsLoadedMsg{
			VODStreams: []xc.VODStream{{ID: xc.NewFlexibleID(1), Name: "Movie"}},
		}
		m.SetStreams(msg)
		if len(m.vodStreams) != 1 {
			t.Errorf("vodStreams = %d, want 1", len(m.vodStreams))
		}
	})

	t.Run("series", func(t *testing.T) {
		msg := tui.StreamsLoadedMsg{
			Series: []xc.Series{{ID: xc.NewFlexibleID(1), Name: "Show"}},
		}
		m.SetStreams(msg)
		if len(m.series) != 1 {
			t.Errorf("series = %d, want 1", len(m.series))
		}
	})
}

func TestStreamsModel_Update_StreamsLoaded(t *testing.T) {
	m := NewStreamsModel()
	msg := tui.StreamsLoadedMsg{
		LiveStreams: []xc.LiveStream{{ID: xc.NewFlexibleID(1), Name: "Test"}},
	}

	m.Update(msg)
	if len(m.liveStreams) != 1 {
		t.Errorf("liveStreams = %d after update, want 1", len(m.liveStreams))
	}
}

func TestStreamsModel_Update_Search(t *testing.T) {
	m := NewStreamsModel()
	m.SetSize(80, 40)

	// Start search
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.searching {
		t.Error("should be in search mode after '/'")
	}

	// Escape cancels search
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.searching {
		t.Error("should exit search mode on Esc")
	}
}

func TestStreamsModel_View(t *testing.T) {
	m := NewStreamsModel()
	m.SetSize(80, 40)
	m.contentType = tui.LiveContent
	m.category = xc.Category{Name: "News"}

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// Test different content types
	m.contentType = tui.VODContent
	if m.View() == "" {
		t.Error("View() for VOD returned empty string")
	}

	m.contentType = tui.SeriesContent
	if m.View() == "" {
		t.Error("View() for Series returned empty string")
	}
}

func TestStreamsModel_View_Searching(t *testing.T) {
	m := NewStreamsModel()
	m.SetSize(80, 40)
	m.searching = true

	view := m.View()
	if view == "" {
		t.Error("View() in search mode returned empty string")
	}
}

func TestStreamsModel_LoadStreams_NoClient(t *testing.T) {
	m := NewStreamsModel()
	m.contentType = tui.LiveContent

	cmd := m.loadStreams()
	if cmd == nil {
		t.Fatal("loadStreams() returned nil")
	}

	msg := cmd()
	if _, ok := msg.(tui.ErrorMsg); !ok {
		t.Errorf("msg type = %T, want ErrorMsg", msg)
	}
}
