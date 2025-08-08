-- +goose Up
-- +goose StatementBegin
create extension if not exists "uuid-ossp";

create table if not exists books
(
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    title         varchar(255)   not null,
    author        varchar(255)   not null,
    chapter_paths varchar(100)[] not null,
    unique (title, author)
);

create table if not exists users
(
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         varchar(255)                   not null unique,
    password_hash varchar(255)                   not null,
    is_admin      boolean          default false not null
);

create table if not exists dictionary
(
    id      uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    text    text        not null,
    lang    varchar(10) not null,
    content jsonb       not null
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists books;
drop table if exists users;
drop table if exists dictionary;
drop extension if exists "uuid-ossp";
-- +goose StatementEnd
