package repo

import (
	"errors"

	"example.com/pz10-auth/internal/core"
	"golang.org/x/crypto/bcrypt"
)

// Внутренний тип для хранения в памяти
type UserRecord struct {
	ID       int64
	Email    string
	Role     string
	Password string // bcrypt-хэш
}

type UserMem struct {
	users []UserRecord
}


// Конструктор с заранее захэшированными пользователями
func NewUserMem() *UserMem {
	// для примера: пароль "secret"
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	return &UserMem{
		users: []UserRecord{
			{ID: 1, Email: "admin@example.com", Role: "admin", Password: string(hash)},
			{ID: 2, Email: "user@example.com", Role: "user", Password: string(hash)},
		},
	}
}

// Реализация интерфейса core.userRepo
func (m *UserMem) CheckPassword(email, pass string) (*core.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pass)) == nil {
				// возвращаем доменную модель core.User
				return &core.User{
					ID:    u.ID,
					Email: u.Email,
					Role:  u.Role,
				}, nil
			}
			return nil, errors.New("invalid password")
		}
	}
	return nil, errors.New("user not found")
}
func (r *UserMem) ByID(id int64) (*core.User, error) {
    for _, u := range r.users {
        if u.ID == id {
            return &core.User{ID: u.ID, Email: u.Email, Role: u.Role}, nil
        }
    }
    return nil, errors.New("not found")
}