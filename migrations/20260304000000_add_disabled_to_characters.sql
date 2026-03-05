-- +goose Up
ALTER TABLE characters ADD COLUMN disabled INTEGER NOT NULL DEFAULT 0;

-- +goose Down
-- SQLite does not support DROP COLUMN easily; manual intervention required.
