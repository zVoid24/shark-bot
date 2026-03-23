package repository

import (
	"database/sql"
	"shark_bot/internal/processednumber"
	"time"

	"github.com/jmoiron/sqlx"
)

type ProcessedNumberRepo struct {
	db *sqlx.DB
}

func NewProcessedNumberRepo(db *sqlx.DB) *ProcessedNumberRepo {
	return &ProcessedNumberRepo{db: db}
}

func (r *ProcessedNumberRepo) IsSeen(phoneNumber string) (bool, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM processed_numbers WHERE phone_number = $1", phoneNumber)
	return count > 0, err
}

func (r *ProcessedNumberRepo) Add(pn processednumber.ProcessedNumber) error {
	_, err := r.db.Exec(`
		INSERT INTO processed_numbers (phone_number, otp_code, service_name, posted)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (phone_number) DO UPDATE SET 
			last_seen = CURRENT_TIMESTAMP,
			otp_code = EXCLUDED.otp_code,
			service_name = EXCLUDED.service_name
	`, pn.PhoneNumber, pn.OTPCode, pn.ServiceName, pn.Posted)
	return err
}

func (r *ProcessedNumberRepo) UpdateLastSeen(phoneNumber string) error {
	_, err := r.db.Exec("UPDATE processed_numbers SET last_seen = CURRENT_TIMESTAMP WHERE phone_number = $1", phoneNumber)
	return err
}

func (r *ProcessedNumberRepo) GetStats() (total int, sessionCount int, firstSeen, lastSeen string, err error) {
	err = r.db.Get(&total, "SELECT COUNT(*) FROM processed_numbers")
	if err != nil {
		return
	}

	var fs, ls sql.NullTime
	err = r.db.QueryRow("SELECT MIN(first_seen), MAX(last_seen) FROM processed_numbers").Scan(&fs, &ls)
	if err != nil {
		return
	}

	if fs.Valid {
		firstSeen = fs.Time.Format(time.RFC3339)
	}
	if ls.Valid {
		lastSeen = ls.Time.Format(time.RFC3339)
	}

	return
}
