package store

import "sync"

type UserProfile struct {
	SessionID string
	Name      string
	CSRFToken string
	Comments  []string // добавлено
}

type Store struct {
	mu    sync.RWMutex
	users map[string]*UserProfile
}

func New() *Store {
	return &Store{
		users: make(map[string]*UserProfile),
	}
}

func (s *Store) Save(profile *UserProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[profile.SessionID] = profile
}

func (s *Store) Get(sessionID string) (*UserProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	profile, ok := s.users[sessionID]
	return profile, ok
}

func (s *Store) UpdateName(sessionID, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	profile, ok := s.users[sessionID]
	if !ok {
		return false
	}
	profile.Name = name
	return true
}

// AddComment добавляет комментарий к профилю пользователя
func (s *Store) AddComment(sessionID, comment string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	profile, ok := s.users[sessionID]
	if !ok {
		return false
	}
	profile.Comments = append(profile.Comments, comment)
	return true
}
