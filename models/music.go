package models

type MusicTrack struct {
	Artist     string `json:"artist"`
	ArtworkUrl string `json:"artwork_url"`
	IsPlaying  bool   `json:"is_playing"`
	Title      string `json:"title"`
	Uri        string `json:"url"`
}
