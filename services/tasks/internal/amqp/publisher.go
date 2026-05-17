package amqp

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishJob сериализует job в JSON и публикует в указанную очередь
func PublishJob(ch *amqp.Channel, queue string, job any) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		context.Background(),
		"",    // exchange: пустой — используем default exchange
		queue, // routing key: имя очереди
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // сообщение переживёт перезапуск брокера
			Body:         body,
		},
	)
}
	