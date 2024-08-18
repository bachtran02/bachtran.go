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

	cfg, err := libs.LoadConfig(*cfgPath)
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
	githubClient := githubv4.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHub.AccessToken},
	)))

	s := libs.NewServer("null", cfg, httpClient, githubClient, assets)
	go s.Start()
	defer s.Close()

	slog.Info(fmt.Sprintf("Started bachtran.dev on %s", cfg.ListenAddr))
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-si
}

func setupLogger(cfg libs.LogConfig) {
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
