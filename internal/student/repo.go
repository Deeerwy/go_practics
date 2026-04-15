package student

import (
	"database/sql"
	"errors"
)

var ErrStudentNotFound = errors.New("student not found")

type StudentRepository struct {
    db *sql.DB
}

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// Опасный пример — так делать нельзя.
func (r *Repo) UnsafeGetByID(rawID string) (*Student, error) {
	query := "SELECT id, full_name, study_group, email FROM students WHERE id = " + rawID

	row := r.db.QueryRow(query)

	var st Student
	err := row.Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	return &st, nil
}

// Безопасный вариант через параметр.
func (r *Repo) GetByID(id int64) (*Student, error) {
	row := r.db.QueryRow(
		"SELECT id, full_name, study_group, email FROM students WHERE id = $1",
		id,
	)

	var st Student
	err := row.Scan(&st.ID, &st.FullName, &st.StudyGroup, &st.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}

	return &st, nil
}

func (r *Repo) PrepareGetByID() (*sql.Stmt, error) {
	return r.db.Prepare("SELECT id, full_name, study_group, email FROM students WHERE id = $1")
}

func (r *StudentRepository) GetByID(id int) (*Student, error) {
    stmt, err := r.db.Prepare(`
        SELECT id, full_name, study_group, email 
        FROM students 
        WHERE id = $1
    `)
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    row := stmt.QueryRow(id)

    var s Student
    err = row.Scan(&s.ID, &s.FullName, &s.StudyGroup, &s.Email)
    if err != nil {
        return nil, err
    }

    return &s, nil
}