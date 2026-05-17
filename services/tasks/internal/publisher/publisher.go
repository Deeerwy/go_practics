package publisher

import (
	"context"
	"encoding/json"
	"time"

	"example.com/pz13-rabbit/services/tasks/internal/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher отвечает за публикацию событий в RabbitMQ.
type Publisher struct {
	ch        *amqp.Channel
	queueName string
}

// New создаёт Publisher и объявляет очередь (durable).
func New(ch *amqp.Channel, queueName string) (*Publisher, error) {
	// Объявляем очередь — durable=true, чтобы пережить перезапуск брокера.
	_, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{ch: ch, queueName: queueName}, nil
}

// PublishTaskCreated публикует событие task.created в очередь.
// RequestID — опциональное поле для трассировки.
func (p *Publisher) PublishTaskCreated(taskID, requestID string) error {
	event := events.TaskEvent{
		Event:     "task.created",
		TaskID:    taskID,
		TS:        time.Now().UTC().Format(time.RFC3339),
		RequestID: requestID,
		Producer:  "tasks-service",
		Version:   "1",
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(
		context.Background(),
		"",          // default exchange
		p.queueName, // routing key = имя очереди
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // сообщение переживает перезапуск
			Body:         body,
		},
	)
}
