package store

import (
	"backend/internal/models"
	"context"
	"database/sql"
	"fmt"
)

type Scanner interface {
	Scan(dest ...any) error
}

type SQLiteStoreConfig[T models.Identifiable] struct {
	Table      string
	SelectSQL  string
	InsertSQL  string
	Scan       func(sc Scanner) (T, error)
	InsertArgs func(item T) []any
}

type SQLiteStore[T models.Identifiable] struct {
	db     *sql.DB
	config SQLiteStoreConfig[T]
}

func NewSQLiteStore[T models.Identifiable](db *sql.DB, config SQLiteStoreConfig[T]) *SQLiteStore[T] {
	return &SQLiteStore[T]{db: db, config: config}
}

func (s *SQLiteStore[T]) GetAll() ([]T, error) {
	rows, err := s.db.QueryContext(context.Background(), s.config.SelectSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []T
	for rows.Next() {
		item, err := s.config.Scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SQLiteStore[T]) Get(id int) (T, error) {
	row := s.db.QueryRowContext(context.Background(), s.config.SelectSQL+" WHERE id = ?", id)
	item, err := s.config.Scan(row)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("%s %d not found", s.config.Table, id)
	}
	return item, nil
}

func (s *SQLiteStore[T]) Create(item T) (T, error) {
	result, err := s.db.ExecContext(context.Background(), s.config.InsertSQL, s.config.InsertArgs(item)...)
	if err != nil {
		var zero T
		return zero, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		var zero T
		return zero, err
	}

	row := s.db.QueryRowContext(context.Background(), s.config.SelectSQL+" WHERE id = ?", lastID)
	created, err := s.config.Scan(row)
	if err != nil {
		var zero T
		return zero, err
	}
	return created, nil
}

func (s *SQLiteStore[T]) Delete(id int) error {
	result, err := s.db.ExecContext(context.Background(), fmt.Sprintf("DELETE FROM %s WHERE id = ?", s.config.Table), id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%s %d not found", s.config.Table, id)
	}
	return nil
}
