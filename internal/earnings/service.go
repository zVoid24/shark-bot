package earnings

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(e Earning) error {
	return s.repo.Add(e)
}

func (s *Service) GetByUser(userID string) ([]Earning, error) {
	return s.repo.GetByUser(userID)
}
