-- +goose Up
-- +goose StatementBegin
create table if not exists genres
(
    id   uuid primary key default uuid_generate_v4(),
    name varchar(100) not null unique
);

alter table books
    add column genre_id uuid,
    add constraint fk_genre_id foreign key (genre_id) references genres (id) on delete set null;

alter table personal_user_books
    add column genre_id uuid,
    add constraint fk_personal_genre_id foreign key (genre_id) references genres (id) on delete set null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table personal_user_books
    drop constraint fk_personal_genre_id,
    drop column genre_id;

alter table books
    drop constraint fk_genre_id,
    drop column genre_id;

drop table if exists genres;
-- +goose StatementEnd
