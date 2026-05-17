package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	rabbitamqp "pz14/services/worker/internal/amqp"
	"pz14/services/worker/internal/jobs"
	"pz14/services/worker/internal/store"

	amqp "github.com/rabbitmq/amqp091-go"
)

// maxAttempts — максимальное число попыток обработки задачи
const maxAttempts = 3

// Consumer читает задачи из очереди и обрабатывает их
type Consumer struct {
	ch        *amqp.Channel
	processed *store.ProcessedStore
}

// NewConsumer создаёт нового consumer
func NewConsumer(ch *amqp.Channel, processed *store.ProcessedStore) *Consumer {
	return &Consumer{
		ch:        ch,
		processed: processed,
	}
}

// Run запускает цикл чтения сообщений из очереди task_jobs
func (c *Consumer) Run() error {
	// Устанавливаем prefetch: worker берёт по одному сообщению за раз
	if err := c.ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("qos error: %w", err)
	}

	msgs, err := c.ch.Consume(
		"task_jobs",
		"worker-consumer", // consumer tag
		false,             // auto-ack: false — подтверждаем вручную
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume error: %w", err)
	}

	log.Println("worker started, waiting for messages...")

	for d := range msgs {
		c.handle(d)
	}

	return nil
}

// handle обрабатывает одно сообщение
func (c *Consumer) handle(d amqp.Delivery) {
	var job jobs.TaskJob

	if err := json.Unmarshal(d.Body, &job); err != nil {
		log.Printf("failed to unmarshal message: %v — sending to DLQ", err)
		c.sendToDLQ(job)
		_ = d.Ack(false)
		return
	}

	log.Printf(
		"received job: task_id=%s message_id=%s attempt=%d",
		job.TaskID, job.MessageID, job.Attempt,
	)

	// ── Идемпотентная проверка ──────────────────────────────────────────────
	// Если это сообщение уже обрабатывалось успешно — пропускаем
	if c.processed.Exists(job.MessageID) {
		log.Printf("duplicate message detected, skipping: message_id=%s", job.MessageID)
		_ = d.Ack(false)
		return
	}

	// ── Выполнение задачи ───────────────────────────────────────────────────
	err := processTask(job)

	if err != nil {
		log.Printf(
			"processing error: task_id=%s attempt=%d error=%v",
			job.TaskID, job.Attempt, err,
		)

		job.Attempt++

		if job.Attempt <= maxAttempts {
			// Есть ещё попытки — публикуем заново в основную очередь
			log.Printf(
				"retrying: task_id=%s next_attempt=%d",
				job.TaskID, job.Attempt,
			)

			if pubErr := rabbitamqp.PublishJob(c.ch, "task_jobs", job); pubErr != nil {
				log.Printf("retry publish error: %v", pubErr)
			}
		} else {
			// Попытки исчерпаны — отправляем в DLQ
			log.Printf(
				"max attempts exceeded, sending to DLQ: task_id=%s message_id=%s",
				job.TaskID, job.MessageID,
			)
			c.sendToDLQ(job)
		}

		// Подтверждаем оригинальное сообщение — оно больше не нужно в очереди,
		// т.к. мы либо опубликовали retry, либо отправили в DLQ
		_ = d.Ack(false)
		return
	}

	// ── Успешная обработка ──────────────────────────────────────────────────
	log.Printf(
		"job processed successfully: task_id=%s message_id=%s",
		job.TaskID, job.MessageID,
	)

	// Помечаем message_id как обработанный для идемпотентности
	c.processed.MarkDone(job.MessageID)

	_ = d.Ack(false)
}

// sendToDLQ публикует сообщение в очередь проблемных сообщений
func (c *Consumer) sendToDLQ(job jobs.TaskJob) {
	if err := rabbitamqp.PublishJob(c.ch, "task_jobs_dlq", job); err != nil {
		log.Printf("DLQ publish error: %v", err)
	}
}

// processTask имитирует тяжёлую бизнес-операцию.
// Задача t_fail всегда завершается ошибкой — для проверки DLQ.
func processTask(job jobs.TaskJob) error {
	log.Printf("processing task: task_id=%s (simulating 2s work...)", job.TaskID)
	time.Sleep(2 * time.Second)

	if job.TaskID == "t_fail" {
		return fmt.Errorf("simulated processing error for task_id=%s", job.TaskID)
	}

	return nil
}
