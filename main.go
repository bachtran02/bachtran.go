package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bachtran02/bachtran.go/libs"
	md "github.com/bachtran02/bachtran.go/models"
	"go.yaml.in/yaml/v2"

	"github.com/shurcooL/githubv4"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

var (
	//go:embed assets
	Assets embed.FS
)

func main() {
	cfgPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		os.Exit(-1)
	}
	setupLogger(cfg.Log)

	slog.Info("Starting bachtran.dev...")

	var assets http.FileSystem = http.FS(Assets)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	githubClient := githubv4.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: cfg.GitHub.AccessToken,
				},
			)))
	musicClient := libs.NewMusicClient(cfg.MusicEndpoint)
	prometheusClient := libs.NewPrometheusClient(cfg.Homelab.Nodes)

	dataService := libs.NewDataService(prometheusClient, musicClient)
	go dataService.StartService(context.Background())

	s := libs.NewServer(cfg, dataService, httpClient, githubClient, assets)
	go s.Start()
	defer s.Close()

	slog.Info(fmt.Sprintf("Started bachtran.dev on %s", cfg.ListenAddr))
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-si
}

func LoadConfig(path string) (md.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return md.Config{}, err
	}
	var cfg md.Config
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return md.Config{}, err
	}
	return cfg, nil
}

func setupLogger(cfg md.LogConfig) {
	opts := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     cfg.Level,
	}
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}
