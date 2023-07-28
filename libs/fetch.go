package libs

import (
	"context"
	"log"
)

func (s *Server) FetchData(ctx context.Context) (*Data, error) {

	github, github_err := s.FetchGithub(ctx)
	weather, weather_err := s.FetchWeather(ctx)
	scoreboard, scoreboard_err := s.FetchScoreboard(ctx)

	if github_err != nil {
		return nil, github_err
	}
	if weather_err != nil {
		log.Println("failed to fetch Weather API:", weather_err)
	}
	if scoreboard_err != nil {
		log.Println("failed to fetch ESPN API:", scoreboard_err)
	}

	return &Data{
		Github:     github,
		Weather:    weather,
		Scoreboard: scoreboard,
	}, nil
}
