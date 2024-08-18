package models

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
		TemperatureF float64 `json:"temp_f"`
		Condition    struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
		} `json:"condition"`
	} `json:"current"`
}
