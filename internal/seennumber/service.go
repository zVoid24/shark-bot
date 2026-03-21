package seennumber

// Service provides seen number business operations via the Repository port.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(userID, number, country string) error {
	return s.repo.Add(userID, number, country)
}

func (s *Service) ResetAll() error {
	return s.repo.ResetAll()
}
