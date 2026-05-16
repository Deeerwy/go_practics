package rest

import (
	"net/http"
	"strings"
)


func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	// /v1/tasks  — коллекция
	mux.HandleFunc("/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListTasks(w, r)
		case http.MethodPost:
			h.CreateTask(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /v1/tasks/{id}  — отдельный ресурс
	mux.HandleFunc("/v1/tasks/", func(w http.ResponseWriter, r *http.Request) {
		// Защита от запроса ровно на /v1/tasks/ без ID
		id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
		if id == "" {
			http.Error(w, "missing task id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.GetTask(w, r)
		case http.MethodPatch:
			h.UpdateTask(w, r)
		case http.MethodDelete:
			h.DeleteTask(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
