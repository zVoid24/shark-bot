package number

// Service provides platform number business operations via the Repository port.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetPlatforms() ([]string, error) {
	return s.repo.GetPlatforms()
}

func (s *Service) GetCountries(platform string) ([]string, error) {
	return s.repo.GetCountries(platform)
}

func (s *Service) CountAvailable(platform, country string) (int, error) {
	return s.repo.CountAvailable(platform, country)
}

func (s *Service) GetNumbers(platform, country, userID string, excludeNums []string, limit int) ([]string, error) {
	return s.repo.GetNumbers(platform, country, userID, excludeNums, limit)
}

func (s *Service) GetNextNumber(platform, country, excludeNum string) (string, error) {
	return s.repo.GetNextNumber(platform, country, excludeNum)
}

func (s *Service) DeleteByPlatformCountry(platform, country string) error {
	return s.repo.DeleteByPlatformCountry(platform, country)
}

func (s *Service) DeleteByNumber(number string) error {
	return s.repo.DeleteByNumber(number)
}

func (s *Service) BulkInsert(platform, country string, numbers []string) (int, error) {
	return s.repo.BulkInsert(platform, country, numbers)
}

func (s *Service) GetPlatformForNumber(num string) (*PlatformNumber, error) {
	return s.repo.GetPlatformForNumber(num)
}
