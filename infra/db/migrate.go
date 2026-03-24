package db

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"shark_bot/pkg/logger"

	"github.com/jmoiron/sqlx"
)

//go:embed migrations
var migrationsFS embed.FS

var migrateLog = logger.New("db.migrate")

// Migrate runs all up.sql files from the embedded migrations directory in order.
// Each sub-folder is named NNN_<table>/ containing up.sql and down.sql.
// Already-applied migrations are idempotent (all DDL uses IF NOT EXISTS).
func Migrate(db *sqlx.DB) error {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("reading migrations dir: %w", err)
	}

	// Sort by folder name so they run in 000001, 000002 ... order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	applied := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		upPath := "migrations/" + e.Name() + "/up.sql"
		sql, err := migrationsFS.ReadFile(upPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", upPath, err)
		}
		if _, err := db.Exec(string(sql)); err != nil {
			return fmt.Errorf("applying %s: %w", e.Name(), err)
		}
		applied++
	}

	// Guard: at least one migration must have run
	if applied == 0 {
		return errors.New("no migration files found")
	}

	migrateLog.Info("migrations complete", "count", applied)
	return nil
}
