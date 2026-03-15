package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"tech-ip-sem2/services/tasks/internal/client/authclient"
	"tech-ip-sem2/services/tasks/internal/service"
	"tech-ip-sem2/shared/middleware"
)

type AuthClient interface {
	Verify(ctx context.Context, authorization string) error
}

type TaskStore interface {
	Create(t service.Task) service.Task
	List() []service.Task
	Get(id string) (service.Task, error)
	Update(id string, update service.Task) (service.Task, error)
	Delete(id string) error
}

type Handler struct {
	store      TaskStore
	authClient AuthClient
}

func NewHandler(store TaskStore, authClient AuthClient) *Handler {
	return &Handler{store: store, authClient: authClient}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/tasks", h.withAuth(h.handleCreate))
	mux.HandleFunc("GET /v1/tasks", h.withAuth(h.handleList))
	mux.HandleFunc("GET /v1/tasks/", h.withAuth(h.handleGet))       // expect /v1/tasks/{id}
	mux.HandleFunc("PATCH /v1/tasks/", h.withAuth(h.handleUpdate))  // expect /v1/tasks/{id}
	mux.HandleFunc("DELETE /v1/tasks/", h.withAuth(h.handleDelete)) // expect /v1/tasks/{id}

	return middleware.Logging(middleware.RequestID(mux))
}

// withAuth wraps handler with auth verification via Auth service.
func (h *Handler) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		authHeader := r.Header.Get("Authorization")
		if err := h.authClient.Verify(ctx, authHeader); err != nil {
			if errors.Is(err, authclient.ErrUnauthorized) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if errors.Is(err, authclient.ErrForbidden) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			// Auth недоступен / внутренняя ошибка gRPC → 503.
			http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	task := service.Task{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Done:        false,
	}
	created := h.store.Create(task)
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	tasks := h.store.List()
	// For list, spec shows short form; keep all fields but that's acceptable.
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := lastPathPart(r.URL.Path)
	task, err := h.store.Get(id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

type updateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	id := lastPathPart(r.URL.Path)
	var upd service.Task
	if req.Title != nil {
		upd.Title = *req.Title
	}
	if req.Description != nil {
		upd.Description = *req.Description
	}
	if req.DueDate != nil {
		upd.DueDate = *req.DueDate
	}
	if req.Done != nil {
		upd.Done = *req.Done
	}

	task, err := h.store.Update(id, upd)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := lastPathPart(r.URL.Path)
	if err := h.store.Delete(id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func lastPathPart(path string) string {
	for len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

