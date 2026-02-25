package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := runMigration(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migration: %w", err)
	}

	return db, nil
}

func runMigration(db *sql.DB) error {
	ddl, err := os.ReadFile("migrations/001_initial_schema.sql")
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}
	if _, err := db.Exec(string(ddl)); err != nil {
		return fmt.Errorf("exec migration: %w", err)
	}
	return nil
}
