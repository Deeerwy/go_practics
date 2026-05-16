package store

import (
	"errors"
	"fmt"
	"sync"
)


type Task struct {
	ID          string
	Title       string
	Description *string
	Done        bool
}

// Store — потокобезопасное in-memory хранилище задач.
type Store struct {
	mu      sync.RWMutex
	tasks   map[string]*Task
	counter int
}

// New создаёт хранилище и заполняет его тестовыми данными.
func New() *Store {
	s := &Store{tasks: make(map[string]*Task)}
	s.tasks["t_001"] = &Task{ID: "t_001", Title: "Первая задача", Description: strPtr("Учебный пример"), Done: false}
	s.tasks["t_002"] = &Task{ID: "t_002", Title: "Вторая задача", Description: strPtr("GraphQL API"), Done: true}
	s.counter = 2
	return s
}

// ListTasks возвращает все задачи (порядок не гарантирован — map).
func (s *Store) ListTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}

// GetTask возвращает задачу по ID или ошибку, если не найдена.
func (s *Store) GetTask(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.tasks[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	return t, nil
}

// CreateTask создаёт новую задачу и возвращает её.
func (s *Store) CreateTask(title string, description *string) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	id := fmt.Sprintf("t_%03d", s.counter)
	t := &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Done:        false,
	}
	s.tasks[id] = t
	return t
}

// UpdateTask обновляет только переданные (не nil) поля задачи.
func (s *Store) UpdateTask(id string, title, description *string, done *bool) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	if title != nil {
		t.Title = *title
	}
	if description != nil {
		t.Description = description
	}
	if done != nil {
		t.Done = *done
	}
	return t, nil
}

// DeleteTask удаляет задачу по ID.
func (s *Store) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return errors.New("task not found")
	}
	delete(s.tasks, id)
	return nil
}

// strPtr — вспомогательная функция, оставлена как была в оригинале.
func strPtr(s string) *string {
	return &s
}
