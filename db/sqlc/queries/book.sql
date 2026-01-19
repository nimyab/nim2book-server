-- name: GetBookById :one
select *
from books
where books.id = $1;

-- name: GetBookGenres :many
select genres.*
from genres
         inner join book_genres on genres.id = book_genres.genre_id
where book_genres.book_id = $1;

-- name: GetBookByAuthorAndTitle :one
select *
from books
where books.author = $1
  and books.title = $2;

-- name: CreateBook :one
insert into books (title, author, chapter_paths, cover)
values ($1, $2, $3, $4)
returning *;

-- name: AddGenreToBook :exec
insert into book_genres (book_id, genre_id)
values ($1, $2)
on conflict (book_id, genre_id) do nothing;

-- name: RemoveGenreFromBook :exec
delete from book_genres
where book_id = $1
  and genre_id = $2;

-- name: RemoveAllGenresFromBook :exec
delete from book_genres
where book_id = $1;

-- name: GetBooks :many
select distinct books.*
from books
         left join book_genres on books.id = book_genres.book_id
where (books.author ilike '%' || sqlc.arg('author') || '%')
  and (books.title ilike '%' || sqlc.arg('title') || '%')
  and (sqlc.narg('genre_id')::uuid IS NULL
    or exists(select 1
              from book_genres bg
              where bg.book_id = books.id
                and bg.genre_id = sqlc.narg('genre_id')::uuid))
order by books.id
limit sqlc.arg('limit') offset sqlc.arg('offset');

-- name: UpdateBook :exec
update books
set title  = $1,
    author = $2,
    cover  = $3
where id = $4;

-- name: GetBooksByGenre :many
select distinct books.*
from books
         inner join book_genres on books.id = book_genres.book_id
where book_genres.genre_id = $1
order by books.id
limit sqlc.arg('limit') offset sqlc.arg('offset');

-- name: DeleteBook :exec
delete from books
where id = $1;
