package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bachtran02/bachtran.go/models"
)

func (s *Server) FetchMusic(ctx context.Context) (*models.MusicTrack, error) {

	url := "http://MusicCatGo:8080/my_track"
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	rs, err := s.httpClient.Do(rq)
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	if rs.StatusCode != http.StatusOK {
		log.Printf("non-OK HTTP status: %d\nReason: %s", rs.StatusCode, http.StatusText(rs.StatusCode))
		return nil, fmt.Errorf("non-OK HTTP status: %d\tReason: %s", rs.StatusCode, http.StatusText(rs.StatusCode))
	}

	var track models.MusicTrack
	if err = json.NewDecoder(rs.Body).Decode(&track); err != nil {
		log.Println("failed to decode JSON:", err)
		return nil, err
	}
	return &track, nil
}
