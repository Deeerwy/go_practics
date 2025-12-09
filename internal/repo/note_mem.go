package repo

import (
  "strings"
  "sync"
  "time"
  "example.com/notes-api/internal/core"
)

// NoteRepoMem — простое in-memory хранилище заметок.
type NoteRepoMem struct {
	mu     sync.RWMutex
	notes  map[int64]core.Note
	nextID int64
}

// NewNoteRepoMem создаёт новый репозиторий с пустым хранилищем.
func NewNoteRepoMem() *NoteRepoMem {
	return &NoteRepoMem{
		notes:  make(map[int64]core.Note),
		nextID: 0,
	}
}

// Create добавляет новую заметку и возвращает её ID.
func (r *NoteRepoMem) Create(n core.Note) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	n.ID = r.nextID

	// Если CreatedAt не установлен — проставим сейчас
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	// UpdatedAt при создании оставляем nil

	r.notes[n.ID] = n

	return n.ID, nil
}

// List возвращает список заметок с пагинацией и фильтром по title.
func (r *NoteRepoMem) List(page, limit int, q string) ([]core.Note, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	q = strings.TrimSpace(q)
	qLower := strings.ToLower(q)

	// Сначала собираем все заметки и фильтруем по q
	var filtered []core.Note
	for _, n := range r.notes {
		if qLower == "" {
			filtered = append(filtered, n)
			continue
		}
		if strings.Contains(strings.ToLower(n.Title), qLower) {
			filtered = append(filtered, n)
		}
	}

	total := len(filtered)

	start := (page - 1) * limit
	if start >= total {
		return []core.Note{}, total, nil
	}

	end := start + limit
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// Get возвращает заметку по ID.
func (r *NoteRepoMem) Get(id int64) (core.Note, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n, ok := r.notes[id]
	return n, ok
}

// Update частично обновляет заметку по ID, используя NoteUpdate.
// Возвращает обновлённую заметку и флаг, была ли она найдена.
func (r *NoteRepoMem) Update(id int64, upd core.NoteUpdate) (core.Note, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.notes[id]
	if !ok {
		return core.Note{}, false
	}

	if upd.Title != nil {
		n.Title = *upd.Title
	}
	if upd.Content != nil {
		n.Content = *upd.Content
	}

	now := time.Now()
	n.UpdatedAt = &now

	r.notes[id] = n
	return n, true
}

// Delete удаляет заметку по ID.
// Возвращает true, если заметка была найдена и удалена.
func (r *NoteRepoMem) Delete(id int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.notes[id]; !ok {
		return false
	}
	delete(r.notes, id)
	return true
}
