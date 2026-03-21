package otp

// OTP represents a parsed one-time password from a Telegram group message.
type OTP struct {
	Code     string
	Number   string
	Platform string
	Country  string
	UserID   string
}
