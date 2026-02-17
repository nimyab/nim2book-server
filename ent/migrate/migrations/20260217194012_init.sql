-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL,
  "is_vip" boolean NOT NULL DEFAULT false,
  "is_admin" boolean NOT NULL DEFAULT false,
  "metadata" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "basic_accounts" table
CREATE TABLE "public"."basic_accounts" (
  "id" uuid NOT NULL,
  "email" character varying(255) NOT NULL,
  "password_hash" character varying(255) NOT NULL,
  "is_verified" boolean NOT NULL DEFAULT false,
  "verify_link" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "user_basic_account" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "basic_accounts_users_basic_account" FOREIGN KEY ("user_basic_account") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "basic_accounts_email_key" to table: "basic_accounts"
CREATE UNIQUE INDEX "basic_accounts_email_key" ON "public"."basic_accounts" ("email");
-- Create index "basic_accounts_user_basic_account_key" to table: "basic_accounts"
CREATE UNIQUE INDEX "basic_accounts_user_basic_account_key" ON "public"."basic_accounts" ("user_basic_account");
-- Create index "basic_accounts_verify_link_key" to table: "basic_accounts"
CREATE UNIQUE INDEX "basic_accounts_verify_link_key" ON "public"."basic_accounts" ("verify_link");
-- Create "authors" table
CREATE TABLE "public"."authors" (
  "id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "books" table
CREATE TABLE "public"."books" (
  "id" uuid NOT NULL,
  "title" character varying(255) NOT NULL,
  "cover_url" character varying(255) NOT NULL,
  "original_lang" character varying(10) NOT NULL DEFAULT 'en',
  "translated_lang" character varying(10) NOT NULL DEFAULT 'ru',
  "created_at" timestamptz NOT NULL,
  "author_books" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "books_authors_books" FOREIGN KEY ("author_books") REFERENCES "public"."authors" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "book_chapters" table
CREATE TABLE "public"."book_chapters" (
  "id" uuid NOT NULL,
  "order" bigint NOT NULL,
  "title" character varying(255) NOT NULL,
  "translated_title" character varying(255) NOT NULL,
  "content_url" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "book_book_chapters" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "book_chapters_books_book_chapters" FOREIGN KEY ("book_book_chapters") REFERENCES "public"."books" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "dictionaries" table
CREATE TABLE "public"."dictionaries" (
  "id" uuid NOT NULL,
  "text" character varying(255) NOT NULL,
  "part_of_speech" character varying(255) NOT NULL,
  "transcription" character varying(255) NOT NULL,
  "from_lang_code" character varying(10) NOT NULL,
  "to_lang_code" character varying(10) NOT NULL,
  "translations" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "dictionary_text_part_of_speech" to table: "dictionaries"
CREATE UNIQUE INDEX "dictionary_text_part_of_speech" ON "public"."dictionaries" ("text", "part_of_speech");
-- Create "dictionary_examples" table
CREATE TABLE "public"."dictionary_examples" (
  "id" uuid NOT NULL,
  "text" character varying NOT NULL,
  "translation" character varying NOT NULL,
  "target_position_start" bigint NOT NULL,
  "target_position_end" bigint NOT NULL,
  "created_at" timestamptz NOT NULL,
  "dictionary_dictionary_examples" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "dictionary_examples_dictionaries_dictionary_examples" FOREIGN KEY ("dictionary_dictionary_examples") REFERENCES "public"."dictionaries" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "fcm_tokens" table
CREATE TABLE "public"."fcm_tokens" (
  "id" uuid NOT NULL,
  "token" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "user_fcm_tokens" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fcm_tokens_users_fcm_tokens" FOREIGN KEY ("user_fcm_tokens") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "fcm_tokens_token_key" to table: "fcm_tokens"
CREATE UNIQUE INDEX "fcm_tokens_token_key" ON "public"."fcm_tokens" ("token");
-- Create "genres" table
CREATE TABLE "public"."genres" (
  "id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "genres_name_key" to table: "genres"
CREATE UNIQUE INDEX "genres_name_key" ON "public"."genres" ("name");
-- Create "genre_books" table
CREATE TABLE "public"."genre_books" (
  "genre_id" uuid NOT NULL,
  "book_id" uuid NOT NULL,
  PRIMARY KEY ("genre_id", "book_id"),
  CONSTRAINT "genre_books_book_id" FOREIGN KEY ("book_id") REFERENCES "public"."books" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "genre_books_genre_id" FOREIGN KEY ("genre_id") REFERENCES "public"."genres" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "personal_books" table
CREATE TABLE "public"."personal_books" (
  "id" uuid NOT NULL,
  "title" character varying(255) NOT NULL,
  "cover_url" character varying(255) NOT NULL,
  "original_lang" character varying(10) NOT NULL DEFAULT 'en',
  "translated_lang" character varying(10) NOT NULL DEFAULT 'ru',
  "created_at" timestamptz NOT NULL,
  "author_personal_books" uuid NULL,
  "user_personal_books" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "personal_books_authors_personal_books" FOREIGN KEY ("author_personal_books") REFERENCES "public"."authors" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "personal_books_users_personal_books" FOREIGN KEY ("user_personal_books") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "genre_personal_books" table
CREATE TABLE "public"."genre_personal_books" (
  "genre_id" uuid NOT NULL,
  "personal_book_id" uuid NOT NULL,
  PRIMARY KEY ("genre_id", "personal_book_id"),
  CONSTRAINT "genre_personal_books_genre_id" FOREIGN KEY ("genre_id") REFERENCES "public"."genres" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "genre_personal_books_personal_book_id" FOREIGN KEY ("personal_book_id") REFERENCES "public"."personal_books" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "google_accounts" table
CREATE TABLE "public"."google_accounts" (
  "id" uuid NOT NULL,
  "sub" character varying(50) NOT NULL,
  "email" character varying(255) NOT NULL,
  "email_verified" boolean NOT NULL,
  "name" character varying(255) NOT NULL,
  "picture" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "user_google_account" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "google_accounts_users_google_account" FOREIGN KEY ("user_google_account") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "google_accounts_email_key" to table: "google_accounts"
CREATE UNIQUE INDEX "google_accounts_email_key" ON "public"."google_accounts" ("email");
-- Create index "google_accounts_sub_key" to table: "google_accounts"
CREATE UNIQUE INDEX "google_accounts_sub_key" ON "public"."google_accounts" ("sub");
-- Create index "google_accounts_user_google_account_key" to table: "google_accounts"
CREATE UNIQUE INDEX "google_accounts_user_google_account_key" ON "public"."google_accounts" ("user_google_account");
-- Create "personal_book_chapters" table
CREATE TABLE "public"."personal_book_chapters" (
  "id" uuid NOT NULL,
  "order" bigint NOT NULL,
  "title" character varying(255) NOT NULL,
  "translated_title" character varying(255) NOT NULL,
  "content_url" character varying(255) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "personal_book_personal_book_chapters" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "personal_book_chapters_personal_books_personal_book_chapters" FOREIGN KEY ("personal_book_personal_book_chapters") REFERENCES "public"."personal_books" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
