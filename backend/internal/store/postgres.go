package store

import (
	"backend/internal/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Scanner interface {
	Scan(dest ...any) error
}

type PostgresStoreConfig[T models.Identifiable] struct {
	Table      string
	SelectSQL  string
	InsertSQL  string
	Scan       func(sc Scanner) (T, error)
	InsertArgs func(item T) []any
}

type PostgresStore[T models.Identifiable] struct {
	pool   *pgxpool.Pool
	config PostgresStoreConfig[T]
}

func NewPostgresStore[T models.Identifiable](pool *pgxpool.Pool, config PostgresStoreConfig[T]) *PostgresStore[T] {
	return &PostgresStore[T]{pool: pool, config: config}
}

func (s *PostgresStore[T]) GetAll() ([]T, error) {
	rows, err := s.pool.Query(context.Background(), s.config.SelectSQL)
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

func (s *PostgresStore[T]) Get(id int) (T, error) {
	row := s.pool.QueryRow(context.Background(), s.config.SelectSQL+" WHERE id = $1", id)
	item, err := s.config.Scan(row)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("%s %d not found", s.config.Table, id)
	}
	return item, nil
}

func (s *PostgresStore[T]) Create(item T) (T, error) {
	row := s.pool.QueryRow(context.Background(), s.config.InsertSQL, s.config.InsertArgs(item)...)
	created, err := s.config.Scan(row)
	if err != nil {
		var zero T
		return zero, err
	}
	return created, nil
}

func (s *PostgresStore[T]) Delete(id int) error {
	tag, err := s.pool.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s WHERE id = $1", s.config.Table), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s %d not found", s.config.Table, id)
	}
	return nil
}
