package models

import (
	"html/template"
)

type Data struct {
	Github *GitHubData
	Home   Home
}

type Home struct {
	Body    string
	Content template.HTML
}

type Error struct {
	Error  string
	Status int
	Path   string
}
