package main

import (
	"log"
	"net/http"
	"os"

	amqpclient "example.com/pz13-rabbit/services/tasks/internal/amqp"
	httphandler "example.com/pz13-rabbit/services/tasks/internal/http"
	"example.com/pz13-rabbit/services/tasks/internal/publisher"
)

func main() {
	// Читаем конфигурацию из переменных окружения
	rabbitURL := getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/")
	queueName := getEnv("QUEUE_NAME", "task_events")
	port := getEnv("TASKS_PORT", "8082")

	// Подключаемся к RabbitMQ
	conn := amqpclient.MustConnect(rabbitURL)
	defer conn.Close()

	// Открываем канал
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("[tasks] channel error: %v", err)
	}
	defer ch.Close()

	// Создаём publisher (объявляет очередь)
	pub, err := publisher.New(ch, queueName)
	if err != nil {
		log.Fatalf("[tasks] publisher init error: %v", err)
	}

	// Регистрируем HTTP-маршруты
	handler := httphandler.New(pub)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status":"ok","service":"tasks"}`))
	})

	mux.HandleFunc("/v1/tasks", handler.CreateTask)

	addr := ":" + port
	log.Printf("[tasks] service started on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[tasks] server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
