package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/bachtran02/bachtran.go/models"
)

func (s *Server) FetchWeather(ctx context.Context) (*models.WeatherData, error) {
	url := fmt.Sprintf("http://api.weatherapi.com/v1/%s?key=%s&q=%s&aqi=no", "current.json", s.cfg.WeatherApi.ApiKey, url.QueryEscape(s.cfg.WeatherApi.City))
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

	var resp models.WeatherApiResponse
	if err = json.NewDecoder(rs.Body).Decode(&resp); err != nil {
		log.Println("failed to decode JSON:", err)
		return nil, err
	}
	return &models.WeatherData{
		City:         resp.LocationData.Name,
		Country:      resp.LocationData.Country,
		WeatherIcon:  fmt.Sprintf("https:%s", resp.CurrentData.Condition.Icon),
		WeatherText:  resp.CurrentData.Condition.Text,
		TemperatureF: resp.CurrentData.TemperatureF,
	}, nil
}
