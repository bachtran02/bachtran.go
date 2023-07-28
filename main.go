package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"portfolio/libs"
	"syscall"
	"time"

	"github.com/shurcooL/githubv4"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

var (
	//go:embed templates/**
	Templates embed.FS

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

	slog.Info("Starting bachtran.dev...")

	var (
		tmplFunc libs.ExecuteTemplateFunc
		assets   http.FileSystem
	)

	tmpl, err := template.New("").ParseFS(Templates, "templates/*.gohtml")
	if err != nil {
		slog.Error("failed to parse templates", slog.Any("error", err))
		os.Exit(-1)
	}

	tmplFunc = tmpl.ExecuteTemplate
	assets = http.FS(Assets)
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	githubClient := githubv4.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHub.AccessToken},
	)))

	s := libs.NewServer("null", cfg, httpClient, githubClient, assets, tmplFunc)
	go s.Start()
	defer s.Close()

	slog.Info(fmt.Sprintf("Started bachtran.dev on %s", cfg.ListenAddr))
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-si
}
