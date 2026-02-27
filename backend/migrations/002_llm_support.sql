-- Topics: add description and keywords
ALTER TABLE topics ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE topics ADD COLUMN keywords    TEXT NOT NULL DEFAULT '';

-- Sources: add new fields (url column kept for backward compat)
ALTER TABLE sources ADD COLUMN type    TEXT NOT NULL DEFAULT 'text';
ALTER TABLE sources ADD COLUMN raw     TEXT NOT NULL DEFAULT '';
ALTER TABLE sources ADD COLUMN content TEXT NOT NULL DEFAULT '';
ALTER TABLE sources ADD COLUMN status  TEXT NOT NULL DEFAULT 'ready';

-- Styles: add tone and example
ALTER TABLE styles ADD COLUMN tone    TEXT NOT NULL DEFAULT '';
ALTER TABLE styles ADD COLUMN example TEXT NOT NULL DEFAULT '';

-- Drafts: add notes
ALTER TABLE drafts ADD COLUMN notes TEXT NOT NULL DEFAULT '';

-- Draft ↔ Source many-to-many
CREATE TABLE IF NOT EXISTS draft_sources (
    draft_id  INTEGER NOT NULL REFERENCES drafts(id)  ON DELETE CASCADE,
    source_id INTEGER NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    PRIMARY KEY (draft_id, source_id)
);

-- Settings key-value store
CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);
