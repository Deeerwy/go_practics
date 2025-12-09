// Package main Notes API server.
//
// @title           Notes API
// @version         1.0
// @description     Учебный REST API для заметок (CRUD).
// @contact.name    Backend Course
// @contact.email   example@university.ru
// @BasePath        /api/v1
package main

import (
  "log"
  "net/http"
  "example.com/notes-api/internal/http"
  "example.com/notes-api/internal/repo"
  httpSwagger "github.com/swaggo/http-swagger"
  _ "example.com/notes-api/docs"
  "example.com/notes-api/docs"
)

func main() {
	    // Настроим SwaggerInfo
    docs.SwaggerInfo.Host = "localhost:8085"
    docs.SwaggerInfo.Schemes = []string{"http"}
    docs.SwaggerInfo.BasePath = "/api/v1"
	mem := repo.NewNoteRepoMem()

	r := httpx.NewRouter(mem)

	r.Get("/docs/*", httpSwagger.WrapHandler)

	log.Println("Server started at :8085")
	log.Fatal(http.ListenAndServe(":8085", r))
}
