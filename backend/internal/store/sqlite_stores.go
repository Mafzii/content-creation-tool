package store

import (
	"backend/internal/models"
	"context"
	"database/sql"
)

func NewTopicSQLiteStore(db *sql.DB) *SQLiteStore[models.Topic] {
	return NewSQLiteStore(db, SQLiteStoreConfig[models.Topic]{
		Table:     "topics",
		SelectSQL: "SELECT id, name, description, keywords FROM topics",
		InsertSQL: "INSERT INTO topics (name, description, keywords) VALUES (?, ?, ?)",
		UpdateSQL: "UPDATE topics SET name=?, description=?, keywords=?",
		Scan: func(sc Scanner) (models.Topic, error) {
			var t models.Topic
			err := sc.Scan(&t.Id, &t.Name, &t.Description, &t.Keywords)
			return t, err
		},
		InsertArgs: func(t models.Topic) []any {
			return []any{t.Name, t.Description, t.Keywords}
		},
		UpdateArgs: func(t models.Topic) []any {
			return []any{t.Name, t.Description, t.Keywords}
		},
	})
}

func NewSourceSQLiteStore(db *sql.DB) *SQLiteStore[models.Source] {
	return NewSQLiteStore(db, SQLiteStoreConfig[models.Source]{
		Table:     "sources",
		SelectSQL: "SELECT id, name, type, raw, content, status, extract_mode, topic_id FROM sources",
		InsertSQL: "INSERT INTO sources (name, url, type, raw, content, status, extract_mode, topic_id) VALUES (?, '', ?, ?, ?, ?, ?, ?)",
		UpdateSQL: "UPDATE sources SET name=?, type=?, raw=?, content=?, status=?, extract_mode=?, topic_id=?",
		Scan: func(sc Scanner) (models.Source, error) {
			var s models.Source
			err := sc.Scan(&s.Id, &s.Name, &s.Type, &s.Raw, &s.Content, &s.Status, &s.ExtractMode, &s.TopicId)
			return s, err
		},
		InsertArgs: func(s models.Source) []any {
			typ := s.Type
			if typ == "" {
				typ = "text"
			}
			status := s.Status
			if status == "" {
				status = "ready"
			}
			extractMode := s.ExtractMode
			if extractMode == "" {
				extractMode = "standard"
			}
			return []any{s.Name, typ, s.Raw, s.Content, status, extractMode, s.TopicId}
		},
		UpdateArgs: func(s models.Source) []any {
			typ := s.Type
			if typ == "" {
				typ = "text"
			}
			status := s.Status
			if status == "" {
				status = "ready"
			}
			extractMode := s.ExtractMode
			if extractMode == "" {
				extractMode = "standard"
			}
			return []any{s.Name, typ, s.Raw, s.Content, status, extractMode, s.TopicId}
		},
	})
}

func NewStyleSQLiteStore(db *sql.DB) *SQLiteStore[models.Style] {
	return NewSQLiteStore(db, SQLiteStoreConfig[models.Style]{
		Table:     "styles",
		SelectSQL: "SELECT id, name, prompt, tone, example FROM styles",
		InsertSQL: "INSERT INTO styles (name, prompt, tone, example) VALUES (?, ?, ?, ?)",
		UpdateSQL: "UPDATE styles SET name=?, prompt=?, tone=?, example=?",
		Scan: func(sc Scanner) (models.Style, error) {
			var s models.Style
			err := sc.Scan(&s.Id, &s.Name, &s.Prompt, &s.Tone, &s.Example)
			return s, err
		},
		InsertArgs: func(s models.Style) []any {
			return []any{s.Name, s.Prompt, s.Tone, s.Example}
		},
		UpdateArgs: func(s models.Style) []any {
			return []any{s.Name, s.Prompt, s.Tone, s.Example}
		},
	})
}

// DraftSQLiteStore is a custom store for drafts that also manages draft_sources.
type DraftSQLiteStore struct {
	db *sql.DB
}

func NewDraftSQLiteStore(db *sql.DB) *DraftSQLiteStore {
	return &DraftSQLiteStore{db: db}
}

func (s *DraftSQLiteStore) GetAll() ([]models.Draft, error) {
	rows, err := s.db.QueryContext(context.Background(),
		"SELECT id, title, content, topic_id, style_id, status, notes FROM drafts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []models.Draft
	for rows.Next() {
		var d models.Draft
		if err := rows.Scan(&d.Id, &d.Title, &d.Content, &d.TopicId, &d.StyleId, &d.Status, &d.Notes); err != nil {
			return nil, err
		}
		drafts = append(drafts, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range drafts {
		ids, err := s.fetchSourceIds(drafts[i].Id)
		if err != nil {
			return nil, err
		}
		drafts[i].SourceIds = ids
	}
	return drafts, nil
}

func (s *DraftSQLiteStore) Get(id int) (models.Draft, error) {
	var d models.Draft
	row := s.db.QueryRowContext(context.Background(),
		"SELECT id, title, content, topic_id, style_id, status, notes FROM drafts WHERE id = ?", id)
	if err := row.Scan(&d.Id, &d.Title, &d.Content, &d.TopicId, &d.StyleId, &d.Status, &d.Notes); err != nil {
		return d, err
	}
	ids, err := s.fetchSourceIds(id)
	if err != nil {
		return d, err
	}
	d.SourceIds = ids
	return d, nil
}

func (s *DraftSQLiteStore) Create(d models.Draft) (models.Draft, error) {
	if d.Status == "" {
		d.Status = "draft"
	}
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return d, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(context.Background(),
		"INSERT INTO drafts (title, content, topic_id, style_id, status, notes) VALUES (?, ?, ?, ?, ?, ?)",
		d.Title, d.Content, d.TopicId, d.StyleId, d.Status, d.Notes)
	if err != nil {
		return d, err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return d, err
	}

	for _, srcID := range d.SourceIds {
		if _, err := tx.ExecContext(context.Background(),
			"INSERT INTO draft_sources (draft_id, source_id) VALUES (?, ?)", lastID, srcID); err != nil {
			return d, err
		}
	}

	if err := tx.Commit(); err != nil {
		return d, err
	}

	return s.Get(int(lastID))
}

func (s *DraftSQLiteStore) Update(id int, d models.Draft) (models.Draft, error) {
	if d.Status == "" {
		d.Status = "draft"
	}
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return d, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(context.Background(),
		"UPDATE drafts SET title=?, content=?, topic_id=?, style_id=?, status=?, notes=? WHERE id=?",
		d.Title, d.Content, d.TopicId, d.StyleId, d.Status, d.Notes, id)
	if err != nil {
		return d, err
	}

	if _, err = tx.ExecContext(context.Background(),
		"DELETE FROM draft_sources WHERE draft_id=?", id); err != nil {
		return d, err
	}

	for _, srcID := range d.SourceIds {
		if _, err := tx.ExecContext(context.Background(),
			"INSERT INTO draft_sources (draft_id, source_id) VALUES (?, ?)", id, srcID); err != nil {
			return d, err
		}
	}

	if err := tx.Commit(); err != nil {
		return d, err
	}

	return s.Get(id)
}

func (s *DraftSQLiteStore) Delete(id int) error {
	result, err := s.db.ExecContext(context.Background(), "DELETE FROM drafts WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *DraftSQLiteStore) fetchSourceIds(draftID int) ([]int, error) {
	rows, err := s.db.QueryContext(context.Background(),
		"SELECT source_id FROM draft_sources WHERE draft_id = ?", draftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
