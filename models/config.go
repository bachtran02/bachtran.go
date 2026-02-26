package models

import "golang.org/x/exp/slog"

type Config struct {
	GitHub        GitHubConfig  `yaml:"github"`
	Homelab       HomelabConfig `yaml:"homelab"`
	MusicEndpoint string        `yaml:"music_endpoint"`
	ListenAddr    string        `yaml:"listen_addr"`
	Log           LogConfig     `yaml:"log"`
}

type GitHubConfig struct {
	AccessToken string `yaml:"access_token"`
	User        string `yaml:"user"`
}

type LogConfig struct {
	Level     slog.Level `yaml:"level"`
	Format    string     `yaml:"format"`
	AddSource bool       `yaml:"add_source"`
}

type HomelabConfig struct {
	Nodes []NodeConfig `yaml:"nodes"`
}

type NodeConfig struct {
	Name            string `yaml:"name"`
	NodeExporterUrl string `yaml:"node_exporter_url"`
}
