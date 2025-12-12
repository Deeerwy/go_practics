package postgres

const (
	qInsert = `
INSERT INTO notes (title, content)
VALUES ($1, $2)
RETURNING id, title, content, created_at;
`

	qGetByID = `
SELECT id, title, content, created_at
FROM notes
WHERE id = $1;
`

	// first page (optional search)
	qListFirst = `
SELECT id, title, content, created_at
FROM notes
WHERE ($1 = '' OR to_tsvector('simple', title) @@ plainto_tsquery('simple', $1))
ORDER BY created_at DESC, id DESC
LIMIT $2;
`

	// next pages (keyset)
	qListAfter = `
SELECT id, title, content, created_at
FROM notes
WHERE
  ($1 = '' OR to_tsvector('simple', title) @@ plainto_tsquery('simple', $1))
  AND (created_at, id) < ($2, $3)
ORDER BY created_at DESC, id DESC
LIMIT $4;
`

	qUpdate = `
UPDATE notes
SET title = $2, content = $3
WHERE id = $1
RETURNING id, title, content, created_at;
`

	qDelete = `
DELETE FROM notes WHERE id = $1;
`

	// batching to avoid N+1
	qBatchTitles = `
SELECT id, title
FROM notes
WHERE id = ANY($1);
`
)