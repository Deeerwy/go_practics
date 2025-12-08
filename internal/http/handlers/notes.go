package handlers

import (
  "encoding/json"
  "net/http"
  "example.com/notes-api/internal/core"
  "example.com/notes-api/internal/repo"
)

type Handler struct {
  Repo *repo.NoteRepoMem
}

func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
  var n core.Note
  if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
    http.Error(w, "Invalid input", http.StatusBadRequest)
    return
  }
  id, _ := h.Repo.Create(n)
  n.ID = id
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  json.NewEncoder(w).Encode(n)
}
