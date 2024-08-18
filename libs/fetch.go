package libs

import (
	"context"
	"portfolio/models"

	"golang.org/x/exp/slog"
)

func (s *Server) FetchData(ctx context.Context) (*models.Data, error) {

	github, github_err := s.FetchGithub(ctx)
	weather, weather_err := s.FetchWeather(ctx)

	if github_err != nil {
		return nil, github_err
	}
	if weather_err != nil {
		slog.Error("failed to fetch Weather API:", weather_err)
	}

	return &models.Data{
		Github:  github,
		Weather: weather,
	}, nil
}
