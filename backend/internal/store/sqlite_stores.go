package store

import (
	"backend/internal/models"
	"database/sql"
)

func NewTopicSQLiteStore(db *sql.DB) *SQLiteStore[models.Topic] {
	return NewSQLiteStore(db, SQLiteStoreConfig[models.Topic]{
		Table:     "topics",
		SelectSQL: "SELECT id, name FROM topics",
		InsertSQL: "INSERT INTO topics (name) VALUES (?)",
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

func NewSourceSQLiteStore(db *sql.DB) *SQLiteStore[models.Source] {
	return NewSQLiteStore(db, SQLiteStoreConfig[models.Source]{
		Table:     "sources",
		SelectSQL: "SELECT id, name, url FROM sources",
		InsertSQL: "INSERT INTO sources (name, url) VALUES (?, ?)",
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

func NewStyleSQLiteStore(db *sql.DB) *SQLiteStore[models.Style] {
	return NewSQLiteStore(db, SQLiteStoreConfig[models.Style]{
		Table:     "styles",
		SelectSQL: "SELECT id, name, prompt FROM styles",
		InsertSQL: "INSERT INTO styles (name, prompt) VALUES (?, ?)",
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

func NewDraftSQLiteStore(db *sql.DB) *SQLiteStore[models.Draft] {
	return NewSQLiteStore(db, SQLiteStoreConfig[models.Draft]{
		Table:     "drafts",
		SelectSQL: "SELECT id, title, content, topic_id, style_id, status FROM drafts",
		InsertSQL: "INSERT INTO drafts (title, content, topic_id, style_id, status) VALUES (?, ?, ?, ?, ?)",
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
