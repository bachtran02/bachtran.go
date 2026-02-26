package models

type Data struct {
	Github      *GitHubData
	NodesConfig []NodeConfig
}

type Error struct {
	Error  string
	Status int
	Path   string
}
