package activenumber

import "time"

// Repository defines the persistence contract for active numbers.
type Repository interface {
	Insert(an ActiveNumber) error
	GetByUser(userID string) ([]ActiveNumber, error)
	GetByNumber(number string) (*ActiveNumber, error)
	GetAll() ([]ActiveNumber, error)
	DeleteByUser(userID string) error
	DeleteByNumber(number string) error
	UpdateMessageID(userID string, msgID int64) error
	DeleteAll() error
	CleanupExpired(ttl time.Duration) error
}
