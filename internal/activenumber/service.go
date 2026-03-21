package activenumber

import "time"

// Service provides active number business operations via the Repository port.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Insert(an ActiveNumber) error {
	return s.repo.Insert(an)
}

func (s *Service) GetByUser(userID string) ([]ActiveNumber, error) {
	return s.repo.GetByUser(userID)
}

func (s *Service) GetByNumber(number string) (*ActiveNumber, error) {
	return s.repo.GetByNumber(number)
}

func (s *Service) GetAll() ([]ActiveNumber, error) {
	return s.repo.GetAll()
}

func (s *Service) DeleteByUser(userID string) error {
	return s.repo.DeleteByUser(userID)
}

func (s *Service) DeleteByNumber(number string) error {
	return s.repo.DeleteByNumber(number)
}

func (s *Service) UpdateMessageID(userID string, msgID int64) error {
	return s.repo.UpdateMessageID(userID, msgID)
}

func (s *Service) DeleteAll() error {
	return s.repo.DeleteAll()
}

func (s *Service) CleanupExpired(ttl time.Duration) error {
	return s.repo.CleanupExpired(ttl)
}
