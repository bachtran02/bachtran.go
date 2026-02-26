package libs

import (
	"context"
	"sync"
	"time"

	"github.com/bachtran02/bachtran.go/models"
)

type DataService struct {
	mu sync.RWMutex

	promClient  *PrometheusClient
	musicClient *MusicClient

	nodeStatuses []models.NodeStatus
	musicData    models.MusicStatus

	musicErr error
	nodeErr  error
}

func NewDataService(promClient *PrometheusClient, musicClient *MusicClient) *DataService {
	return &DataService{
		promClient:  promClient,
		musicClient: musicClient,
	}
}

func (ds *DataService) StartService(ctx context.Context) {
	/* Update data every 15 seconds */
	ticker := time.NewTicker(15 * time.Second)
	for {
		select {
		case <-ticker.C:
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				nodes, err := ds.promClient.FetchNodesStatus(ctx)
				ds.mu.Lock()
				defer ds.mu.Unlock()
				if err == nil {
					ds.nodeStatuses = nodes
				}
				ds.nodeErr = err
			}()

			go func() {
				defer wg.Done()
				data, err := ds.musicClient.FetchMusicStatus(ctx)
				ds.mu.Lock()
				defer ds.mu.Unlock()
				if err == nil {
					ds.musicData = *data
				}
				ds.musicErr = err
			}()

			wg.Wait()
		case <-ctx.Done():
			return
		}
	}
}

func (ds *DataService) GetMusicData() (models.MusicStatus, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.musicData, ds.musicErr
}

func (ds *DataService) GetNodeStatuses() ([]models.NodeStatus, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.nodeStatuses, ds.nodeErr
}
