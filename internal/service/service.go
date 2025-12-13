package service

import (
    "context"
    "example.com/pz16-integration/internal/models"
    "example.com/pz16-integration/internal/repo"
)

type Service struct{ Notes repo.NoteRepo }

func (s Service) Create(ctx context.Context, n *models.Note) error {
    // можно добавить валидацию
    // проверка на пустой заголовок
    if n.Title == "" {
        return ValidationError{Field: "title", Message: "title cannot be empty"}
    }
    return s.Notes.Create(ctx, n)
}

func (s Service) Get(ctx context.Context, id int64) (models.Note, error) {
    return s.Notes.Get(ctx, id)
}

func (s Service) Update(ctx context.Context, id int64, title, content string) error {
    // Валидация входных данных
    if title == "" {
        return ValidationError{Field: "title", Message: "title cannot be empty"}
    }
    return s.Notes.Update(ctx, id, title, content)
}

func (s Service) Delete(ctx context.Context, id int64) error {
    return s.Notes.Delete(ctx, id)
}

func (s Service) List(ctx context.Context, limit, offset int64) ([]models.Note, error) {
    // Устанавливаем разумные ограничения
    if limit <= 0 {
        limit = 10 // дефолтное значение
    }
    if limit > 100 {
        limit = 100 // максимальное значение
    }
    if offset < 0 {
        offset = 0
    }
    
    return s.Notes.List(ctx, limit, offset)
}

// ListAll возвращает все заметки без пагинации
func (s Service) ListAll(ctx context.Context) ([]models.Note, error) {
    return s.Notes.ListAll(ctx)
}

// ValidationError для ошибок валидации
type ValidationError struct {
    Field   string
    Message string
}

func (v ValidationError) Error() string {
    return v.Field + ": " + v.Message
}