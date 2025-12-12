package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/go_practics/internal/model"
	"example.com/go_practics/internal/pagination"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Create(ctx context.Context, title, content string) (model.Note, error) {
	var n model.Note
	err := r.pool.QueryRow(ctx, qInsert, title, content).Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt)
	return n, err
}

func (r *Repo) Get(ctx context.Context, id int64) (model.Note, error) {
	var n model.Note
	err := r.pool.QueryRow(ctx, qGetByID, id).Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Note{}, pgx.ErrNoRows
	}
	return n, err
}

func (r *Repo) List(ctx context.Context, q string, limit int, cursor *pagination.Cursor) ([]model.Note, *pagination.Cursor, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var rows pgx.Rows
	var err error

	if cursor == nil {
		rows, err = r.pool.Query(ctx, qListFirst, q, limit)
	} else {
		rows, err = r.pool.Query(ctx, qListAfter, q, cursor.CreatedAt, cursor.ID, limit)
	}
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	out := make([]model.Note, 0, limit)
	for rows.Next() {
		var n model.Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
			return nil, nil, err
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// next cursor = last item
	if len(out) == 0 {
		return out, nil, nil
	}
	last := out[len(out)-1]
	next := &pagination.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	return out, next, nil
}

func (r *Repo) UpdateTx(ctx context.Context, id int64, title, content string) (model.Note, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return model.Note{}, err
	}
	defer tx.Rollback(ctx)

	var n model.Note
	err = tx.QueryRow(ctx, qUpdate, id, title, content).Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt)
	if err != nil {
		return model.Note{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Note{}, err
	}
	return n, nil
}

func (r *Repo) Delete(ctx context.Context, id int64) error {
	ct, err := r.pool.Exec(ctx, qDelete, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repo) BatchTitles(ctx context.Context, ids []int64) (map[int64]string, error) {
	rows, err := r.pool.Query(ctx, qBatchTitles, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[int64]string, len(ids))
	for rows.Next() {
		var id int64
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, err
		}
		res[id] = title
	}
	return res, rows.Err()
}