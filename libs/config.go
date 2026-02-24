package libs

import (
	"os"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

type Config struct {
	GitHub        GitHubConfig  `yaml:"github"`
	Homelab       HomelabConfig `yaml:"homelab"`
	MusicEndpoint string        `yaml:"music_endpoint"`
	ListenAddr    string        `yaml:"listen_addr"`
	Log           LogConfig     `yaml:"log"`
	// Scoreboard  ScoreboardConfig `yaml:"scoreboard"`
}

type GitHubConfig struct {
	AccessToken string `yaml:"access_token"`
	User        string `yaml:"user"`
}

// type ScoreboardConfig struct {
// 	Team string `yaml:"team"`
// }

type LogConfig struct {
	Level     slog.Level `yaml:"level"`
	Format    string     `yaml:"format"`
	AddSource bool       `yaml:"add_source"`
}

type HomelabConfig struct {
	HomelabServer struct {
		NodesConfig []NodeConfig `yaml:"nodes"`
	} `yaml:"homelab_server"`
}

type NodeConfig struct {
	Name            string `yaml:"name"`
	NodeExporterUrl string `yaml:"node_exporter_url"`
}
