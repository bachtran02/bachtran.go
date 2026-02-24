package libs

import (
	"io"
	"log"
	"net/http"

	"github.com/shurcooL/githubv4"
)

type ExecuteTemplateFunc func(wr io.Writer, name string, data any) error

func NewServer(version string, cfg Config, httpClient *http.Client, githubClient *githubv4.Client, assets http.FileSystem) *Server {
	s := &Server{
		version:          version,
		cfg:              cfg,
		httpClient:       httpClient,
		githubClient:     githubClient,
		musicClient:      NewMusicClient(cfg.MusicEndpoint),
		prometheusClient: NewPrometheusClient(cfg.Homelab.HomelabServer.NodesConfig),
		assets:           assets,
	}

	s.server = &http.Server{
		Addr:    s.cfg.ListenAddr,
		Handler: s.Routes(),
	}

	return s
}

type Server struct {
	version          string
	cfg              Config
	httpClient       *http.Client
	githubClient     *githubv4.Client
	musicClient      *MusicClient
	prometheusClient *PrometheusClient
	server           *http.Server
	assets           http.FileSystem
}

func (s *Server) Start() {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalln("Error while listening:", err)
	}
}

func (s *Server) Close() {
	if err := s.server.Close(); err != nil {
		log.Println("Error while closing server:", err)
	}
}
