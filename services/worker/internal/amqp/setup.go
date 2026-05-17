package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareQueues(ch *amqp.Channel) error {
	// Сначала объявляем DLQ
	_, err := ch.QueueDeclare(
		"task_jobs_dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Основная очередь — теперь с обоими обязательными аргументами
	args := amqp.Table{
		"x-dead-letter-exchange":    "",              // default exchange
		"x-dead-letter-routing-key": "task_jobs_dlq", // имя DLQ как routing key
	}

	_, err = ch.QueueDeclare(
		"task_jobs",
		true,
		false,
		false,
		false,
		args,
	)
	return err
}
