package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

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
	migrations := []string{
		"migrations/001_initial_schema.sql",
		"migrations/002_llm_support.sql",
		"migrations/003_ai_extract.sql",
	}
	for _, path := range migrations {
		ddl, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		// SQLite ignores "duplicate column" errors for ALTER TABLE ADD COLUMN;
		// split on semicolons and execute each statement, ignoring "duplicate column" errors.
		stmts := splitSQL(string(ddl))
		for _, stmt := range stmts {
			if _, err := db.Exec(stmt); err != nil {
				// ignore duplicate column errors from idempotent ALTER TABLE
				if !isDuplicateColumnErr(err) {
					return fmt.Errorf("exec %s: %w", path, err)
				}
			}
		}
	}
	return nil
}

func splitSQL(sql string) []string {
	var stmts []string
	for _, s := range splitSemicolon(sql) {
		s = trimSpace(s)
		if s != "" {
			stmts = append(stmts, s)
		}
	}
	return stmts
}

func splitSemicolon(s string) []string {
	var parts []string
	var buf []byte
	for i := 0; i < len(s); i++ {
		if s[i] == ';' {
			parts = append(parts, string(buf))
			buf = buf[:0]
		} else {
			buf = append(buf, s[i])
		}
	}
	if len(buf) > 0 {
		parts = append(parts, string(buf))
	}
	return parts
}

func trimSpace(s string) string {
	start, end := 0, len(s)-1
	for start <= end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end >= start && (s[end] == ' ' || s[end] == '\t' || s[end] == '\n' || s[end] == '\r') {
		end--
	}
	if start > end {
		return ""
	}
	return s[start : end+1]
}

func isDuplicateColumnErr(err error) bool {
	return err != nil && len(err.Error()) > 0 &&
		(contains(err.Error(), "duplicate column name") || contains(err.Error(), "already exists"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
