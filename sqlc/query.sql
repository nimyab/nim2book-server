-- name: GetUser :one
select u.id,
  u.is_admin,
  u.is_vip,
  u.metadata,
  g.email as google_email,
  g.email_verified as google_email_verified,
  g.name as google_name,
  g.picture as google_picture,
  g.sub as google_sub,
  e.id as email_password_id,
  e.email as email_password_email,
  e.password_hash
from users as u
  left join google_accounts as g on u.google_account_sub = g.sub
  left join email_password_accounts as e on u.email_password_account_id = e.id
where (
    sqlc.narg('user_id')::uuid IS NULL
    OR u.id = sqlc.narg('user_id')::uuid
  )
  AND (
    sqlc.narg('google_sub')::varchar(40) IS NULL
    OR u.google_account_sub = sqlc.narg('google_sub')::varchar(40)
  )
  AND (
    sqlc.narg('google_email')::varchar(255) IS NULL
    OR g.email = sqlc.narg('google_email')::varchar(255)
  )
  AND (
    sqlc.narg('email_password_account_id')::uuid IS NULL
    OR u.email_password_account_id = sqlc.narg('email_password_account_id')::uuid
  )
  AND (
    sqlc.narg('email')::varchar(255) IS NULL
    OR e.email = sqlc.narg('email')::varchar(255)
  );
-- name: GetEmailPasswordAccountByEmail :one
select id
from email_password_accounts
where email = $1;
-- name: CreateEmailPasswordAccount :one
insert into email_password_accounts (email, password_hash)
values ($1, $2)
returning id;
-- name: CreateUserByEmailPasswordAccountId :one
insert into users (email_password_account_id)
values ($1)
returning id,
  is_admin,
  is_vip,
  metadata;
-- name: CreateGoogleAccount :exec
insert into google_accounts (sub, email, email_verified, name, picture)
values ($1, $2, $3, $4, $5);
-- name: CreateUserByGoogleSub :one
insert into users (google_account_sub)
values ($1)
returning id,
  is_admin,
  is_vip,
  metadata;
-- name: UpdateUserMetadata :exec
update users
set metadata = $1
where id = $2;
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
where (author ilike '%' || $1 || '%')
  and (title ilike '%' || $2 || '%')
limit $3 offset $4;
-- name: UpdateBook :exec
update books
set title = $1,
  author = $2,
  cover = $3
where id = $4;
-- name: GetDictionaryData :one
select content
from dictionary
where text = $1
  and lang = $2;
-- name: DictionaryDataExists :one
select exists(
    select id
    from dictionary
    where text = $1
      and lang = $2
  );
-- name: CreateDictionaryData :exec
insert into dictionary (text, lang, content)
values ($1, $2, $3);
-- name: GetFcmTokensByUserId :many
select *
from fcm_tokens
where user_id = $1;
-- name: GetFcmTokenByToken :one
select *
from fcm_tokens
where token = $1;
-- name: AddFcmToken :one
insert into fcm_tokens (token, user_id)
values ($1, $2)
returning token,
  user_id,
  create_at;
-- name: DeleteFcmToken :exec
delete from fcm_tokens
where token = $1
  and user_id = $2;