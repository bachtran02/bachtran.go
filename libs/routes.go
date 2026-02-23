package libs

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bachtran02/bachtran.go/models"

	views "github.com/bachtran02/bachtran.go/views"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
)

func createFakeHomelabDevices() []models.HomelabStatus {
	return []models.HomelabStatus{
		{
			CPU: models.CPUMetrics{
				UsagePercent: 35.7,
				Cores:        8,
				Temperature:  52.3,
			},
			Memory: models.MemoryMetrics{
				Total:       17179869184,
				Used:        10737418240,
				UsedPercent: 62.5,
				Available:   6442450944,
			},
			Disk: models.DiskMetrics{
				Total:       1099511627776,
				Used:        659706494976,
				UsedPercent: 60.0,
				Available:   439805132800,
			},
			Network: models.NetworkMetrics{
				BytesSent:     549755813888,
				BytesReceived: 824633720832,
			},
			SystemInfo: models.SystemInfo{
				Hostname: "pi-cluster-01",
				OS:       "Ubuntu 22.04 LTS",
				Kernel:   "5.15.0-1048-raspi",
			},
			Uptime: 1814400, // 21 days
		},
		{
			CPU: models.CPUMetrics{
				UsagePercent: 18.3,
				Cores:        4,
				Temperature:  45.8,
			},
			Memory: models.MemoryMetrics{
				Total:       8589934592,
				Used:        3221225472,
				UsedPercent: 37.5,
				Available:   5368709120,
			},
			Disk: models.DiskMetrics{
				Total:       549755813888,
				Used:        247669399552,
				UsedPercent: 45.0,
				Available:   302086414336,
			},
			Network: models.NetworkMetrics{
				BytesSent:     274877906944,
				BytesReceived: 412316860416,
			},
			SystemInfo: models.SystemInfo{
				Hostname: "pi-cluster-02",
				OS:       "Ubuntu 22.04 LTS",
				Kernel:   "5.15.0-1048-raspi",
			},
			Uptime: 2592000, // 30 days
		},
		{
			CPU: models.CPUMetrics{
				UsagePercent: 72.4,
				Cores:        8,
				Temperature:  68.2,
			},
			Memory: models.MemoryMetrics{
				Total:       17179869184,
				Used:        14073748480,
				UsedPercent: 81.9,
				Available:   3106120704,
			},
			Disk: models.DiskMetrics{
				Total:       2199023255552,
				Used:        1759218604442,
				UsedPercent: 80.0,
				Available:   439804651110,
			},
			Network: models.NetworkMetrics{
				BytesSent:     1099511627776,
				BytesReceived: 1649267441664,
			},
			SystemInfo: models.SystemInfo{
				Hostname: "media-server",
				OS:       "Ubuntu 22.04 LTS",
				Kernel:   "5.15.0-91-generic",
			},
			Uptime: 604800, // 7 days
		},
		{
			CPU: models.CPUMetrics{
				UsagePercent: 8.5,
				Cores:        4,
				Temperature:  0, // No temp sensor
			},
			Memory: models.MemoryMetrics{
				Total:       4294967296,
				Used:        1288490189,
				UsedPercent: 30.0,
				Available:   3006477107,
			},
			Disk: models.DiskMetrics{
				Total:       274877906944,
				Used:        82463372083,
				UsedPercent: 30.0,
				Available:   192414534861,
			},
			Network: models.NetworkMetrics{
				BytesSent:     137438953472,
				BytesReceived: 206158430208,
			},
			SystemInfo: models.SystemInfo{
				Hostname: "dns-server",
				OS:       "Ubuntu 22.04 LTS",
				Kernel:   "5.15.0-1048-raspi",
			},
			Uptime: 3456000, // 40 days
		},
	}
}

func (s *Server) handleProjectsFragment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Project fragment requested!") // This will print on every click
	data, err := s.FetchData(r.Context())
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to fetch data: %w", err), http.StatusInternalServerError)
		return
	}
	views.Projects(*data).Render(r.Context(), w)
}

func (s *Server) handleHomelabFragment(w http.ResponseWriter, r *http.Request) {
	devices := createFakeHomelabDevices()

	// Render JUST the component
	views.Homelab(devices).Render(r.Context(), w)
}

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
		r.Route("/fragments", func(r chi.Router) {
			r.Get("/projects", s.handleProjectsFragment)
			r.Get("/homelab", s.handleHomelabFragment)
		})
		r.Route("/api", func(r chi.Router) {
			r.Route("/music", func(r chi.Router) {
				r.Get("/", s.music)
			})
			r.Route("/homelab", func(r chi.Router) {
				r.Get("/", s.homelabAPI)
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

	if err = s.ParseMarkdown(data); err != nil {
		s.error(w, r, fmt.Errorf("failed to parse ABOUTME.md: %w", err), http.StatusInternalServerError)
		return
	}

	ch := &templ.ComponentHandler{
		Component:   views.Index(*data),
		ContentType: "text/html; charset=utf-8",
	}
	ch.ServeHTTP(w, r)
}

func (s *Server) music(w http.ResponseWriter, r *http.Request) {
	views.Music().Render(r.Context(), w)
}

// func (s *Server) projects(w http.ResponseWriter, r *http.Request) {
// 	data, err := s.FetchData(r.Context())
// 	if err != nil {
// 		s.error(w, r, fmt.Errorf("failed to fetch data: %w", err), http.StatusInternalServerError)
// 		return
// 	}

// 	ch := &templ.ComponentHandler{
// 		Component:   views.ProjectsPage(*data),
// 		ContentType: "text/html; charset=utf-8",
// 	}
// 	ch.ServeHTTP(w, r)
// }

// func (s *Server) homelabPage(w http.ResponseWriter, r *http.Request) {
// 	status, err := s.prometheusClient.FetchHomelabStatus()
// 	if err != nil {
// 		s.error(w, r, fmt.Errorf("failed to fetch homelab status: %w", err), http.StatusInternalServerError)
// 		return
// 	}
// 	devices := []models.HomelabStatus{*status}

// 	ch := &templ.ComponentHandler{
// 		Component:   tmpl.HomelabPage(devices),
// 		ContentType: "text/html; charset=utf-8",
// 	}
// 	ch.ServeHTTP(w, r)
// }

func (s *Server) homelabAPI(w http.ResponseWriter, r *http.Request) {
	status, err := s.prometheusClient.FetchHomelabStatus()
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to fetch homelab status: %w", err), http.StatusInternalServerError)
		return
	}
	devices := []models.HomelabStatus{*status}
	views.Homelab(devices).Render(r.Context(), w)
}

func (s *Server) redirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) error(w http.ResponseWriter, r *http.Request, err error, status int) {
	if status == http.StatusInternalServerError {
		slog.ErrorCtx(r.Context(), "internal server error", slog.Any("error", err))
	}
	w.WriteHeader(status)

	vars := models.Error{
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
