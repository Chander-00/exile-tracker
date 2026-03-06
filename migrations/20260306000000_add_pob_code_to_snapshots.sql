-- +goose Up
-- +goose StatementBegin
ALTER TABLE pobsnapshots ADD COLUMN pob_code TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pobsnapshots DROP COLUMN pob_code;
-- +goose StatementEnd
