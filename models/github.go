package models

import (
	"time"

	"github.com/a-h/templ"
)

type Project struct {
	Name        string
	Description string
	URL         templ.SafeURL
	Stars       int
	Forks       int
	UpdatedAt   time.Time
	Language    *Language
	Topics      []Topic
}

type Language struct {
	Name  string
	Color string
}

type Topic struct {
	Name string
	URL  templ.SafeURL
}

type GitHubData struct {
	User     User
	Projects []Project
}

type User struct {
	Name      string
	AvatarURL string
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
