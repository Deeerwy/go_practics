package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	rabbitamqp "pz14/services/tasks/internal/amqp"
	"pz14/services/tasks/internal/jobs"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/google/uuid"
)

// Handler хранит зависимости HTTP-обработчика
type Handler struct {
	ch *amqp.Channel
}

// NewHandler создаёт новый обработчик
func NewHandler(ch *amqp.Channel) *Handler {
	return &Handler{ch: ch}
}

// processTaskRequest — тело входящего запроса
type processTaskRequest struct {
	TaskID string `json:"task_id"`
}

// processTaskResponse — тело ответа клиенту
type processTaskResponse struct {
	Status string `json:"status"`
	TaskID string `json:"task_id"`
}

// ServeHTTP реализует интерфейс http.Handler — роутинг запросов
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию: в учебном варианте просто проверяем наличие токена
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v1/jobs/process-task":
		h.handleProcessTask(w, r)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

// handleProcessTask принимает запрос и ставит задачу в очередь
func (h *Handler) handleProcessTask(w http.ResponseWriter, r *http.Request) {
	var req processTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.TaskID == "" {
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}

	// Формируем сообщение задачи
	job := jobs.TaskJob{
		Job:       "process_task",
		TaskID:    req.TaskID,
		Attempt:   1,                // первая попытка
		MessageID: uuid.NewString(), // уникальный идентификатор сообщения
	}

	if err := rabbitamqp.PublishJob(h.ch, "task_jobs", job); err != nil {
		log.Printf("failed to publish job: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("job published: task_id=%s message_id=%s", job.TaskID, job.MessageID)

	resp := processTaskResponse{
		Status: "accepted",
		TaskID: req.TaskID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("encode response error: %v", err)
	}

	_ = fmt.Sprintf("") // заглушка, чтобы пакет fmt не вызывал ошибку импорта
}
