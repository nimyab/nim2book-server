-- +goose Up
-- +goose StatementBegin
create table if not exists google_accounts
(
    sub            varchar(40) primary key,
    email          varchar(255) not null,
    email_verified boolean      not null,
    name           varchar(255) not null,
    picture        varchar(255) null
);

create table if not exists email_password_accounts
(
    id            uuid primary key default uuid_generate_v4(),
    email         varchar(255) not null unique,
    password_hash varchar(255) not null
);

alter table users
    drop column password_hash,
    drop column email,
    add column google_account_sub        varchar(40),
    add constraint fk_google_account_sub foreign key (google_account_sub) references google_accounts (sub),
    add column email_password_account_id uuid,
    add constraint fk_email_password_account_id foreign key (email_password_account_id) references email_password_accounts (id),

    add constraint at_least_one_account_type
        CHECK (google_account_sub is not null and email_password_account_id is null
            or
               email_password_account_id is not null and google_account_sub is null);

create index idx_google_account_sub ON users (google_account_sub);
create index idx_email_password_account_id on users (email_password_account_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists google_account;
drop table if exists email_password_accounts;

alter table users
    drop column google_account_sub,
    drop column email_password_account_id,
    drop constraint fk_google_account_sub,
    drop constraint fk_email_password_account_id,
    drop constraint at_least_one_account_type,
    add column email         varchar(255) not null unique,
    add column password_hash varchar(255) not null;

drop index if exists idx_google_account_sub;
drop index if exists idx_email_password_account_id;
-- +goose StatementEnd
