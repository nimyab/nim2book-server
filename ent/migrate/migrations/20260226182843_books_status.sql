-- Modify "books" table
ALTER TABLE "public"."books" ADD COLUMN "process_status" character varying(30) NOT NULL DEFAULT 'in_progress';
-- Modify "personal_books" table
ALTER TABLE "public"."personal_books" ALTER COLUMN "process_status" SET DEFAULT 'in_progress';
