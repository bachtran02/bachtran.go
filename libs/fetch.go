package libs

import (
	"context"

	"github.com/bachtran02/bachtran.go/models"
)

func (s *Server) FetchData(ctx context.Context) (*models.Data, error) {

	github, githubErr := s.FetchGithub(ctx)
	if githubErr != nil {
		return nil, githubErr
	}

	return &models.Data{
		Github:      github,
		NodesConfig: s.cfg.Homelab.Nodes,
	}, nil
}
