package store

import (
	"backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewTopicPostgresStore(pool *pgxpool.Pool) *PostgresStore[models.Topic] {
	return NewPostgresStore(pool, PostgresStoreConfig[models.Topic]{
		Table:     "topics",
		SelectSQL: "SELECT id, name FROM topics",
		InsertSQL: "INSERT INTO topics (name) VALUES ($1) RETURNING id, name",
		Scan: func(sc Scanner) (models.Topic, error) {
			var t models.Topic
			err := sc.Scan(&t.Id, &t.Name)
			return t, err
		},
		InsertArgs: func(t models.Topic) []any {
			return []any{t.Name}
		},
	})
}

func NewSourcePostgresStore(pool *pgxpool.Pool) *PostgresStore[models.Source] {
	return NewPostgresStore(pool, PostgresStoreConfig[models.Source]{
		Table:     "sources",
		SelectSQL: "SELECT id, name, url FROM sources",
		InsertSQL: "INSERT INTO sources (name, url) VALUES ($1, $2) RETURNING id, name, url",
		Scan: func(sc Scanner) (models.Source, error) {
			var s models.Source
			err := sc.Scan(&s.Id, &s.Name, &s.Url)
			return s, err
		},
		InsertArgs: func(s models.Source) []any {
			return []any{s.Name, s.Url}
		},
	})
}

func NewStylePostgresStore(pool *pgxpool.Pool) *PostgresStore[models.Style] {
	return NewPostgresStore(pool, PostgresStoreConfig[models.Style]{
		Table:     "styles",
		SelectSQL: "SELECT id, name, prompt FROM styles",
		InsertSQL: "INSERT INTO styles (name, prompt) VALUES ($1, $2) RETURNING id, name, prompt",
		Scan: func(sc Scanner) (models.Style, error) {
			var s models.Style
			err := sc.Scan(&s.Id, &s.Name, &s.Prompt)
			return s, err
		},
		InsertArgs: func(s models.Style) []any {
			return []any{s.Name, s.Prompt}
		},
	})
}

func NewDraftPostgresStore(pool *pgxpool.Pool) *PostgresStore[models.Draft] {
	return NewPostgresStore(pool, PostgresStoreConfig[models.Draft]{
		Table:     "drafts",
		SelectSQL: "SELECT id, title, content, topic_id, style_id, status FROM drafts",
		InsertSQL: "INSERT INTO drafts (title, content, topic_id, style_id, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, title, content, topic_id, style_id, status",
		Scan: func(sc Scanner) (models.Draft, error) {
			var d models.Draft
			err := sc.Scan(&d.Id, &d.Title, &d.Content, &d.TopicId, &d.StyleId, &d.Status)
			return d, err
		},
		InsertArgs: func(d models.Draft) []any {
			return []any{d.Title, d.Content, d.TopicId, d.StyleId, d.Status}
		},
	})
}
