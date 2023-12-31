package libs

import (
	"context"
	"html/template"
	"time"

	"github.com/shurcooL/githubv4"
)

type GitHubData struct {
	User     User
	HomeRaw  string
	Projects []Project
}

type User struct {
	Name      string
	AvatarURL template.URL
}

type UserRepo struct {
	Object struct {
		Blob struct {
			Text string
		} `graphql:"... on Blob"`
	} `graphql:"object(expression: $expression)"`
}

type Repositories struct {
	Nodes []struct {
		Nodes struct {
			Name             string
			URL              string
			Description      string
			StargazerCount   int
			ForkCount        int
			PushedAt         time.Time
			RepositoryTopics struct {
				Nodes []struct {
					Topic struct {
						Name string
					}
					URL string
				}
			} `graphql:"repositoryTopics(first: 5)"`
			Languages struct {
				Nodes []struct {
					Name  string
					Color string
				}
			} `graphql:"languages(first: 1, orderBy: {field: SIZE, direction: DESC})"`
		} `graphql:"... on Repository"`
	}
	TotalCount int
}

func (s *Server) FetchGithub(ctx context.Context) (*GitHubData, error) {
	var query struct {
		User struct {
			Login       string
			AvatarURL   string
			UserRepo    UserRepo     `graphql:"repository(name: $user)"`
			PinnedItems Repositories `graphql:"pinnedItems(first: 3, types: REPOSITORY)"`
		} `graphql:"user(login: $user)"`
	}
	variables := map[string]interface{}{
		"user":       githubv4.String(s.cfg.GitHub.User),
		"expression": githubv4.String("main:README.md"),
	}
	if err := s.githubClient.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	return &GitHubData{
		User: User{
			Name:      query.User.Login,
			AvatarURL: template.URL(query.User.AvatarURL),
		},
		HomeRaw:  query.User.UserRepo.Object.Blob.Text,
		Projects: parseRepositories(query.User.PinnedItems),
	}, nil
}

func parseRepositories(pinnedItems Repositories) []Project {
	projects := make([]Project, 0, len(pinnedItems.Nodes))
	for _, onode := range pinnedItems.Nodes {
		var language *Language
		node := onode.Nodes
		if len(node.Languages.Nodes) > 0 {
			lNode := node.Languages.Nodes[0]
			language = &Language{
				Name:  lNode.Name,
				Color: lNode.Color,
			}
		}

		topics := make([]Topic, 0, len(node.RepositoryTopics.Nodes))
		for _, tNode := range node.RepositoryTopics.Nodes {
			topics = append(topics, Topic{
				Name: tNode.Topic.Name,
				URL:  tNode.URL,
			})
		}

		projects = append(projects, Project{
			Name:        node.Name,
			Description: node.Description,
			URL:         template.URL(node.URL),
			Stars:       node.StargazerCount,
			Forks:       node.ForkCount,
			UpdatedAt:   node.PushedAt,
			Language:    language,
			Topics:      topics,
		})
	}

	return projects
}
