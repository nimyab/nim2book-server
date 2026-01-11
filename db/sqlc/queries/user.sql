-- name: GetUser :one
select
  sqlc.embed(u),
  sqlc.embed(g),
  sqlc.embed(e)
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
