-- +goose Up
-- +goose StatementBegin

-- Create junction table for books and genres (many-to-many)
create table if not exists book_genres
(
    book_id  uuid not null,
    genre_id uuid not null,
    constraint fk_book_genres_book_id foreign key (book_id) references books (id) on delete cascade,
    constraint fk_book_genres_genre_id foreign key (genre_id) references genres (id) on delete cascade,
    primary key (book_id, genre_id)
);

-- Create junction table for personal_user_books and genres (many-to-many)
create table if not exists personal_user_book_genres
(
    personal_user_book_id uuid not null,
    genre_id              uuid not null,
    constraint fk_personal_user_book_genres_personal_user_book_id foreign key (personal_user_book_id) references personal_user_books (id) on delete cascade,
    constraint fk_personal_user_book_genres_genre_id foreign key (genre_id) references genres (id) on delete cascade,
    primary key (personal_user_book_id, genre_id)
);

-- Create indexes for better performance
create index if not exists idx_book_genres_book_id on book_genres (book_id);
create index if not exists idx_book_genres_genre_id on book_genres (genre_id);
create index if not exists idx_personal_user_book_genres_book_id on personal_user_book_genres (personal_user_book_id);
create index if not exists idx_personal_user_book_genres_genre_id on personal_user_book_genres (genre_id);

-- Migrate existing data from books.genre_id to book_genres
insert into book_genres (book_id, genre_id)
select id, genre_id
from books
where genre_id is not null;

-- Migrate existing data from personal_user_books.genre_id to personal_user_book_genres
insert into personal_user_book_genres (personal_user_book_id, genre_id)
select id, genre_id
from personal_user_books
where genre_id is not null;

-- Drop foreign key constraints before dropping columns
alter table books
    drop constraint if exists fk_genre_id;

alter table personal_user_books
    drop constraint if exists fk_genre_id;

-- Drop old genre_id columns
alter table books
    drop column if exists genre_id;

alter table personal_user_books
    drop column if exists genre_id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Add back genre_id columns
alter table books
    add column genre_id uuid null;

alter table personal_user_books
    add column genre_id uuid null;

-- Add back foreign key constraints
alter table books
    add constraint fk_genre_id foreign key (genre_id) references genres (id) on delete set null;

alter table personal_user_books
    add constraint fk_genre_id foreign key (genre_id) references genres (id) on delete set null;

-- Migrate data back (only first genre if multiple exist)
update books
set genre_id = bg.genre_id
from (select distinct on (book_id) book_id, genre_id
      from book_genres) bg
where books.id = bg.book_id;

update personal_user_books
set genre_id = pubg.genre_id
from (select distinct on (personal_user_book_id) personal_user_book_id, genre_id
      from personal_user_book_genres) pubg
where personal_user_books.id = pubg.personal_user_book_id;

-- Drop indexes
drop index if exists idx_book_genres_book_id;
drop index if exists idx_book_genres_genre_id;
drop index if exists idx_personal_user_book_genres_book_id;
drop index if exists idx_personal_user_book_genres_genre_id;

-- Drop junction tables (foreign keys will be dropped automatically)
drop table if exists book_genres;
drop table if exists personal_user_book_genres;

-- +goose StatementEnd
