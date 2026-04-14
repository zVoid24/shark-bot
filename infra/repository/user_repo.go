package repository

import (
	"database/sql"
	"shark_bot/internal/user"

	"github.com/jmoiron/sqlx"
)

// UserRepo implements user.Repository.
type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) EnsureUser(userID, fullName string) (isNew bool, err error) {
	var existing string
	err = r.db.Get(&existing, "SELECT user_id FROM users WHERE user_id = $1", userID)
	if err == sql.ErrNoRows {
		_, err = r.db.Exec("INSERT INTO users (user_id, full_name) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, fullName)
		return true, err
	}
	return false, err
}

func (r *UserRepo) IsBlocked(userID string) (bool, error) {
	var blocked bool
	err := r.db.Get(&blocked, "SELECT is_blocked FROM users WHERE user_id = $1", userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return blocked, err
}

func (r *UserRepo) BlockUser(userID string) error {
	_, err := r.db.Exec(`INSERT INTO users (user_id, full_name, is_blocked) VALUES ($1, '', TRUE)
		ON CONFLICT (user_id) DO UPDATE SET is_blocked = TRUE`, userID)
	return err
}

func (r *UserRepo) UnblockUser(userID string) error {
	_, err := r.db.Exec("UPDATE users SET is_blocked = FALSE WHERE user_id = $1", userID)
	return err
}

func (r *UserRepo) UnblockAll() error {
	_, err := r.db.Exec("UPDATE users SET is_blocked = FALSE")
	return err
}

func (r *UserRepo) GetAllUserIDs() ([]string, error) {
	var ids []string
	err := r.db.Select(&ids, "SELECT user_id FROM users")
	return ids, err
}

func (r *UserRepo) GetBlockedUsers() ([]string, error) {
	var ids []string
	err := r.db.Select(&ids, "SELECT user_id FROM users WHERE is_blocked = TRUE")
	return ids, err
}

func (r *UserRepo) GetUnblockedUserIDs() ([]string, error) {
	var ids []string
	err := r.db.Select(&ids, "SELECT user_id FROM users WHERE is_blocked = FALSE")
	return ids, err
}

func (r *UserRepo) GetUser(userID string) (*user.User, error) {
	var u user.User
	err := r.db.Get(&u, "SELECT user_id, full_name, is_blocked, balance FROM users WHERE user_id = $1", userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepo) AddBalance(userID string, amount float64) error {
	_, err := r.db.Exec("UPDATE users SET balance = balance + $1 WHERE user_id = $2", amount, userID)
	return err
}

func (r *UserRepo) DeductBalance(userID string, amount float64) error {
	_, err := r.db.Exec("UPDATE users SET balance = balance - $1 WHERE user_id = $2", amount, userID)
	return err
}
