package router

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"

    "example.com/pz10-auth/internal/http/middleware"
    "example.com/pz10-auth/internal/platform/config"
    "example.com/pz10-auth/internal/platform/jwt"
    "example.com/pz10-auth/internal/repo"
    "example.com/pz10-auth/internal/core"
)

func Build(cfg config.Config) http.Handler {
    r := chi.NewRouter()

    // DI
    userRepo := repo.NewUserMem()
    accessJWT := jwt.NewHS256(cfg.JWTSecret, 15*time.Minute)
    refreshJWT := jwt.NewHS256(cfg.JWTSecret, 7*24*time.Hour) // 7 дней

    svc := core.NewService(userRepo, accessJWT, refreshJWT)

    // Публичные маршруты
    r.Post("/api/v1/login", svc.LoginHandler)
    r.Post("/api/v1/refresh", svc.RefreshHandler)

    // Защищённые маршруты (access Bearer)
    r.Group(func(priv chi.Router) {
        priv.Use(middleware.AuthN(accessJWT))                // access токен
        priv.Use(middleware.AuthZRoles("admin","user"))
        priv.Get("/api/v1/me", svc.MeHandler)
    })

    // Только для админов
    r.Group(func(admin chi.Router) {
        admin.Use(middleware.AuthN(accessJWT))
        admin.Use(middleware.AuthZRoles("admin"))
        admin.Get("/api/v1/admin/stats", svc.AdminStats)
    })

	// Защищённые маршруты (access Bearer)
r.Group(func(priv chi.Router) {
    priv.Use(middleware.AuthN(accessJWT))                // access токен
    priv.Use(middleware.AuthZRoles("admin","user"))      // базовая RBAC
    priv.Get("/api/v1/me", svc.MeHandler)
    priv.Get("/api/v1/users/{id}", svc.UserHandler)      // <-- ABAC внутри хендлера
})


    return r
}
