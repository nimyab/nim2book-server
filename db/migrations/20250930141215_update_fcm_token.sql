-- +goose Up
-- +goose StatementBegin
alter table users
    add constraint users_google_account_sub_key UNIQUE (google_account_sub),
    add constraint users_email_password_account_id_key UNIQUE (email_password_account_id),
    drop column fcm_token;

create table if not exists fcm_tokens
(
    token     varchar(255) primary key,
    user_id   uuid,
    create_at timestamp with time zone default CURRENT_TIMESTAMP,

    constraint fk_user_id foreign key (user_id) references users (id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users
    drop constraint users_google_account_sub_key,
    drop constraint users_email_password_account_id_key,
    add column fcm_token varchar(255) null;

drop table if exists fcm_tokens;
-- +goose StatementEnd
