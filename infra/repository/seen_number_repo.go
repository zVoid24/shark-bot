package repository

import (
	"github.com/jmoiron/sqlx"
)

// SeenNumberRepo implements seennumber.Repository.
type SeenNumberRepo struct {
	db *sqlx.DB
}

func NewSeenNumberRepo(db *sqlx.DB) *SeenNumberRepo {
	return &SeenNumberRepo{db: db}
}

func (r *SeenNumberRepo) Add(userID, number, country string) error {
	_, err := r.db.Exec(`INSERT INTO seen_numbers (user_id, number, country) VALUES ($1, $2, $3)`,
		userID, number, country)
	return err
}

func (r *SeenNumberRepo) ResetAll() error {
	_, err := r.db.Exec("DELETE FROM seen_numbers")
	return err
}
