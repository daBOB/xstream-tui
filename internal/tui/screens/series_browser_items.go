package screens

import (
	"fmt"

	"github.com/altmueller/xstream-tui/internal/xc"
)

// seasonBrowserItem wraps SeasonInfo for FancyList.
type seasonBrowserItem struct {
	xc.SeasonInfo
	ActualEpisodeCount int
}

func (s seasonBrowserItem) FilterValue() string { return s.Name }
func (s seasonBrowserItem) Title() string       { return s.Name }
func (s seasonBrowserItem) Description() string {
	return fmt.Sprintf("%d episodes", s.ActualEpisodeCount)
}

// episodeBrowserItem wraps Episode for FancyList.
type episodeBrowserItem struct {
	xc.Episode
}

func (e episodeBrowserItem) FilterValue() string { return e.Episode.Title }
func (e episodeBrowserItem) Title() string       { return e.Episode.Title }
func (e episodeBrowserItem) Description() string {
	if e.Info.Duration != "" {
		return "⏱ " + e.Info.Duration
	}
	return ""
}
