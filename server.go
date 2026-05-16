package main

import (
    "log"
    "net/http"
    "os"

    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/playground"
    "example.com/pz11-graphql/graph"
    "example.com/pz11-graphql/middleware" 
)

const defaultPort = "8080"

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = defaultPort
    }

    srv := handler.NewDefaultServer(
        graph.NewExecutableSchema(graph.Config{
            Resolvers: &graph.Resolver{},
        }),
    )

    http.Handle("/", playground.Handler("GraphQL Playground", "/query"))

    // Оборачиваем endpoint в middleware авторизации
    http.Handle("/query", middleware.AuthMiddleware(srv))

    log.Printf("Сервер запущен: http://localhost:%s/", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
