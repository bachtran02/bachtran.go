package libs

import (
	"context"

	"github.com/bachtran02/bachtran.go/models"
)

func (s *Server) FetchData(ctx context.Context) (*models.Data, error) {

	github, github_err := s.FetchGithub(ctx)

	if github_err != nil {
		return nil, github_err
	}

	return &models.Data{
		Github: github,
	}, nil
}
