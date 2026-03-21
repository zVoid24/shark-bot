package repository

import (
	"fmt"
	"strings"

	"shark_bot/internal/number"

	"github.com/jmoiron/sqlx"
)

// NumberRepo implements number.Repository.
type NumberRepo struct {
	db *sqlx.DB
}

func NewNumberRepo(db *sqlx.DB) *NumberRepo {
	return &NumberRepo{db: db}
}

func (r *NumberRepo) GetPlatforms() ([]string, error) {
	var platforms []string
	err := r.db.Select(&platforms, "SELECT DISTINCT platform FROM platform_numbers ORDER BY platform")
	return platforms, err
}

func (r *NumberRepo) GetCountries(platform string) ([]string, error) {
	var countries []string
	err := r.db.Select(&countries, "SELECT DISTINCT country FROM platform_numbers WHERE platform = $1 ORDER BY country", platform)
	return countries, err
}

func (r *NumberRepo) CountAvailable(platform, country string) (int, error) {
	var count int
	var err error
	if platform != "" && country != "" {
		err = r.db.Get(&count, `SELECT COUNT(*) FROM platform_numbers
			WHERE platform = $1 AND country = $2
			AND number NOT IN (SELECT number FROM active_numbers)`, platform, country)
	} else if platform != "" {
		err = r.db.Get(&count, `SELECT COUNT(*) FROM platform_numbers
			WHERE platform = $1
			AND number NOT IN (SELECT number FROM active_numbers)`, platform)
	}
	return count, err
}

func (r *NumberRepo) GetNumbers(platform, country, userID string, excludeNums []string, limit int) ([]string, error) {
	q1 := `SELECT number FROM platform_numbers
		WHERE platform = $1 AND country = $2
		AND number NOT IN (SELECT number FROM active_numbers)
		AND number NOT IN (SELECT number FROM seen_numbers WHERE user_id = $3)
		ORDER BY RANDOM() LIMIT $4`
	var numbers []string
	err := r.db.Select(&numbers, q1, platform, country, userID, limit)
	if err != nil {
		return nil, err
	}
	if len(numbers) > 0 {
		return numbers, nil
	}

	if len(excludeNums) > 0 {
		placeholders := make([]string, len(excludeNums))
		args := []interface{}{platform, country}
		for i, n := range excludeNums {
			placeholders[i] = fmt.Sprintf("$%d", i+3)
			args = append(args, n)
		}
		args = append(args, limit)
		q2 := fmt.Sprintf(`SELECT number FROM platform_numbers
			WHERE platform = $1 AND country = $2
			AND number NOT IN (SELECT number FROM active_numbers)
			AND number NOT IN (%s)
			ORDER BY RANDOM() LIMIT $%d`, strings.Join(placeholders, ","), len(args))
		err = r.db.Select(&numbers, q2, args...)
	} else {
		err = r.db.Select(&numbers, `SELECT number FROM platform_numbers
			WHERE platform = $1 AND country = $2
			AND number NOT IN (SELECT number FROM active_numbers)
			ORDER BY RANDOM() LIMIT $3`, platform, country, limit)
	}
	return numbers, err
}

func (r *NumberRepo) GetNextNumber(platform, country, excludeNum string) (string, error) {
	var n string
	err := r.db.Get(&n, `SELECT number FROM platform_numbers
		WHERE platform = $1 AND country = $2 AND number != $3
		AND number NOT IN (SELECT number FROM active_numbers)
		ORDER BY RANDOM() LIMIT 1`, platform, country, excludeNum)
	return n, err
}

func (r *NumberRepo) DeleteByPlatformCountry(platform, country string) error {
	_, err := r.db.Exec("DELETE FROM platform_numbers WHERE platform = $1 AND country = $2", platform, country)
	return err
}

func (r *NumberRepo) DeleteByNumber(num string) error {
	_, err := r.db.Exec("DELETE FROM platform_numbers WHERE number = $1", num)
	return err
}

func (r *NumberRepo) BulkInsert(platform, country string, numbers []string) (int, error) {
	count := 0
	for _, num := range numbers {
		num = strings.TrimSpace(num)
		if num == "" {
			continue
		}
		res, err := r.db.Exec(`INSERT INTO platform_numbers (platform, country, number)
			VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, platform, country, num)
		if err != nil {
			continue
		}
		rows, _ := res.RowsAffected()
		count += int(rows)
	}
	return count, nil
}

func (r *NumberRepo) GetPlatformForNumber(num string) (*number.PlatformNumber, error) {
	var pn number.PlatformNumber
	err := r.db.Get(&pn, "SELECT id, platform, country, number FROM platform_numbers WHERE number = $1 LIMIT 1", num)
	return &pn, err
}
