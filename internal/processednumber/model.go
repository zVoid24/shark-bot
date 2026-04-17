package processednumber

//test
import "time"

type ProcessedNumber struct {
	PhoneNumber string    `db:"phone_number"`
	FirstSeen   time.Time `db:"first_seen"`
	LastSeen    time.Time `db:"last_seen"`
	OTPCode     string    `db:"otp_code"`
	ServiceName string    `db:"service_name"`
	Posted      bool      `db:"posted"`
}
