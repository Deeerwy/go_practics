package repo

import (
	"context"
	"database/sql"
	"errors"
	"example.com/pz16-integration/internal/models"
)

type NoteRepo struct{ DB *sql.DB }

func (r NoteRepo) ListAll(ctx context.Context) ([]models.Note, error) {
	panic("unimplemented")
}

func (r NoteRepo) Create(ctx context.Context, n *models.Note) error {
	return r.DB.QueryRowContext(ctx,
		`INSERT INTO notes(title, content) VALUES($1,$2) RETURNING id`,
		n.Title, n.Content,
	).Scan(&n.ID)
}

func (r NoteRepo) Get(ctx context.Context, id int64) (models.Note, error) {
	var n models.Note
	err := r.DB.QueryRowContext(ctx,
		`SELECT id, title, content, created_at, updated_at FROM notes WHERE id=$1`, id,
	).Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return models.Note{}, errors.New("not found")
	}
	return n, err
}

func (r NoteRepo) Update(ctx context.Context, id int64, title, content string) error {
	result, err := r.DB.ExecContext(ctx,
		`UPDATE notes SET title=$1, content=$2, updated_at=NOW() WHERE id=$3`,
		title, content, id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("not found")
	}

	return nil
}

func (r NoteRepo) Delete(ctx context.Context, id int64) error {
	result, err := r.DB.ExecContext(ctx,
		`DELETE FROM notes WHERE id=$1`,
		id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("not found")
	}

	return nil
}

func (r NoteRepo) List(ctx context.Context, limit, offset int64) ([]models.Note, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, title, content, created_at, updated_at 
         FROM notes 
         ORDER BY created_at DESC 
         LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var n models.Note
		err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}
