package repository

import (
	"shark_bot/internal/stats"

	"github.com/jmoiron/sqlx"
)

// StatsRepo implements stats.Repository.
type StatsRepo struct {
	db *sqlx.DB
}

func NewStatsRepo(db *sqlx.DB) *StatsRepo {
	return &StatsRepo{db: db}
}

func (r *StatsRepo) IncrOtpStat(country string) error {
	_, err := r.db.Exec(`INSERT INTO otp_stats (country, count) VALUES ($1, 1)
		ON CONFLICT (country) DO UPDATE SET count = otp_stats.count + 1`, country)
	return err
}

func (r *StatsRepo) IncrUserOtpStat(userID, country string) error {
	_, err := r.db.Exec(`INSERT INTO user_otp_stats (user_id, country, count) VALUES ($1, $2, 1)
		ON CONFLICT (user_id, country) DO UPDATE SET count = user_otp_stats.count + 1`, userID, country)
	return err
}

func (r *StatsRepo) GetAllOtpStats() ([]stats.OtpStat, error) {
	var s []stats.OtpStat
	err := r.db.Select(&s, "SELECT country, count FROM otp_stats ORDER BY count DESC")
	return s, err
}

func (r *StatsRepo) GetUserOtpStats(userID string) ([]stats.UserOtpStat, error) {
	var s []stats.UserOtpStat
	err := r.db.Select(&s, "SELECT user_id, country, count FROM user_otp_stats WHERE user_id = $1 ORDER BY count DESC", userID)
	return s, err
}

func (r *StatsRepo) ResetAll() error {
	_, err := r.db.Exec("DELETE FROM otp_stats")
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM user_otp_stats")
	return err
}
