package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"example.com/pz11-graphql/graph"
	"example.com/pz11-graphql/middleware"
	"example.com/pz11-graphql/rest"
	"example.com/pz11-graphql/store"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// ── Общее хранилище для REST и GraphQL ──────────────────────────────
	s := store.New()

	// ── GraphQL ─────────────────────────────────────────────────────────
	gqlSrv := handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{
			Resolvers: &graph.Resolver{Store: s},
		}),
	)

	// ── REST ────────────────────────────────────────────────────────────
	restHandler := rest.NewHandler(s)
	restRouter := rest.NewRouter(restHandler)

	// ── Маршруты ────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	// GraphQL Playground — открывается в браузере
	mux.Handle("/", playground.Handler("GraphQL Playground", "/query"))

	// GraphQL endpoint — обёрнут в JWT middleware
	mux.Handle("/query", middleware.AuthMiddleware(gqlSrv))

	// REST endpoints — весь префикс /v1/ уходит в REST-роутер
	mux.Handle("/v1/", restRouter)

	// ── Запуск ──────────────────────────────────────────────────────────
	log.Printf("Сервер запущен на порту %s", port)
	log.Printf("  GraphQL Playground → http://localhost:%s/", port)
	log.Printf("  GraphQL endpoint   → http://localhost:%s/query", port)
	log.Printf("  REST tasks         → http://localhost:%s/v1/tasks", port)

	log.Fatal(http.ListenAndServe(":"+port, mux))
}
