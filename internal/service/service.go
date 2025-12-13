package service

type Service struct{ repo UserRepo }

func New(repo UserRepo) *Service { return &Service{repo: repo} }

// FindIDByEmail возвращает ID пользователя по email или 0/ошибку.
func (s *Service) FindIDByEmail(email string) (int64, error) {
	u, err := s.repo.ByEmail(email)
	if err != nil { return 0, err }
	return u.ID, nil
}
