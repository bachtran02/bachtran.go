package models

type MusicTrack struct {
	Title      string `json:"title"`
	Author     string `json:"author"`
	ArtworkURL string `json:"artworkUrl"`
	Length     int    `json:"length"`
	URI        string `json:"uri"`
}

type MusicStatus struct {
	Position int        `json:"position"`
	Playing  bool       `json:"playing"`
	Paused   bool       `json:"paused"`
	Track    MusicTrack `json:"track"`
}
