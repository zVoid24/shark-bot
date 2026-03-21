package user

// Repository defines the persistence contract for users.
type Repository interface {
	EnsureUser(userID, fullName string) (isNew bool, err error)
	IsBlocked(userID string) (bool, error)
	BlockUser(userID string) error
	UnblockUser(userID string) error
	UnblockAll() error
	GetAllUserIDs() ([]string, error)
	GetBlockedUsers() ([]string, error)
	GetUser(userID string) (*User, error)
}
