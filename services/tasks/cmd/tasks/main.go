package main

import (
	"log"
	"net/http"
	"os"

	rabbitamqp "pz14/services/tasks/internal/amqp"
	httphandler "pz14/services/tasks/internal/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// Читаем URL RabbitMQ из переменной окружения
	rabbitURL := os.Getenv("RABBIT_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	// Подключаемся к RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	// Объявляем очереди при старте сервиса
	if err := rabbitamqp.DeclareQueues(ch); err != nil {
		log.Fatalf("failed to declare queues: %v", err)
	}

	log.Println("queues declared: task_jobs, task_jobs_dlq")

	handler := httphandler.NewHandler(ch)

	addr := ":8082"
	log.Printf("tasks service listening on %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("http server error: %v", err)
	}
}
