-- name: GetGenreById :one
select *
from genres
where id = $1;

-- name: GetGenreByName :one
select *
from genres
where name = $1;

-- name: CreateGenre :one
insert into genres (name)
values ($1)
returning *;

-- name: GetAllGenres :many
select *
from genres;

-- name: UpdateGenre :one
update genres
set name = $2
where id = $1
returning *;

-- name: DeleteGenre :exec
delete from genres
where id = $1;
