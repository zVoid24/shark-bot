package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	"shark_bot/config"
	"shark_bot/infra/db"

	_ "modernc.org/sqlite"
)

type sqliteUser struct {
	UserID    string
	FullName  string
	IsBlocked int64
}

func main() {
	sqlitePath := flag.String("sqlite", "bot_data.db", "Path to legacy SQLite database file")
	flag.Parse()

	cnf := config.Load()

	pgDB, err := db.NewConnection(&cnf.Database)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}
	defer pgDB.Close()

	sqliteDB, err := sql.Open("sqlite", *sqlitePath)
	if err != nil {
		log.Fatalf("sqlite open failed: %v", err)
	}
	defer sqliteDB.Close()

	if err := sqliteDB.Ping(); err != nil {
		log.Fatalf("sqlite ping failed: %v", err)
	}

	rows, err := sqliteDB.Query(`
		SELECT
			CAST(user_id AS TEXT) AS user_id,
			COALESCE(full_name, '') AS full_name,
			COALESCE(is_blocked, 0) AS is_blocked
		FROM users
	`)
	if err != nil {
		log.Fatalf("reading sqlite users failed: %v", err)
	}
	defer rows.Close()

	tx, err := pgDB.Beginx()
	if err != nil {
		log.Fatalf("postgres tx begin failed: %v", err)
	}

	stmt, err := tx.Preparex(`
		INSERT INTO users (user_id, full_name, is_blocked)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET full_name = EXCLUDED.full_name,
		    is_blocked = EXCLUDED.is_blocked
	`)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("postgres prepare failed: %v", err)
	}
	defer stmt.Close()

	var total int
	for rows.Next() {
		var u sqliteUser
		if err := rows.Scan(&u.UserID, &u.FullName, &u.IsBlocked); err != nil {
			_ = tx.Rollback()
			log.Fatalf("sqlite row scan failed: %v", err)
		}

		u.UserID = strings.TrimSpace(u.UserID)
		if u.UserID == "" {
			continue
		}

		isBlocked := u.IsBlocked != 0
		if _, err := stmt.Exec(u.UserID, u.FullName, isBlocked); err != nil {
			_ = tx.Rollback()
			log.Fatalf("postgres upsert failed for user_id=%s: %v", u.UserID, err)
		}
		total++
	}

	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		log.Fatalf("sqlite rows error: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("postgres commit failed: %v", err)
	}

	fmt.Printf("Imported %d users from %s into PostgreSQL\n", total, *sqlitePath)
}
