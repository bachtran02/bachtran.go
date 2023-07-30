package libs

import (
	"html/template"
	"time"
)

type Data struct {
	Github  *GitHubData
	Weather *WeatherData
	Home    Home
}

type Home struct {
	Body    string
	Content template.HTML
}

type Project struct {
	Name        string
	Description string
	URL         template.URL
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
	URL  string
}
