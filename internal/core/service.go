package core

import (
    "encoding/json"
    "net/http"
    "sync"
	"github.com/golang-jwt/jwt/v5"
	"strconv"

    "github.com/go-chi/chi/v5"	
)

type userRepo interface {
    CheckPassword(email, pass string) (*User, error)
    ByID(id int64) (*User, error)
}


type jwtAccess interface {
    Sign(userID int64, email, role string) (string, error)
}
type jwtRefresh interface {
    SignRefresh(userID int64, email, role string) (string, string, error)
    Parse(tokenStr string) (jwt.MapClaims, error) // ← вместо map[string]any
}

type Service struct {
    repo        userRepo
    accessJWT   jwtAccess
    refreshJWT  jwtRefresh
    revokedMu   sync.RWMutex
    revokedJTI  map[string]int64 // jti -> exp (unix)
}

func NewService(r userRepo, access jwtAccess, refresh jwtRefresh) *Service {
    return &Service{
        repo:       r,
        accessJWT:  access,
        refreshJWT: refresh,
        revokedJTI: make(map[string]int64),
    }
}

func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
    var in struct{ Email, Password string }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Email == "" || in.Password == "" {
        httpError(w, 400, "invalid_credentials"); return
    }
    u, err := s.repo.CheckPassword(in.Email, in.Password)
    if err != nil {
        httpError(w, 401, "unauthorized"); return
    }
    access, err := s.accessJWT.Sign(u.ID, u.Email, u.Role)
    if err != nil { httpError(w, 500, "token_error"); return }
    refresh, _, err := s.refreshJWT.SignRefresh(u.ID, u.Email, u.Role)
    if err != nil { httpError(w, 500, "token_error"); return }

    // Возвращаем пару токенов
    jsonOK(w, map[string]any{
        "access":  access,
        "refresh": refresh,
        // можно вернуть jti для дебага, но в проде обычно не возвращают
    })
}

func (s *Service) RefreshHandler(w http.ResponseWriter, r *http.Request) {
    var in struct{ Refresh string }
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Refresh == "" {
        httpError(w, 400, "invalid_refresh"); return
    }
    claims, err := s.refreshJWT.Parse(in.Refresh)
    if err != nil {
        httpError(w, 401, "unauthorized"); return
    }

    // Извлекаем jti, exp, субьекта
    jti, _ := claims["jti"].(string)
    expF, _ := claims["exp"].(float64)
    userIDF, _ := claims["sub"].(float64)
    email, _ := claims["email"].(string)
    role, _ := claims["role"].(string)
    if jti == "" || expF == 0 || email == "" || role == "" {
        httpError(w, 400, "bad_claims"); return
    }

    // Проверяем blacklist
    if s.isRevoked(jti) {
        httpError(w, 401, "refresh_revoked"); return
    }
    // Отзываем использованный refresh (one-time use)
    s.revoke(jti, int64(expF))

    // Выпускаем новую пару access+refresh
    access, err := s.accessJWT.Sign(int64(userIDF), email, role)
    if err != nil { httpError(w, 500, "token_error"); return }
    refresh, _, err := s.refreshJWT.SignRefresh(int64(userIDF), email, role)
    if err != nil { httpError(w, 500, "token_error"); return }

    jsonOK(w, map[string]any{
        "access":  access,
        "refresh": refresh,
        // newJTI можно не возвращать
    })
}

func (s *Service) MeHandler(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value(ctxClaims{}).(map[string]any)
    jsonOK(w, map[string]any{
        "id": claims["sub"], "email": claims["email"], "role": claims["role"],
    })
}

func (s *Service) AdminStats(w http.ResponseWriter, r *http.Request) {
    jsonOK(w, map[string]any{"users": 2, "version": "1.0"})
}

// blacklist helpers
func (s *Service) isRevoked(jti string) bool {
    s.revokedMu.RLock()
    defer s.revokedMu.RUnlock()
    _, ok := s.revokedJTI[jti]
    return ok
}
func (s *Service) revoke(jti string, exp int64) {
    s.revokedMu.Lock()
    defer s.revokedMu.Unlock()
    s.revokedJTI[jti] = exp
}

// утилиты/контекст
type ctxClaims struct{}
func jsonOK(w http.ResponseWriter, v any) { w.Header().Set("Content-Type","application/json"); _ = json.NewEncoder(w).Encode(v) }
func httpError(w http.ResponseWriter, code int, msg string){
    w.Header().Set("Content-Type","application/json"); w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
func (s *Service) UserHandler(w http.ResponseWriter, r *http.Request) {
    // извлекаем id из пути
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        httpError(w, 400, "bad_id")
        return
    }

    // достаём клеймы из контекста
    claimsAny := r.Context().Value(CtxClaimsKey{})
    if claimsAny == nil {
        httpError(w, 401, "unauthorized")
        return
    }
    claims, ok := claimsAny.(map[string]any)
    if !ok {
        httpError(w, 500, "internal_error")
        return
    }

    // роль и sub
    role, _ := claims["role"].(string)

    // sub может быть float64 (jwt.MapClaims), приводим
    var subID int64
    switch v := claims["sub"].(type) {
    case float64:
        subID = int64(v)
    case int64:
        subID = v
    case int:
        subID = int64(v)
    case string:
        // если sub хранится как строка
        if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
            subID = parsed
        }
    }

    // ABAC: если роль user — разрешаем только свой профиль
    if role == "user" && id != subID {
        httpError(w, 403, "forbidden")
        return
    }

    // далее репозиторий возвращает пользователя по id
    u, err := s.repo.ByID(id)
    if err != nil {
        httpError(w, 404, "not_found")
        return
    }

    jsonOK(w, u)
}