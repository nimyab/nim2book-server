create table if not exists google_accounts (
  sub varchar(40) primary key,
  email varchar(255) not null,
  email_verified boolean not null,
  name varchar(255) not null,
  picture varchar(255) null
);
create table if not exists email_password_accounts (
  id uuid primary key default uuid_generate_v4(),
  email varchar(255) not null unique,
  password_hash varchar(255) not null
);
create table if not exists fcm_tokens (
  token varchar(255) primary key,
  user_id uuid,
  create_at timestamp with time zone default CURRENT_TIMESTAMP,
  constraint fk_user_id foreign key (user_id) references users (id) on delete cascade
);
create table if not exists users(
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  is_admin BOOLEAN DEFAULT FALSE NOT NULL,
  is_vip BOOLEAN DEFAULT FALSE,
  metadata JSONB DEFAULT '{}'::JSONB,
  google_account_sub varchar(40),
  email_password_account_id UUID,
  constraint fk_google_account_sub foreign key (google_account_sub) references google_accounts (sub),
  constraint fk_email_password_account_id foreign key (email_password_account_id) references email_password_accounts (id),
  constraint at_least_one_account_type CHECK (
    google_account_sub is not null
    and email_password_account_id is null
    or email_password_account_id is not null
    and google_account_sub is null
  ),
  constraint users_google_account_sub_key UNIQUE (google_account_sub),
  constraint users_email_password_account_id_key UNIQUE (email_password_account_id)
);
create table if not exists books (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  title varchar(255) not null,
  author varchar(255) not null,
  chapter_paths varchar(100) [] not null,
  cover varchar(255) null,
  unique (title, author)
);
create table if not exists dictionary (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  text text not null,
  lang varchar(10) not null,
  content jsonb not null
);