-- +goose Up
-- +goose StatementBegin
ALTER TABLE characters_to_fetch ADD COLUMN updated_at TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE characters_to_fetch DROP COLUMN updated_at;
-- +goose StatementEnd
