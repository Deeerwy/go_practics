package service

import "fmt"

type User struct{ ID int64; Email string }
var ErrNotFound = fmt.Errorf("not found")

type UserRepo interface {
	ByEmail(email string) (User, error)
}