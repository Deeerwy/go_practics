package repo

import (
  "sync"
  "example.com/notes-api/internal/core"
)

type NoteRepoMem struct {
  mu    sync.Mutex
  notes map[int64]*core.Note
  next  int64
}

func NewNoteRepoMem() *NoteRepoMem {
  return &NoteRepoMem{notes: make(map[int64]*core.Note)}
}

func (r *NoteRepoMem) Create(n core.Note) (int64, error) {
  r.mu.Lock(); defer r.mu.Unlock()
  r.next++
  n.ID = r.next
  r.notes[n.ID] = &n
  return n.ID, nil
}
