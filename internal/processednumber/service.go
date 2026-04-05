package processednumber

type Service struct {
	repo           Repository
	sessionNumbers map[string]bool
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:           repo,
		sessionNumbers: make(map[string]bool),
	}
}

func (s *Service) fingerprint(phoneNumber, otpCode string) string {
	return phoneNumber + "|" + otpCode
}

func (s *Service) IsSeen(phoneNumber, otpCode string) (bool, error) {
	fp := s.fingerprint(phoneNumber, otpCode)
	if s.sessionNumbers[fp] {
		return true, nil
	}
	return s.repo.IsSeen(phoneNumber, otpCode)
}

func (s *Service) Add(pn ProcessedNumber) error {
	s.sessionNumbers[s.fingerprint(pn.PhoneNumber, pn.OTPCode)] = true
	return s.repo.Add(pn)
}

func (s *Service) UpdateLastSeen(phoneNumber string) error {
	return s.repo.UpdateLastSeen(phoneNumber)
}

func (s *Service) GetStats() (total int, sessionCount int, firstSeen, lastSeen string, err error) {
	total, _, firstSeen, lastSeen, err = s.repo.GetStats()
	sessionCount = len(s.sessionNumbers)
	return
}
