package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"tech-ip-sem2/services/auth/internal/service"
	"tech-ip-sem2/shared/middleware"
)

type AuthService interface {
	Login(username, password string) (string, error)
	Verify(token string) (bool, string)
}

type Handler struct {
	svc AuthService
}

func NewHandler(svc AuthService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/login", h.handleLogin)
	mux.HandleFunc("GET /v1/auth/verify", h.handleVerify)

	return middleware.Logging(middleware.RequestID(mux))
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	token, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	})
}

type verifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (h *Handler) handleVerify(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSON(w, http.StatusUnauthorized, verifyResponse{
			Valid: false,
			Error: "unauthorized",
		})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		writeJSON(w, http.StatusUnauthorized, verifyResponse{
			Valid: false,
			Error: "unauthorized",
		})
		return
	}

	token := parts[1]
	valid, subject := h.svc.Verify(token)
	if !valid {
		writeJSON(w, http.StatusUnauthorized, verifyResponse{
			Valid: false,
			Error: "unauthorized",
		})
		return
	}

	writeJSON(w, http.StatusOK, verifyResponse{
		Valid:   true,
		Subject: subject,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

