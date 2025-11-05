package main

import (
	"context"
	"database/sql"
	"time"
)

// Task — модель для сканирования результатов SELECT
type Task struct {
	ID        int
	Title     string
	Done      bool
	CreatedAt time.Time
}

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo { return &Repo{DB: db} }

// CreateTask — параметризованный INSERT с возвратом id
func (r *Repo) CreateTask(ctx context.Context, title string) (int, error) {
	var id int
	const q = `INSERT INTO tasks (title) VALUES ($1) RETURNING id;`
	err := r.DB.QueryRowContext(ctx, q, title).Scan(&id)
	return id, err
}

// ListTasks — базовый SELECT всех задач (демо для занятия)
func (r *Repo) ListTasks(ctx context.Context) ([]Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks ORDER BY id;`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}


func (r *Repo) ListDone(ctx context.Context, done bool) ([]Task, error) {
    const q = `SELECT id, title, done, created_at FROM tasks WHERE done = $1 ORDER BY id;`
    rows, err := r.DB.QueryContext(ctx, q, done)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []Task
    for rows.Next() {
        var t Task
        if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
            return nil, err
        }
        out = append(out, t)
    }
    return out, rows.Err()
}

func (r *Repo) FindByID(ctx context.Context, id int) (*Task, error) {
    const q = `SELECT id, title, done, created_at FROM tasks WHERE id = $1;`
    row := r.DB.QueryRowContext(ctx, q, id)

    var t Task
    if err := row.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
        return nil, err
    }
    return &t, nil
}

func (r *Repo) CreateMany(ctx context.Context, titles []string) error {
    tx, err := r.DB.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    stmt, err := tx.PrepareContext(ctx, `INSERT INTO tasks (title) VALUES ($1)`)
    if err != nil {
        tx.Rollback()
        return err
    }
    defer stmt.Close()

    for _, title := range titles {
        if _, err := stmt.ExecContext(ctx, title); err != nil {
            tx.Rollback()
            return err
        }
    }

    return tx.Commit()
}


