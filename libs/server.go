package libs

import (
	"io"
	"log"
	"net/http"

	"github.com/shurcooL/githubv4"

	md "github.com/bachtran02/bachtran.go/models"
)

type ExecuteTemplateFunc func(wr io.Writer, name string, data any) error

func NewServer(cfg md.Config, dataService *DataService, httpClient *http.Client, githubClient *githubv4.Client, assets http.FileSystem) *Server {
	s := &Server{
		cfg:          cfg,
		httpClient:   httpClient,
		githubClient: githubClient,
		dataService:  dataService,
		assets:       assets,
	}

	s.server = &http.Server{
		Addr:    s.cfg.ListenAddr,
		Handler: s.Routes(),
	}

	return s
}

type Server struct {
	cfg          md.Config
	httpClient   *http.Client
	githubClient *githubv4.Client
	dataService  *DataService
	server       *http.Server
	assets       http.FileSystem
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
