package admin

// Repository defines the persistence contract for admins.
type Repository interface {
	IsAdmin(userID string) (bool, error)
	GetAll() ([]string, error)
	Add(userID string) error
	Remove(userID string) error
	Count() (int, error)
	SeedOwners(ownerIDs []string) error
}
