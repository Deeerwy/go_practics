package handlers

import (
  "encoding/json"
  "net/http"
  "example.com/notes-api/internal/core"
  "example.com/notes-api/internal/repo"
  "github.com/go-chi/chi/v5"
  "strconv"
  "strings"
  "time"
)
// Handlers объединяет все HTTP-обработчики для заметок.
type Handlers struct {
	Repo *repo.NoteRepoMem
}

// writeJSON — вспомогательная функция для успешных ответов.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError — вспомогательная функция для ошибок в формате map[string]string{"error": "..."}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ListNotes godoc
// @Summary      Список заметок
// @Description  Возвращает список заметок с пагинацией и фильтром по заголовку
// @Tags         notes
// @Param        page   query  int     false  "Номер страницы"
// @Param        limit  query  int     false  "Размер страницы"
// @Param        q      query  string  false  "Поиск по title"
// @Success      200    {array}  core.Note
// @Header       200    {integer}  X-Total-Count  "Общее количество"
// @Failure      500    {object}  map[string]string
// @Router       /notes [get]
func (h *Handlers) ListNotes(w http.ResponseWriter, r *http.Request) {
	// Значения по умолчанию
	page := 1
	limit := 10

	// page
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	// limit
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	// фильтр по title
	q := r.URL.Query().Get("q")

	notes, total, err := h.Repo.List(page, limit, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list notes")
		return
	}

	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	writeJSON(w, http.StatusOK, notes)
}

// CreateNote godoc
// @Summary      Создать заметку
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        input  body     core.NoteCreate  true  "Данные новой заметки"
// @Success      201    {object} core.Note
// @Failure      400    {object} map[string]string
// @Failure      500    {object} map[string]string
// @Router       /notes [post]
func (h *Handlers) CreateNote(w http.ResponseWriter, r *http.Request) {
	var in core.NoteCreate
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Простейшая валидация — пустой title не принимаем
	if strings.TrimSpace(in.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	now := time.Now()
	note := core.Note{
		Title:     in.Title,
		Content:   in.Content,
		CreatedAt: now,
		UpdatedAt: nil,
	}

	id, err := h.Repo.Create(note)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create note")
		return
	}
	note.ID = id

	writeJSON(w, http.StatusCreated, note)
}

// GetNote godoc
// @Summary      Получить заметку
// @Tags         notes
// @Param        id   path   int  true  "ID"
// @Success      200  {object}  core.Note
// @Failure      404  {object}  map[string]string
// @Router       /notes/{id} [get]
func (h *Handlers) GetNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	note, ok := h.Repo.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}

	writeJSON(w, http.StatusOK, note)
}

// PatchNote godoc
// @Summary      Обновить заметку (частично)
// @Tags         notes
// @Accept       json
// @Param        id     path   int              true  "ID"
// @Param        input  body   core.NoteUpdate  true  "Поля для обновления"
// @Success      200    {object}  core.Note
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Router       /notes/{id} [patch]
func (h *Handlers) PatchNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var upd core.NoteUpdate
	if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Если оба поля nil — нечего обновлять
	if upd.Title == nil && upd.Content == nil {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	note, ok := h.Repo.Update(id, upd)
	if !ok {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}

	writeJSON(w, http.StatusOK, note)
}

// DeleteNote godoc
// @Summary      Удалить заметку
// @Tags         notes
// @Param        id  path  int  true  "ID"
// @Success      204  "No Content"
// @Failure      404  {object}  map[string]string
// @Router       /notes/{id} [delete]
func (h *Handlers) DeleteNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	ok := h.Repo.Delete(id)
	if !ok {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}

	// 204 — без тела
	w.WriteHeader(http.StatusNoContent)
}