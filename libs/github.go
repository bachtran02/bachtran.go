package libs

import (
	"context"
	"html/template"

	"github.com/a-h/templ"
	"github.com/shurcooL/githubv4"

	"github.com/bachtran02/bachtran.go/models"
)

func (s *Server) FetchGithub(ctx context.Context) (*models.GitHubData, error) {
	var query struct {
		User struct {
			Login       string
			AvatarURL   string
			UserRepo    models.UserRepo     `graphql:"repository(name: $user)"`
			PinnedItems models.Repositories `graphql:"pinnedItems(first: 3, types: REPOSITORY)"`
		} `graphql:"user(login: $user)"`
	}
	variables := map[string]interface{}{
		"user":       githubv4.String(s.cfg.GitHub.User),
		"expression": githubv4.String("main:README.md"),
	}
	if err := s.githubClient.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	return &models.GitHubData{
		User: models.User{
			Name:      query.User.Login,
			AvatarURL: template.URL(query.User.AvatarURL),
		},
		HomeRaw:  query.User.UserRepo.Object.Blob.Text,
		Projects: parseRepositories(query.User.PinnedItems),
	}, nil
}

func parseRepositories(pinnedItems models.Repositories) []models.Project {
	projects := make([]models.Project, 0, len(pinnedItems.Nodes))
	for _, onode := range pinnedItems.Nodes {
		var language *models.Language
		node := onode.Nodes
		if len(node.Languages.Nodes) > 0 {
			lNode := node.Languages.Nodes[0]
			language = &models.Language{
				Name:  lNode.Name,
				Color: lNode.Color,
			}
		}

		topics := make([]models.Topic, 0, len(node.RepositoryTopics.Nodes))
		for _, tNode := range node.RepositoryTopics.Nodes {
			topics = append(topics, models.Topic{
				Name: tNode.Topic.Name,
				URL:  templ.SafeURL(tNode.URL),
			})
		}

		projects = append(projects, models.Project{
			Name:        node.Name,
			Description: node.Description,
			URL:         templ.SafeURL(node.URL),
			Stars:       node.StargazerCount,
			Forks:       node.ForkCount,
			UpdatedAt:   node.PushedAt,
			Language:    language,
			Topics:      topics,
		})
	}

	return projects
}
