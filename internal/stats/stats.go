package stats

type OtpStat struct {
	Country string `db:"country"`
	Count   int    `db:"count"`
}

type UserOtpStat struct {
	UserID  string `db:"user_id"`
	Country string `db:"country"`
	Count   int    `db:"count"`
}
