package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strconv"
	"strings"

	"example.com/go_practics/internal/pagination"
)

type createReq struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type patchReq struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (s *Server) handleCreateNote(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	if req.Title == "" || req.Content == "" {
		WriteError(w, http.StatusBadRequest, "title and content are required")
		return
	}

	n, err := s.repo.Create(r.Context(), req.Title, req.Content)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Можно сразу положить в кэш
	_ = s.cache.SetNote(r.Context(), n)

	WriteJSON(w, http.StatusCreated, n)
}

func (s *Server) handleGetNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	// 1) cache
	if n, hit, err := s.cache.GetNote(r.Context(), id); err == nil && hit {
		WriteJSON(w, http.StatusOK, map[string]any{"note": n, "cache": "hit"})
		return
	}

	// 2) db
	n, err := s.repo.Get(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = s.cache.SetNote(r.Context(), n)
	WriteJSON(w, http.StatusOK, map[string]any{"note": n, "cache": "miss"})
}

func (s *Server) handleListNotes(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	cursorStr := strings.TrimSpace(r.URL.Query().Get("cursor"))

	var cur *pagination.Cursor
	if cursorStr != "" {
		c, err := pagination.Decode(cursorStr)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid cursor")
			return
		}
		cur = &c
	}

	notes, next, err := s.repo.List(r.Context(), q, limit, cur)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var nextCursor string
	if next != nil {
		nextCursor, _ = pagination.Encode(*next)
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"items":       notes,
		"next_cursor": nextCursor,
	})
}

func (s *Server) handlePatchNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var req patchReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	if req.Title == "" || req.Content == "" {
		WriteError(w, http.StatusBadRequest, "title and content are required")
		return
	}

	n, err := s.repo.UpdateTx(r.Context(), id, req.Title, req.Content)
	if errors.Is(err, pgx.ErrNoRows) {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = s.cache.DelNote(r.Context(), id)
	WriteJSON(w, http.StatusOK, n)
}

func (s *Server) handleDeleteNote(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := s.repo.Delete(r.Context(), id); errors.Is(err, pgx.ErrNoRows) {
		WriteError(w, http.StatusNotFound, "not found")
		return
	} else if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = s.cache.DelNote(r.Context(), id)
	WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

// Демонстрация батчинга (ANY) для устранения N+1
// TODO: GET /notes/batch?ids=1,2,3
func (s *Server) handleBatchTitles(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("ids"))
	if raw == "" {
		WriteError(w, http.StatusBadRequest, "ids is required")
		return
	}

	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil || v <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid ids")
			return
		}
		ids = append(ids, v)
	}

	ctx := r.Context()
	// (необязательно) ограничим по времени, чтобы батчи не висели
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	m, err := s.repo.BatchTitles(ctx, ids)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{"titles": m})
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}