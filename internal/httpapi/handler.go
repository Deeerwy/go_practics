package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"example.com/pz5-security/internal/student"
)

type Handler struct {
	repo *student.Repo
	stmt *sql.Stmt
}

func NewHandler(repo *student.Repo, stmt *sql.Stmt) *Handler {
	return &Handler{
		repo: repo,
		stmt: stmt,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"scheme": "https",
	})
}

func (h *Handler) GetStudentByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawID := r.URL.Query().Get("id")
	if rawID == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var st student.Student
	err = h.stmt.QueryRow(id).Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(st)
}

func RedirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	target := "https://localhost:8443" + r.RequestURI
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}
