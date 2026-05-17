package store

import "sync"

// ProcessedStore хранит идентификаторы уже обработанных сообщений.
// Использует sync.Mutex для безопасности при параллельном доступе.
type ProcessedStore struct {
	mu    sync.Mutex
	items map[string]bool
}

// NewProcessedStore создаёт новое хранилище обработанных сообщений
func NewProcessedStore() *ProcessedStore {
	return &ProcessedStore{
		items: make(map[string]bool),
	}
}

// Exists возвращает true, если сообщение с данным ID уже обрабатывалось
func (s *ProcessedStore) Exists(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.items[id]
}

// MarkDone помечает сообщение как успешно обработанное
func (s *ProcessedStore) MarkDone(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[id] = true
}
