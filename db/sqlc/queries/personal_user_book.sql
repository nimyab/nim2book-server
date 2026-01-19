-- name: GetPersonalUserBookById :one
select *
from personal_user_books
where personal_user_books.id = $1;

-- name: GetPersonalUserBookGenres :many
select genres.*
from genres
         inner join personal_user_book_genres on genres.id = personal_user_book_genres.genre_id
where personal_user_book_genres.personal_user_book_id = $1;

-- name: GetPersonalUserBookByAuthorAndTitle :one
select *
from personal_user_books
where personal_user_books.author = sqlc.arg('author')
  and personal_user_books.title = sqlc.arg('title')
  and personal_user_books.user_id = sqlc.arg('user_id');

-- name: CreatePersonalUserBook :one
insert into personal_user_books (title, author, chapter_paths, cover, user_id)
values ($1, $2, $3, $4, $5)
returning *;

-- name: AddGenreToPersonalUserBook :exec
insert into personal_user_book_genres (personal_user_book_id, genre_id)
values ($1, $2)
on conflict (personal_user_book_id, genre_id) do nothing;

-- name: RemoveGenreFromPersonalUserBook :exec
delete from personal_user_book_genres
where personal_user_book_id = $1
  and genre_id = $2;

-- name: RemoveAllGenresFromPersonalUserBook :exec
delete from personal_user_book_genres
where personal_user_book_id = $1;

-- name: GetPersonalUserBooksByUserId :many
select distinct personal_user_books.*
from personal_user_books
         left join personal_user_book_genres on personal_user_books.id = personal_user_book_genres.personal_user_book_id
where personal_user_books.user_id = sqlc.arg('user_id')
  and (personal_user_books.author ilike '%' || sqlc.arg('author') || '%')
  and (personal_user_books.title ilike '%' || sqlc.arg('title') || '%')
  and (sqlc.narg('genre_id')::uuid IS NULL
    or exists(select 1
              from personal_user_book_genres pubg
              where pubg.personal_user_book_id = personal_user_books.id
                and pubg.genre_id = sqlc.narg('genre_id')::uuid))
order by personal_user_books.id
limit sqlc.arg('limit') offset sqlc.arg('offset');

-- name: UpdatePersonalUserBook :exec
update personal_user_books
set title  = $1,
    author = $2,
    cover  = $3
where id = $4
  and user_id = $5;

-- name: DeletePersonalUserBook :exec
delete from personal_user_books
where id = $1
  and user_id = $2;
