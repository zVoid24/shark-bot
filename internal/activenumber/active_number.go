package activenumber

import "time"

type ActiveNumber struct {
	Number    string    `db:"number"`
	UserID    string    `db:"user_id"`
	Timestamp time.Time `db:"timestamp"`
	MessageID int64     `db:"message_id"`
	Platform  string    `db:"platform"`
	Country   string    `db:"country"`
}
