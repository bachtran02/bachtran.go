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
		Name           string  `json:"name"`
		Region         string  `json:"region"`
		Country        string  `json:"country"`
		Latitude       float64 `json:"lat"`
		Longitude      float64 `json:"lon"`
		TimezoneID     string  `json:"tz_id"`
		LocaltimeEpoch int64   `json:"localtime_epoch"`
		Localtime      string  `json:"localtime"`
	} `json:"location"`
	CurrentData struct {
		LastUpdatedEpoch int64   `json:"last_updated_epoch"`
		LastUpdated      string  `json:"last_updated"`
		TemperatureC     float64 `json:"temp_c"`
		TemperatureF     float64 `json:"temp_f"`
		IsDay            int     `json:"is_day"`
		Condition        struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		WindSpeedMPH    float64 `json:"wind_mph"`
		WindSpeedKPH    float64 `json:"wind_kph"`
		WindDegree      int     `json:"wind_degree"`
		WindDirection   string  `json:"wind_dir"`
		PressureMB      float64 `json:"pressure_mb"`
		PressureIN      float64 `json:"pressure_in"`
		PrecipitationMM float64 `json:"precip_mm"`
		PrecipitationIN float64 `json:"precip_in"`
		Humidity        int     `json:"humidity"`
		Cloudiness      int     `json:"cloud"`
		FeelsLikeC      float64 `json:"feelslike_c"`
		FeelsLikeF      float64 `json:"feelslike_f"`
		VisibilityKM    float64 `json:"vis_km"`
		VisibilityMiles float64 `json:"vis_miles"`
		UVIndex         float64 `json:"uv"`
		GustSpeedMPH    float64 `json:"gust_mph"`
		GustSpeedKPH    float64 `json:"gust_kph"`
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
