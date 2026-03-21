package db

import (
	"fmt"
	"time"

	"shark_bot/config"
	"shark_bot/pkg/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var connLog = logger.New("db.connection")

func GetConnectionString(dbCnf *config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbCnf.Host,
		dbCnf.Port,
		dbCnf.User,
		dbCnf.Password,
		dbCnf.Name,
		dbCnf.SSLMode,
	)
}

func NewConnection(dbCnf *config.DatabaseConfig) (*sqlx.DB, error) {
	connStr := GetConnectionString(dbCnf)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	connLog.Info("connected to PostgreSQL")
	return db, nil
}
