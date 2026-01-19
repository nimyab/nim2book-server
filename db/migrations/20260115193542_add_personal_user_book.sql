-- +goose Up
-- +goose StatementBegin
create table if not exists personal_user_books (
    id uuid primary key default uuid_generate_v4(),
    title         varchar(255)   not null,
    author        varchar(255)   not null,
    chapter_paths varchar(100)[] not null,
    cover varchar(255) null,
    user_id uuid not null,
    constraint fk_user_id foreign key (user_id) references users (id) on delete cascade,
    unique (title, author)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists personal_user_books;
-- +goose StatementEnd
