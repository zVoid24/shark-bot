package earnings

type Repository interface {
	Add(e Earning) error
	GetByUser(userID string) ([]Earning, error)
}
