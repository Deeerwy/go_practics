package main

import (
	"log"
	"net/http"

	"example.com/pz6-web-security/internal/httpapi"
	"example.com/pz6-web-security/internal/store"
)

func main() {
	st := store.New()

	handler, err := httpapi.NewHandler(st)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login", handler.Login)
	mux.HandleFunc("/profile", handler.Profile)
	mux.HandleFunc("/hello", handler.Hello)
	mux.HandleFunc("/comments", handler.Comments) // добавлено

	log.Println("server started on http://localhost:8080")
	log.Println("open http://localhost:8080/login")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

