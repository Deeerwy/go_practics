package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// Хранилище задач в памяти
var tasks = []Task{
	{ID: 1, Title: "Изучить Kubernetes", Done: false},
	{ID: 2, Title: "Написать манифесты", Done: true},
	{ID: 3, Title: "Применить kubectl apply", Done: false},
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{
		Status:  "ok",
		Service: "tasks",
	})
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func main() {
	// Читаем конфигурацию из переменных окружения (переданных через ConfigMap)
	port := os.Getenv("TASKS_PORT")
	if port == "" {
		port = "8082"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	authBaseURL := os.Getenv("AUTH_BASE_URL")

	log.Printf("[INFO] Сервис tasks запускается")
	log.Printf("[INFO] LOG_LEVEL=%s", logLevel)
	log.Printf("[INFO] AUTH_BASE_URL=%s", authBaseURL)
	log.Printf("[INFO] Слушаем порт :%s", port)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/tasks", tasksHandler)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("[ERROR] Сервер завершился с ошибкой: %v", err)
	}
}
