package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	users map[string]User
}

func (r stubRepo) ByEmail(email string) (User, error) {
	u, ok := r.users[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func TestService_FindIDByEmail(t *testing.T) {
	svc := New(stubRepo{
		users: map[string]User{
			"a@b.com": {ID: 42, Email: "a@b.com"},
		},
	})

	t.Run("found", func(t *testing.T) {
		id, err := svc.FindIDByEmail("a@b.com")
		require.NoError(t, err)
		assert.Equal(t, int64(42), id)
	})

	t.Run("not_found", func(t *testing.T) {
		id, err := svc.FindIDByEmail("missing@b.com")
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Equal(t, int64(0), id)
	})
}