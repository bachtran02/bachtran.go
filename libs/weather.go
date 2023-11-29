package libs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type WeatherData struct {
	City         string
	Country      string
	WeatherIcon  string
	WeatherText  string
	TemperatureF float64
}

type WeatherApiResponse struct {
	LocationData struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	CurrentData struct {
		// TemperatureC     float64 `json:"temp_c"`
		TemperatureF float64 `json:"temp_f"`
		Condition    struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
		} `json:"condition"`
		// UVIndex float64 `json:"uv"`
	} `json:"current"`
}

func (s *Server) FetchWeather(ctx context.Context) (*WeatherData, error) {
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

	var resp WeatherApiResponse
	if err = json.NewDecoder(rs.Body).Decode(&resp); err != nil {
		log.Println("failed to decode JSON:", err)
		return nil, err
	}
	return &WeatherData{
		City:         resp.LocationData.Name,
		Country:      resp.LocationData.Country,
		WeatherIcon:  fmt.Sprintf("https:%s", resp.CurrentData.Condition.Icon),
		WeatherText:  resp.CurrentData.Condition.Text,
		TemperatureF: resp.CurrentData.TemperatureF,
	}, nil
}
