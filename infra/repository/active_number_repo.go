package repository

import (
	"database/sql"
	"time"

	"shark_bot/internal/activenumber"

	"github.com/jmoiron/sqlx"
)

// ActiveNumberRepo implements activenumber.Repository.
type ActiveNumberRepo struct {
	db *sqlx.DB
}

func NewActiveNumberRepo(db *sqlx.DB) *ActiveNumberRepo {
	return &ActiveNumberRepo{db: db}
}

func (r *ActiveNumberRepo) Insert(an activenumber.ActiveNumber) error {
	_, err := r.db.Exec(`INSERT INTO active_numbers (number, user_id, timestamp, message_id, platform, country)
		VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (number) DO NOTHING`,
		an.Number, an.UserID, an.Timestamp, an.MessageID, an.Platform, an.Country)
	return err
}

func (r *ActiveNumberRepo) GetByUser(userID string) ([]activenumber.ActiveNumber, error) {
	var ans []activenumber.ActiveNumber
	err := r.db.Select(&ans, `SELECT number, user_id, timestamp, message_id, platform, country
		FROM active_numbers WHERE user_id = $1`, userID)
	return ans, err
}

func (r *ActiveNumberRepo) GetByNumber(number string) (*activenumber.ActiveNumber, error) {
	var an activenumber.ActiveNumber
	err := r.db.Get(&an, `SELECT number, user_id, timestamp, message_id, platform, country
		FROM active_numbers WHERE number = $1`, number)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &an, err
}

func (r *ActiveNumberRepo) GetAll() ([]activenumber.ActiveNumber, error) {
	var ans []activenumber.ActiveNumber
	err := r.db.Select(&ans, `SELECT number, user_id, timestamp, message_id, platform, country FROM active_numbers`)
	return ans, err
}

func (r *ActiveNumberRepo) DeleteByUser(userID string) error {
	_, err := r.db.Exec("DELETE FROM active_numbers WHERE user_id = $1", userID)
	return err
}

func (r *ActiveNumberRepo) DeleteByNumber(number string) error {
	_, err := r.db.Exec("DELETE FROM active_numbers WHERE number = $1", number)
	return err
}

func (r *ActiveNumberRepo) UpdateMessageID(userID string, msgID int64) error {
	_, err := r.db.Exec("UPDATE active_numbers SET message_id = $1 WHERE user_id = $2", msgID, userID)
	return err
}

func (r *ActiveNumberRepo) DeleteAll() error {
	_, err := r.db.Exec("DELETE FROM active_numbers")
	return err
}

func (r *ActiveNumberRepo) CleanupExpired(ttl time.Duration) error {
	cutoff := time.Now().Add(-ttl)
	_, err := r.db.Exec("DELETE FROM active_numbers WHERE timestamp < $1", cutoff)
	return err
}
