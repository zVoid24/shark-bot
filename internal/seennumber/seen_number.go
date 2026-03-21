package seennumber

type SeenNumber struct {
	UserID  string `db:"user_id"`
	Number  string `db:"number"`
	Country string `db:"country"`
}
