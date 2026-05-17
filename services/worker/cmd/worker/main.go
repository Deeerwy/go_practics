package main

import (
	"log"
	"os"

	rabbitamqp "pz14/services/worker/internal/amqp"
	"pz14/services/worker/internal/consumer"
	"pz14/services/worker/internal/store"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitURL := os.Getenv("RABBIT_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

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

	// Объявляем очереди
	if err := rabbitamqp.DeclareQueues(ch); err != nil {
		log.Fatalf("failed to declare queues: %v", err)
	}

	log.Println("queues declared: task_jobs, task_jobs_dlq")

	// Инициализируем хранилище обработанных сообщений
	processed := store.NewProcessedStore()

	// Запускаем consumer
	c := consumer.NewConsumer(ch, processed)
	if err := c.Run(); err != nil {
		log.Fatalf("consumer error: %v", err)
	}
}
