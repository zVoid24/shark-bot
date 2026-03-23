package settings

// Repository defines the persistence contract for application settings.
type Repository interface {
	Get(key string) (string, error)
	Set(key, value string) error
	GetRemovePolicy(platform, country string) bool
	SetRemovePolicy(platform, country, status string) error
	GetNumberLimit(platform, country string) int
	SetNumberLimit(platform, country string, limit int) error
}
