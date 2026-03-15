package service

import "errors"

const (
	demoUsername = "student"
	demoPassword = "student"
	demoToken    = "demo-token"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(username, password string) (string, error) {
	if username == demoUsername && password == demoPassword {
		return demoToken, nil
	}
	return "", ErrInvalidCredentials
}

func (s *AuthService) Verify(token string) (bool, string) {
	if token == demoToken {
		return true, demoUsername
	}
	return false, ""
}

