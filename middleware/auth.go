package middleware

import (
    "context"
    "net/http"
    "strings"
)

// Ключ для хранения userID в контексте запроса
type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware извлекает JWT токен из заголовка Authorization.
// В учебной версии токен не верифицируется криптографически —
// просто проверяется его наличие и извлекается значение.
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")

        if authHeader != "" {
            // Ожидаем формат: "Bearer <token>"
            parts := strings.SplitN(authHeader, " ", 2)
            if len(parts) == 2 && parts[0] == "Bearer" {
                token := parts[1]

                // В учебной версии токен используется как userID напрямую.
                // В продакшене здесь была бы верификация подписи JWT.
                ctx := context.WithValue(r.Context(), UserIDKey, token)
                r = r.WithContext(ctx)
            }
        }

        next.ServeHTTP(w, r)
    })
}
