package repository

import (
	"shark_bot/internal/earnings"

	"github.com/jmoiron/sqlx"
)

type EarningsRepo struct {
	db *sqlx.DB
}

func NewEarningsRepo(db *sqlx.DB) *EarningsRepo {
	return &EarningsRepo{db: db}
}

func (r *EarningsRepo) Add(e earnings.Earning) error {
	_, err := r.db.Exec(`INSERT INTO earnings_log (user_id, amount, source, timestamp)
		VALUES ($1, $2, $3, $4)`,
		e.UserID, e.Amount, e.Source, e.Timestamp)
	return err
}

func (r *EarningsRepo) GetByUser(userID string) ([]earnings.Earning, error) {
	var es []earnings.Earning
	err := r.db.Select(&es, "SELECT id, user_id, amount, source, timestamp FROM earnings_log WHERE user_id = $1", userID)
	return es, err
}
