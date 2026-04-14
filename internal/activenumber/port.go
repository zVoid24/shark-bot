package activenumber

import "time"

// Repository defines the persistence contract for active numbers.
type Repository interface {
	Insert(an ActiveNumber) error
	GetByUser(userID string) ([]ActiveNumber, error)
	GetByNumber(number, platform string) (*ActiveNumber, error)
	GetAll() ([]ActiveNumber, error)
	DeleteByUser(userID string) error
	DeleteByNumber(number, platform string) error
	UpdateMessageID(userID string, msgID int64) error
	DeleteAll() error
	CleanupExpired(ttl time.Duration) error
	MarkPayoutDone(number, platform string) error
}
