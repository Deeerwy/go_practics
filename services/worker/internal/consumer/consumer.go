package consumer

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TaskEvent — структура получаемого события.
type TaskEvent struct {
	Event     string `json:"event"`
	TaskID    string `json:"task_id"`
	TS        string `json:"ts"`
	RequestID string `json:"request_id,omitempty"`
	Producer  string `json:"producer,omitempty"`
	Version   string `json:"version,omitempty"`
}

// Consumer читает сообщения из очереди и подтверждает обработку.
type Consumer struct {
	ch        *amqp.Channel
	queueName string
	prefetch  int
}

// New создаёт Consumer, объявляет очередь и выставляет prefetch.
func New(ch *amqp.Channel, queueName string, prefetch int) (*Consumer, error) {
	// Объявляем очередь с теми же параметрами, что и producer.
	// Если очередь уже существует — просто убеждаемся, что параметры совпадают.
	_, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Prefetch = 1: брокер не даст worker'у следующее сообщение,
	// пока не получит ack за текущее. Предотвращает перегрузку.
	if err := ch.Qos(prefetch, 0, false); err != nil {
		return nil, err
	}

	return &Consumer{ch: ch, queueName: queueName, prefetch: prefetch}, nil
}

// Run запускает бесконечный цикл чтения сообщений.
func (c *Consumer) Run() error {
	msgs, err := c.ch.Consume(
		c.queueName,
		"",    // consumer tag (auto)
		false, // autoAck = false → ручное подтверждение
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("[worker] started, prefetch=%d, queue=%s", c.prefetch, c.queueName)
	log.Println("[worker] waiting for messages...")

	for d := range msgs {
		c.handle(d)
	}

	return nil
}

// handle обрабатывает одно сообщение.
func (c *Consumer) handle(d amqp.Delivery) {
	var ev TaskEvent
	if err := json.Unmarshal(d.Body, &ev); err != nil {
		log.Printf("[worker] ERROR bad message body: %v | raw: %s", err, d.Body)
		// Nack(multiple=false, requeue=false) — отбрасываем невалидное сообщение
		_ = d.Nack(false, false)
		return
	}

	// Основная "бизнес-логика" worker'а
	log.Printf(
		"[worker] received event=%s task_id=%s ts=%s request_id=%s producer=%s",
		ev.Event, ev.TaskID, ev.TS, ev.RequestID, ev.Producer,
	)

	// Подтверждаем успешную обработку.
	// После Ack брокер удаляет сообщение из очереди.
	if err := d.Ack(false); err != nil {
		log.Printf("[worker] ERROR ack failed: %v", err)
	}
}
