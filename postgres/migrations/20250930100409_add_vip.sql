-- +goose Up
-- +goose StatementBegin
alter table users
    add column is_vip boolean default false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users
    drop column is_vip;
-- +goose StatementEnd
