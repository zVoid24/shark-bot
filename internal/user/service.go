package user

// Service provides user business operations via the Repository port.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) EnsureUser(userID, fullName string) (bool, error) {
	return s.repo.EnsureUser(userID, fullName)
}

func (s *Service) IsBlocked(userID string) (bool, error) {
	return s.repo.IsBlocked(userID)
}

func (s *Service) BlockUser(userID string) error {
	return s.repo.BlockUser(userID)
}

func (s *Service) UnblockUser(userID string) error {
	return s.repo.UnblockUser(userID)
}

func (s *Service) UnblockAll() error {
	return s.repo.UnblockAll()
}

func (s *Service) GetAllUserIDs() ([]string, error) {
	return s.repo.GetAllUserIDs()
}

func (s *Service) GetBlockedUsers() ([]string, error) {
	return s.repo.GetBlockedUsers()
}

func (s *Service) GetUnblockedUserIDs() ([]string, error) {
	return s.repo.GetUnblockedUserIDs()
}

func (s *Service) GetUser(userID string) (*User, error) {
	return s.repo.GetUser(userID)
}

func (s *Service) AddBalance(userID string, amount float64) error {
	return s.repo.AddBalance(userID, amount)
}

func (s *Service) DeductBalance(userID string, amount float64) error {
	return s.repo.DeductBalance(userID, amount)
}
