package db

import (
	"fmt"
	"log"
	"shark_bot/config"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func GetConnectionString(dbCnf *config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbCnf.Host,
		dbCnf.Port,
		dbCnf.User,
		dbCnf.Password,
		dbCnf.Name,
		dbCnf.SSLMode, // correct usage
	)
}

func NewConnection(dbCnf *config.DatabaseConfig) (*sqlx.DB, error) {
	connStr := GetConnectionString(dbCnf)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// 🔥 Connection Pool Settings (IMPORTANT)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 🔥 Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	return db, nil
}
