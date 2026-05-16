package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"example.com/pz11-graphql/store"
)

// Handler содержит зависимости REST-обработчиков.
type Handler struct {
	store *store.Store
}

// NewHandler создаёт новый Handler с переданным хранилищем.
func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

// ---- DTO ---------------------------------------------------------------

// taskResponse — структура ответа для клиента.
type taskResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description *string `json:"description"`
	Done        bool   `json:"done"`
}

// createTaskRequest — тело запроса при создании задачи.
type createTaskRequest struct {
	Title       string `json:"title"`
	Description *string `json:"description"`
}

// updateTaskRequest — тело запроса при обновлении задачи (все поля опциональны).
type updateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Done        *bool   `json:"done"`
}

// errorResponse — единый формат ошибки.
type errorResponse struct {
	Error string `json:"error"`
}

// ---- helpers -----------------------------------------------------------

func toResponse(t *store.Task) taskResponse {
	return taskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Done:        t.Done,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// ---- handlers ----------------------------------------------------------

// ListTasks godoc
// GET /v1/tasks
// Возвращает список всех задач.
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.store.ListTasks()
	resp := make([]taskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, toResponse(t))
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetTask godoc
// GET /v1/tasks/{id}
// Возвращает одну задачу по ID.
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/v1/tasks/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing task id")
		return
	}

	t, err := h.store.GetTask(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toResponse(t))
}

// CreateTask godoc
// POST /v1/tasks
// Создаёт новую задачу.
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	t := h.store.CreateTask(req.Title, req.Description)
	writeJSON(w, http.StatusCreated, toResponse(t))
}

// UpdateTask godoc
// PATCH /v1/tasks/{id}
// Частично обновляет задачу.
func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/v1/tasks/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing task id")
		return
	}

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t, err := h.store.UpdateTask(id, req.Title, req.Description, req.Done)
	if err != nil {
		if errors.Is(err, errors.New("task not found")) || err.Error() == "task not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toResponse(t))
}

// DeleteTask godoc
// DELETE /v1/tasks/{id}
// Удаляет задачу.
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/v1/tasks/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing task id")
		return
	}

	if err := h.store.DeleteTask(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// extractID вычленяет ID из пути вида /v1/tasks/{id}.
func extractID(path, prefix string) string {
	return strings.TrimPrefix(path, prefix)
}
