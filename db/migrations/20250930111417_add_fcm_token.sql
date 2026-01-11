-- +goose Up
-- +goose StatementBegin
alter table users
    add column fcm_token varchar(255) null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users
    drop column fcm_token;
-- +goose StatementEnd
