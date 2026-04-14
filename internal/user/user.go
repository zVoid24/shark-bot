package user

type User struct {
	UserID    string  `db:"user_id"`
	FullName  string  `db:"full_name"`
	IsBlocked bool    `db:"is_blocked"`
	Balance   float64 `db:"balance"`
}
