package models

import "github.com/a-h/templ"

type MusicTrack struct {
	Artist     string        `json:"artist"`
	ArtworkUrl string        `json:"artwork_url"`
	IsPlaying  bool          `json:"is_playing"`
	Title      string        `json:"title"`
	Uri        templ.SafeURL `json:"url"`
}
