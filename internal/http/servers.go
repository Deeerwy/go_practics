package http

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"example.com/go_practics/internal/storage/postgres"
	rediscache "example.com/go_practics/internal/storage/redis"
)

type Server struct {
	repo  *postgres.Repo
	cache *rediscache.Cache
}

func NewServer(repo *postgres.Repo, cache *rediscache.Cache) *Server {
	return &Server{repo: repo, cache: cache}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(5 * time.Second))
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	r.Route("/notes", func(r chi.Router) {
		r.Post("/", s.handleCreateNote)
		r.Get("/", s.handleListNotes)

		r.Get("/batch", s.handleBatchTitles)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.handleGetNote)
			r.Patch("/", s.handlePatchNote)
			r.Delete("/", s.handleDeleteNote)
		})
	})

	return r
}