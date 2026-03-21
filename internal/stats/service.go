package stats

// Service provides stats business operations via the Repository port.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) IncrOtpStat(country string) error {
	return s.repo.IncrOtpStat(country)
}

func (s *Service) IncrUserOtpStat(userID, country string) error {
	return s.repo.IncrUserOtpStat(userID, country)
}

func (s *Service) GetAllOtpStats() ([]OtpStat, error) {
	return s.repo.GetAllOtpStats()
}

func (s *Service) GetUserOtpStats(userID string) ([]UserOtpStat, error) {
	return s.repo.GetUserOtpStats(userID)
}

func (s *Service) ResetAll() error {
	return s.repo.ResetAll()
}
