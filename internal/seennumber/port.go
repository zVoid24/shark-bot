package seennumber

// Repository defines the persistence contract for seen numbers.
type Repository interface {
	Add(userID, number, country string) error
	ResetAll() error
}
