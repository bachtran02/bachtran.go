package libs

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
)

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()

	r.Mount("/assets", http.FileServer(s.assets))
	r.Group(func(r chi.Router) {
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
	if err = s.ParseMarkdown(data); err != nil {
		log.Println("failed to parse ABOUTME.md:", err)
		return
	}

	if err = s.tmpl(w, "index.gohtml", data); err != nil {
		log.Println("failed to execute template:", err)
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
