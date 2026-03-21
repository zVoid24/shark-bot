package number

type PlatformNumber struct {
	ID       int    `db:"id"`
	Platform string `db:"platform"`
	Country  string `db:"country"`
	Number   string `db:"number"`
}
