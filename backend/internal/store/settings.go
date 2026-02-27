package store

import (
	"context"
	"database/sql"
)

type SettingsStore struct {
	db *sql.DB
}

func NewSettingsStore(db *sql.DB) *SettingsStore {
	return &SettingsStore{db: db}
}

func (s *SettingsStore) GetAll() (map[string]string, error) {
	rows, err := s.db.QueryContext(context.Background(), "SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, rows.Err()
}

func (s *SettingsStore) Set(key, value string) error {
	_, err := s.db.ExecContext(context.Background(),
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value",
		key, value)
	return err
}
