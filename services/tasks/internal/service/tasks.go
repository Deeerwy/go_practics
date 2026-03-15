package service

import (
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("task not found")
)

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Done        bool   `json:"done"`
}

type Store struct {
	mu    sync.RWMutex
	tasks map[string]Task
	seq   int
}

func NewStore() *Store {
	return &Store{
		tasks: make(map[string]Task),
	}
}

func (s *Store) Create(t Task) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	t.ID = s.nextID()
	s.tasks[t.ID] = t
	return t
}

func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		res = append(res, t)
	}
	return res
}

func (s *Store) Get(id string) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return t, nil
}

func (s *Store) Update(id string, update Task) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	if update.Title != "" {
		cur.Title = update.Title
	}
	if update.Description != "" {
		cur.Description = update.Description
	}
	if update.DueDate != "" {
		cur.DueDate = update.DueDate
	}
	cur.Done = update.Done
	s.tasks[id] = cur
	return cur, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}

func (s *Store) nextID() string {
	return "t_" + itoa(s.seq)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

