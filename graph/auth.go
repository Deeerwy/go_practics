package graph

import (
    "context"
    "fmt"

    "example.com/pz11-graphql/middleware"
)

// GetUserID извлекает userID из контекста запроса.
// Возвращает ошибку если пользователь не аутентифицирован.
func GetUserID(ctx context.Context) (string, error) {
    userID, ok := ctx.Value(middleware.UserIDKey).(string)
    if !ok || userID == "" {
        return "", fmt.Errorf("unauthorized: токен не предоставлен или недействителен")
    }
    return userID, nil
}

// IsAuthenticated проверяет наличие аутентификации без возврата ошибки.
func IsAuthenticated(ctx context.Context) bool {
    userID, _ := ctx.Value(middleware.UserIDKey).(string)
    return userID != ""
}
