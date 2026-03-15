package main

import (
	"log"
	"net/http"
	"os"
	"time"

	taskhttp "tech-ip-sem2/services/tasks/internal/http"
	"tech-ip-sem2/services/tasks/internal/client/authclient"
	"tech-ip-sem2/services/tasks/internal/service"
)

func main() {
	port := os.Getenv("TASKS_PORT")
	if port == "" {
		port = "8082"
	}
	authBase := os.Getenv("AUTH_BASE_URL")
	if authBase == "" {
		authBase = "http://localhost:8081"
	}

	store := service.NewStore()
	authCli := authclient.New(authBase, 3*time.Second)
	handler := taskhttp.NewHandler(store, authCli).Router()

	addr := ":" + port
	log.Printf("tasks service listening on %s (auth=%s)", addr, authBase)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

