-- name: GetBookById :one
select *
from books
where id = $1;

-- name: GetBookByAuthorAndTitle :one
select *
from books
where author = $1
  and title = $2;

-- name: CreateBook :one
insert into books (title, author, chapter_paths, cover)
values ($1, $2, $3, $4)
returning id;

-- name: GetBooks :many
select *
from books
where (author ilike '%' || sqlc.arg('author') || '%')
  and (title ilike '%' || sqlc.arg('title') || '%')
limit sqlc.arg('limit')
offset sqlc.arg('offset');

-- name: UpdateBook :exec
update books
set title = $1,
  author = $2,
  cover = $3
where id = $4;
