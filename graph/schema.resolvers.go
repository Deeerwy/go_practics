package graph

import (
	"context"
	"fmt"

	"example.com/pz11-graphql/graph/model"
	"example.com/pz11-graphql/store"
)

// ─── QUERY ───────────────────────────────────────────────

// Tasks — возвращает список всех задач (авторизация не требуется)
func (r *queryResolver) Tasks(ctx context.Context) ([]*model.Task, error) {
	result := make([]*model.Task, 0, len(store.Tasks))
	for _, t := range store.Tasks {
		result = append(result, &model.Task{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			Done:        t.Done,
		})
	}
	return result, nil
}

// Task — возвращает одну задачу по ID (авторизация не требуется)
func (r *queryResolver) Task(ctx context.Context, id string) (*model.Task, error) {
	for _, t := range store.Tasks {
		if t.ID == id {
			return &model.Task{
				ID:          t.ID,
				Title:       t.Title,
				Description: t.Description,
				Done:        t.Done,
			}, nil
		}
	}
	return nil, nil // клиент получит null без ошибки
}

// ─── MUTATION ────────────────────────────────────────────

// CreateTask — создаёт новую задачу (требует авторизации)
func (r *mutationResolver) CreateTask(ctx context.Context, input model.CreateTaskInput) (*model.Task, error) {
	// Проверяем JWT-токен
	if _, err := GetUserID(ctx); err != nil {
		return nil, err
	}

	id := fmt.Sprintf("t_%03d", len(store.Tasks)+1)
	task := &store.Task{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		Done:        false,
	}
	store.Tasks = append(store.Tasks, task)

	return &model.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Done:        task.Done,
	}, nil
}

// UpdateTask — обновляет поля задачи (требует авторизации)
func (r *mutationResolver) UpdateTask(ctx context.Context, id string, input model.UpdateTaskInput) (*model.Task, error) {
	// Проверяем JWT-токен
	if _, err := GetUserID(ctx); err != nil {
		return nil, err
	}

	for _, t := range store.Tasks {
		if t.ID == id {
			if input.Title != nil {
				t.Title = *input.Title
			}
			if input.Description != nil {
				t.Description = input.Description
			}
			if input.Done != nil {
				t.Done = *input.Done
			}
			return &model.Task{
				ID:          t.ID,
				Title:       t.Title,
				Description: t.Description,
				Done:        t.Done,
			}, nil
		}
	}
	return nil, fmt.Errorf("task with id=%s not found", id)
}

// DeleteTask — удаляет задачу по ID (требует авторизации)
func (r *mutationResolver) DeleteTask(ctx context.Context, id string) (bool, error) {
	// Проверяем JWT-токен
	if _, err := GetUserID(ctx); err != nil {
		return false, err
	}

	for i, t := range store.Tasks {
		if t.ID == id {
			store.Tasks = append(store.Tasks[:i], store.Tasks[i+1:]...)
			return true, nil
		}
	}
	return false, nil // задача не найдена — не ошибка, просто false
}

// ─── ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ──────────────────────────────

func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

type queryResolver    struct{ *Resolver }
type mutationResolver struct{ *Resolver }
