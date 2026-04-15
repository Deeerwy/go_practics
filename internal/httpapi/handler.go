package httpapi

import (
	"html/template"
	"net/http"
	"strings"

	"example.com/pz6-web-security/internal/auth"
	"example.com/pz6-web-security/internal/store"
)

type Handler struct {
	store        *store.Store
	profileTmpl  *template.Template
	helloTmpl    *template.Template
	commentsTmpl *template.Template // добавлено
}

func NewHandler(s *store.Store) (*Handler, error) {
	profileTmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		return nil, err
	}

	helloTmpl, err := template.ParseFiles("templates/hello.html")
	if err != nil {
		return nil, err
	}

	// Загружаем новый шаблон
	commentsTmpl, err := template.ParseFiles("templates/comments.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		store:        s,
		profileTmpl:  profileTmpl,
		helloTmpl:    helloTmpl,
		commentsTmpl: commentsTmpl,
	}, nil
}


func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID, err := auth.RandomToken(16)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	csrfToken, err := auth.RandomToken(16)
	if err != nil {
		http.Error(w, "failed to create csrf token", http.StatusInternalServerError)
		return
	}

	h.store.Save(&store.UserProfile{
		SessionID: sessionID,
		Name:      "Студент",
		CSRFToken: csrfToken,
	})

	auth.SetSessionCookie(w, sessionID)

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	sessionID, err := auth.ReadSessionCookie(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profile, ok := h.store.Get(sessionID)
	if !ok {
		http.Error(w, "session not found", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		data := struct {
			Name      string
			CSRFToken string
		}{
			Name:      profile.Name,
			CSRFToken: profile.CSRFToken,
		}

		if err := h.profileTmpl.Execute(w, data); err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		tokenFromForm := r.FormValue("csrf_token")
		if tokenFromForm == "" || tokenFromForm != profile.CSRFToken {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		h.store.UpdateName(sessionID, name)
		http.Redirect(w, r, "/hello", http.StatusFound)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Hello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID, err := auth.ReadSessionCookie(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profile, ok := h.store.Get(sessionID)
	if !ok {
		http.Error(w, "session not found", http.StatusUnauthorized)
		return
	}

	data := struct {
		Name string
	}{
		Name: profile.Name,
	}

	if err := h.helloTmpl.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Comments(w http.ResponseWriter, r *http.Request) {
	// Проверяем наличие сессии
	sessionID, err := auth.ReadSessionCookie(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profile, ok := h.store.Get(sessionID)
	if !ok {
		http.Error(w, "session not found", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Передаём данные в шаблон
		data := struct {
			Name      string
			CSRFToken string
			Comments  []string
		}{
			Name:      profile.Name,
			CSRFToken: profile.CSRFToken,
			Comments:  profile.Comments,
		}

		if err := h.commentsTmpl.Execute(w, data); err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		// Проверка CSRF-токена — аналогично обработчику Profile
		tokenFromForm := r.FormValue("csrf_token")
		if tokenFromForm == "" || tokenFromForm != profile.CSRFToken {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}

		comment := strings.TrimSpace(r.FormValue("comment"))
		if comment == "" {
			http.Error(w, "comment is required", http.StatusBadRequest)
			return
		}

		h.store.AddComment(sessionID, comment)

		// После добавления — редирект обратно на страницу комментариев
		http.Redirect(w, r, "/comments", http.StatusFound)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
