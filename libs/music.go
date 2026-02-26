package libs

import (
	"context"
	"encoding/json"
	"fmt"
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
		return nil, fmt.Errorf("api returned status: %d", resp.StatusCode)
	}

	var status models.MusicStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}
