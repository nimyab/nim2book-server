-- Create index "personalbookchapter_order_personal_book_personal_book_chapters" to table: "personal_book_chapters"
CREATE UNIQUE INDEX "personalbookchapter_order_personal_book_personal_book_chapters" ON "public"."personal_book_chapters" ("order", "personal_book_personal_book_chapters");
