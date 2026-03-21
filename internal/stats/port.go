package stats

// Repository defines the persistence contract for OTP statistics.
type Repository interface {
	IncrOtpStat(country string) error
	IncrUserOtpStat(userID, country string) error
	GetAllOtpStats() ([]OtpStat, error)
	GetUserOtpStats(userID string) ([]UserOtpStat, error)
	ResetAll() error
}
