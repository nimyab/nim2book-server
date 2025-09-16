-- +goose Up
-- +goose StatementBegin
alter table books
add column cover varchar(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table books
drop column cover;
-- +goose StatementEnd
