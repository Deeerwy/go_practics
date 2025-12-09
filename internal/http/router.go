package httpx

import (
  "github.com/go-chi/chi/v5"
  "example.com/notes-api/internal/http/handlers"
  "example.com/notes-api/internal/repo"
)

func NewRouter(repoMem *repo.NoteRepoMem) *chi.Mux {
	h := &handlers.Handlers{Repo: repoMem}

	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/notes", h.ListNotes)
		r.Post("/notes", h.CreateNote)
		r.Get("/notes/{id}", h.GetNote)
		r.Patch("/notes/{id}", h.PatchNote)
		r.Delete("/notes/{id}", h.DeleteNote)
	})

	return r
}
