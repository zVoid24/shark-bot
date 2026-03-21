package repository

import (
	"github.com/jmoiron/sqlx"
)

// AdminRepo implements admin.Repository.
type AdminRepo struct {
	db *sqlx.DB
}

func NewAdminRepo(db *sqlx.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) IsAdmin(userID string) (bool, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM admins WHERE user_id = $1", userID)
	return count > 0, err
}

func (r *AdminRepo) GetAll() ([]string, error) {
	var ids []string
	err := r.db.Select(&ids, "SELECT user_id FROM admins")
	return ids, err
}

func (r *AdminRepo) Add(userID string) error {
	_, err := r.db.Exec("INSERT INTO admins (user_id) VALUES ($1) ON CONFLICT DO NOTHING", userID)
	return err
}

func (r *AdminRepo) Remove(userID string) error {
	_, err := r.db.Exec("DELETE FROM admins WHERE user_id = $1", userID)
	return err
}

func (r *AdminRepo) Count() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM admins")
	return count, err
}

// SeedOwners inserts initial owner IDs as admins.
func (r *AdminRepo) SeedOwners(ownerIDs []string) error {
	for _, uid := range ownerIDs {
		if uid == "" {
			continue
		}
		_, err := r.db.Exec("INSERT INTO admins (user_id) VALUES ($1) ON CONFLICT DO NOTHING", uid)
		if err != nil {
			return err
		}
	}
	return nil
}
