package graph

import (
	"context"

	"example.com/pz11-graphql/graph/model"
)

// ─── QUERY ───────────────────────────────────────────────

// Tasks — возвращает список всех задач (авторизация не требуется)
func (r *queryResolver) Tasks(ctx context.Context) ([]*model.Task, error) {
	tasks := r.Store.ListTasks()
	result := make([]*model.Task, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, &model.Task{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			Done:        t.Done,
		})
	}
	return result, nil
}

// Task — возвращает одну задачу по ID (авторизация не требуется).
// Если задача не найдена — возвращает null без ошибки (поведение сохранено).
func (r *queryResolver) Task(ctx context.Context, id string) (*model.Task, error) {
	t, err := r.Store.GetTask(id)
	if err != nil {
		// task not found → возвращаем null, как было раньше
		return nil, nil
	}
	return &model.Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Done:        t.Done,
	}, nil
}

// ─── MUTATION ────────────────────────────────────────────

// CreateTask — создаёт новую задачу (требует авторизации)
func (r *mutationResolver) CreateTask(ctx context.Context, input model.CreateTaskInput) (*model.Task, error) {
	if _, err := GetUserID(ctx); err != nil {
		return nil, err
	}

	t := r.Store.CreateTask(input.Title, input.Description)
	return &model.Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Done:        t.Done,
	}, nil
}

// UpdateTask — обновляет поля задачи (требует авторизации)
func (r *mutationResolver) UpdateTask(ctx context.Context, id string, input model.UpdateTaskInput) (*model.Task, error) {
	if _, err := GetUserID(ctx); err != nil {
		return nil, err
	}

	t, err := r.Store.UpdateTask(id, input.Title, input.Description, input.Done)
	if err != nil {
		return nil, err
	}
	return &model.Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Done:        t.Done,
	}, nil
}

// DeleteTask — удаляет задачу по ID (требует авторизации).
// Если задача не найдена — возвращает false без ошибки (поведение сохранено).
func (r *mutationResolver) DeleteTask(ctx context.Context, id string) (bool, error) {
	if _, err := GetUserID(ctx); err != nil {
		return false, err
	}

	err := r.Store.DeleteTask(id)
	if err != nil {
		// task not found → возвращаем false, как было раньше
		return false, nil
	}
	return true, nil
}

// ─── ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ──────────────────────────────

func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

type queryResolver    struct{ *Resolver }
type mutationResolver struct{ *Resolver }
