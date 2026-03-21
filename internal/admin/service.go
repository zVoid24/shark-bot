package admin

// Service provides admin business operations via the Repository port.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) IsAdmin(userID string) (bool, error) {
	return s.repo.IsAdmin(userID)
}

func (s *Service) GetAll() ([]string, error) {
	return s.repo.GetAll()
}

func (s *Service) Add(userID string) error {
	return s.repo.Add(userID)
}

func (s *Service) Remove(userID string) error {
	return s.repo.Remove(userID)
}

func (s *Service) Count() (int, error) {
	return s.repo.Count()
}

func (s *Service) SeedOwners(ownerIDs []string) error {
	return s.repo.SeedOwners(ownerIDs)
}
