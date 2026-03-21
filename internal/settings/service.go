package settings

// Service provides settings business operations via the Repository port.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(key string) (string, error) {
	return s.repo.Get(key)
}

func (s *Service) Set(key, value string) error {
	return s.repo.Set(key, value)
}

func (s *Service) GetGroupLink() string {
	return s.repo.GetGroupLink()
}

func (s *Service) GetRemovePolicy(platform, country string) bool {
	return s.repo.GetRemovePolicy(platform, country)
}

func (s *Service) SetRemovePolicy(platform, country, status string) error {
	return s.repo.SetRemovePolicy(platform, country, status)
}

func (s *Service) GetNumberLimit(platform, country string) int {
	return s.repo.GetNumberLimit(platform, country)
}

func (s *Service) SetNumberLimit(platform, country string, limit int) error {
	return s.repo.SetNumberLimit(platform, country, limit)
}
