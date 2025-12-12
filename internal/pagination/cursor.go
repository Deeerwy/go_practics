package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
}

func Encode(c Cursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func Decode(s string) (Cursor, error) {
	var c Cursor
	if s == "" {
		return c, errors.New("empty cursor")
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	if c.CreatedAt.IsZero() || c.ID == 0 {
		return c, errors.New("invalid cursor payload")
	}
	return c, nil
}