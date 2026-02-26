package libs

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	md "github.com/bachtran02/bachtran.go/models"

	views "github.com/bachtran02/bachtran.go/views"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
)

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.CleanPath)
	r.Use(middleware.RealIP)
	r.Use(middleware.Maybe(
		middleware.RequestLogger(&middleware.DefaultLogFormatter{
			Logger:  log.Default(),
			NoColor: true,
		}),
		func(r *http.Request) bool {
			prefixes := []string{"/assets", "/api"}
			for _, prefix := range prefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					return false
				}
			}
			return true
		},
	))

	// r.Use(middleware.NoCache)
	r.Use(cacheControl)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	r.Mount("/assets", http.FileServer(s.assets))
	r.Group(func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Route("/music", func(r chi.Router) {
				r.Get("/", s.music)
			})
			r.Route("/homelab", func(r chi.Router) {
				r.Get("/", s.homelab)
			})
		})
		r.Route("/", func(r chi.Router) {
			r.Get("/", s.index)
			r.Head("/", s.index)
		})
	})
	r.NotFound(s.redirectRoot)

	return r
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	data, err := s.FetchData(r.Context())
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to fetch data: %w", err), http.StatusInternalServerError)
		return
	}

	ch := &templ.ComponentHandler{
		Component:   views.Index(*data),
		ContentType: "text/html; charset=utf-8",
	}
	ch.ServeHTTP(w, r)
}

func (s *Server) music(w http.ResponseWriter, r *http.Request) {

	musicStatus, err := s.dataService.GetMusicData()
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to fetch music status: %w", err), http.StatusInternalServerError)
		return
	}
	views.MusicPlayerContent(musicStatus).Render(r.Context(), w)
}

func (s *Server) homelab(w http.ResponseWriter, r *http.Request) {
	nodeStatuses, err := s.dataService.GetNodeStatuses()
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to fetch status: %w", err), http.StatusInternalServerError)
		return
	}
	if len(nodeStatuses) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	views.Homelab(nodeStatuses).Render(r.Context(), w)
}

func (s *Server) redirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) error(w http.ResponseWriter, r *http.Request, err error, status int) {
	if status == http.StatusInternalServerError {
		slog.ErrorCtx(r.Context(), "internal server error", slog.Any("error", err))
	}
	w.WriteHeader(status)

	vars := md.Error{
		Error:  err.Error(),
		Status: status,
		Path:   r.URL.Path,
	}

	ch := &templ.ComponentHandler{
		Component:   views.Error(vars),
		ContentType: "text/html; charset=utf-8",
	}
	ch.ServeHTTP(w, r)
}

func cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			// Disable caching for assets in development
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		next.ServeHTTP(w, r)
	})
}
