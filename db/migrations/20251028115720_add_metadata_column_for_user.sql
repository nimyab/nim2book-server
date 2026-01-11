-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN metadata JSONB default '{}'::JSONB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN metadata;
-- +goose StatementEnd