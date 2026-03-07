-- Sources: add AI extraction fields
ALTER TABLE sources ADD COLUMN extract_mode TEXT NOT NULL DEFAULT 'standard';
ALTER TABLE sources ADD COLUMN topic_id INTEGER NOT NULL DEFAULT 0;
