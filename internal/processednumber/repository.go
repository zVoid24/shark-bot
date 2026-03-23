package processednumber

type Repository interface {
	IsSeen(phoneNumber string) (bool, error)
	Add(pn ProcessedNumber) error
	UpdateLastSeen(phoneNumber string) error
	GetStats() (total int, sessionCount int, firstSeen, lastSeen string, err error)
}
