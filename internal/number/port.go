package number

// Repository defines the persistence contract for platform numbers.
type Repository interface {
	GetPlatforms() ([]string, error)
	GetCountries(platform string) ([]string, error)
	CountAvailable(platform, country string) (int, error)
	GetNumbers(platform, country, userID string, excludeNums []string, limit int) ([]string, error)
	GetNextNumber(platform, country, excludeNum string) (string, error)
	DeleteByPlatformCountry(platform, country string) error
	DeleteByNumber(number string) error
	BulkInsert(platform, country string, numbers []string) (int, error)
	GetPlatformForNumber(number string) (*PlatformNumber, error)
}
