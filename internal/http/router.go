package httpapi

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func BuildRouter(d *gorm.DB) *chi.Mux {
	r := chi.NewRouter()
	h := NewHandlers(d)

	r.Get("/health", h.Health)

	// Пользователи (упрощённо)
	r.Post("/users", h.CreateUser)

	// Заметки
	r.Post("/notes", h.CreateNote)      // создаём заметку с тегами
	r.Get("/notes/{id}", h.GetNoteByID) // получаем заметку с автором и тегами

	return r
}
