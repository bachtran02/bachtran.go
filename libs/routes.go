package libs

import (
	"fmt"
	"log"
	"net/http"
	"portfolio/tmpl"
	"strings"

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
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	r.Mount("/assets", http.FileServer(s.assets))
	r.Group(func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Route("/scoreboard", func(r chi.Router) {
				r.Get("/", s.scoreboard)
			})
		})
		r.Route("/", func(r chi.Router) {
			// r.Get("/", templ.Handler(tmpl.Index("peter")).ServeHTTP)
			r.Get("/", s.index)
			// r.Head("/", s.index)
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

	if err = s.ParseMarkdown(data); err != nil {
		s.error(w, r, fmt.Errorf("failed to parse ABOUTME.md: %w", err), http.StatusInternalServerError)
		return
	}

	ch := &templ.ComponentHandler{
		Component:   tmpl.Index(*data),
		ContentType: "text/html; charset=utf-8",
	}
	ch.ServeHTTP(w, r)
}

func (s *Server) scoreboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := s.FetchScoreboard(ctx)

	if err := s.tmpl(w, "scoreboard.gohtml", vars); err != nil {
		slog.ErrorCtx(ctx, "failed to render scoreboard template", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) redirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) error(w http.ResponseWriter, r *http.Request, err error, status int) {
	if status == http.StatusInternalServerError {
		slog.ErrorCtx(r.Context(), "internal server error", slog.Any("error", err))
	}
	w.WriteHeader(status)

	vars := map[string]any{
		"Error":  err.Error(),
		"Status": status,
		"Path":   r.URL.Path,
	}
	if tmplErr := s.tmpl(w, "error.gohtml", vars); tmplErr != nil && tmplErr != http.ErrHandlerTimeout {
		slog.ErrorCtx(r.Context(), "failed to render error template", slog.Any("error", tmplErr))
	}
}
