-- +goose Up
-- +goose StatementBegin
create extension if not exists "uuid-ossp";

create table if not exists books(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    title varchar(255) not null,
    author varchar(255) not null,
    chapter_paths varchar(100)[] not null,
    unique (title, author)
);

create table if not exists users(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    login varchar(255) not null unique,
    email varchar(255) not null unique,
    password_hash varchar(255) not null
);

create table if not exists user_books(
    user_id uuid not null,
    book_id uuid not null,
    PRIMARY KEY (user_id, book_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
);

create table if not exists dictionary(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    text text not null,
    lang varchar(10) not null,
    content jsonb not null
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists user_books;
drop table if exists books;
drop table if exists users;
drop table if exists dictionary;
drop extension if exists "uuid-ossp";
-- +goose StatementEnd
