-- Modify "personal_books" table
ALTER TABLE "public"."personal_books" ADD COLUMN "process_status" character varying(30) NOT NULL DEFAULT 'not_started';
