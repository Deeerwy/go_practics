package httphandler

import (
	"encoding/json"
	"log"
	"net/http"

	"example.com/pz13-rabbit/services/tasks/internal/publisher"
)

// CreateTaskRequest — тело запроса POST /v1/tasks.
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CreateTaskResponse — успешный ответ.
type CreateTaskResponse struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// Handler хранит зависимость от publisher.
type Handler struct {
	pub *publisher.Publisher
}

// New создаёт Handler с инжектированным publisher.
func New(pub *publisher.Publisher) *Handler {
	return &Handler{pub: pub}
}

// CreateTask обрабатывает POST /v1/tasks.
// Логика:
//  1. Принимаем запрос
//  2. Создаём задачу (в учебном варианте — в памяти)
//  3. Публикуем событие (best effort — ошибка только логируется)
//  4. Возвращаем 201 Created
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Читаем тело запроса
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	// Получаем request_id для трассировки
	requestID := r.Header.Get("X-Request-ID")

	// Имитируем создание задачи (в реальном проекте — запись в БД)
	taskID := "t_" + requestID
	log.Printf("[tasks] task created: id=%s title=%s", taskID, req.Title)

	// Публикуем событие ПОСЛЕ успешного создания задачи.
	// Режим: best effort — ошибка публикации только логируется,
	// клиент всё равно получает успешный ответ.
	if err := h.pub.PublishTaskCreated(taskID, requestID); err != nil {
		log.Printf("[tasks] WARNING: publish failed (best effort): %v", err)
		// В strict-режиме здесь был бы: http.Error(..., 500); return
	} else {
		log.Printf("[tasks] event published: task_id=%s", taskID)
	}

	// Отвечаем клиенту
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateTaskResponse{
		TaskID:  taskID,
		Status:  "created",
		Title:   req.Title,
		Message: "task created and event published",
	})
}
