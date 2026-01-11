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
