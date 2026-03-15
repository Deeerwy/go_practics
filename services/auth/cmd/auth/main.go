package main

import (
	"log"
	"net/http"
	"os"

	authhttp "tech-ip-sem2/services/auth/internal/http"
	"tech-ip-sem2/services/auth/internal/service"
)

func main() {
	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "8081"
	}

	svc := service.NewAuthService()
	handler := authhttp.NewHandler(svc).Router()

	addr := ":" + port
	log.Printf("auth service listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

