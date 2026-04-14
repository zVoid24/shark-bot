package earnings

import "time"

type Earning struct {
	ID        int64     `db:"id"`
	UserID    string    `db:"user_id"`
	Amount    float64   `db:"amount"`
	Source    string    `db:"source"`
	Timestamp time.Time `db:"timestamp"`
}
