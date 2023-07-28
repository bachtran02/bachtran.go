package libs

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
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
		fmt.Println("failed to fetch data: %w", err)
		return
	}
	if err = s.ParseMarkdown(data); err != nil {
		log.Println("failed to parse README.md:", err)
		return
	}

	if err = s.tmpl(w, "index.gohtml", data); err != nil {
		log.Println("failed to execute template:", err)
	}
}

func (s *Server) redirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
