package libs

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/bachtran02/bachtran.go/models"
)

type MusicClient struct {
	APIEndpoint string
	HttpClient  *http.Client
}

func NewMusicClient(endpoint string) *MusicClient {
	return &MusicClient{
		APIEndpoint: endpoint,
		HttpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (mc *MusicClient) FetchMusicStatus(ctx context.Context) (*models.MusicStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mc.APIEndpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := mc.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("status code not ok", slog.Any("status", resp.StatusCode))
		return &models.MusicStatus{Playing: false}, nil
	}

	var status models.MusicStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		slog.Warn("failed to decode music status response", "error", err)
		return &models.MusicStatus{Playing: false}, nil
	}

	return &status, nil
}
