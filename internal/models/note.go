package models

type Note struct {
  ID        int64  `db:"id" json:"id"`
  Title     string `db:"title" json:"title"`
  Content   string `db:"content" json:"content"`
  CreatedAt string `db:"created_at" json:"created_at"`
  UpdatedAt string `db:"updated_at" json:"updated_at"`
}
