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
	GitHub      GitHubConfig     `yaml:"github"`
	WeatherApi  WeatherApiConfig `yaml:"weather_api"`
	Scoreboard  ScoreboardConfig `yaml:"scoreboard"`
	ListenAddr  string           `yaml:"listen_addr"`
	AboutMePath string           `yaml:"aboutme_path"`
	Log         LogConfig        `yaml:"log"`
}

type GitHubConfig struct {
	AccessToken string `yaml:"access_token"`
	User        string `yaml:"user"`
}

type WeatherApiConfig struct {
	ApiKey string `yaml:"api_key"`
	City   string `yaml:"city"`
}

type ScoreboardConfig struct {
	Team     string `yaml:"team"`
	Timezone string `yaml:"timezone"`
}

type LogConfig struct {
	Level     slog.Level `yaml:"level"`
	Format    string     `yaml:"format"`
	AddSource bool       `yaml:"add_source"`
}
