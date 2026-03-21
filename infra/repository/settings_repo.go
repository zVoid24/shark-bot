package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// SettingsRepo implements settings.Repository.
type SettingsRepo struct {
	db *sqlx.DB
}

func NewSettingsRepo(db *sqlx.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) Get(key string) (string, error) {
	var val string
	err := r.db.Get(&val, "SELECT value FROM settings WHERE key = $1", key)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

func (r *SettingsRepo) Set(key, value string) error {
	_, err := r.db.Exec(`INSERT INTO settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $2`, key, value)
	return err
}

func (r *SettingsRepo) GetGroupLink() string {
	val, _ := r.Get("group_link")
	if val == "" {
		return "https://t.me/tgwscreatebdotp"
	}
	return val
}

func (r *SettingsRepo) GetRemovePolicy(platform, country string) bool {
	key := "remove_policy::" + strings.ToLower(platform) + "::" + strings.ToLower(country)
	val, _ := r.Get(key)
	return val == "on"
}

func (r *SettingsRepo) SetRemovePolicy(platform, country, status string) error {
	key := "remove_policy::" + strings.ToLower(platform) + "::" + strings.ToLower(country)
	return r.Set(key, status)
}

func (r *SettingsRepo) GetNumberLimit(platform, country string) int {
	key := "limit::" + strings.ToLower(platform) + "::" + strings.ToLower(country)
	val, _ := r.Get(key)
	if val == "" {
		return 2
	}
	n := 0
	fmt.Sscanf(val, "%d", &n)
	if n <= 0 {
		return 2
	}
	return n
}

func (r *SettingsRepo) SetNumberLimit(platform, country string, limit int) error {
	key := "limit::" + strings.ToLower(platform) + "::" + strings.ToLower(country)
	return r.Set(key, fmt.Sprintf("%d", limit))
}
