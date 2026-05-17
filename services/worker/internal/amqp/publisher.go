package amqp

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishJob публикует job в указанную очередь
func PublishJob(ch *amqp.Channel, queue string, job any) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		context.Background(),
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
