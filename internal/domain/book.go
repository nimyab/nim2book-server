package domain

type Book struct {
	Id           Id       `json:"id" db:"id"`
	Title        string   `json:"title" db:"title"`
	Author       string   `json:"author" db:"author"`
	ChapterPaths []string `json:"chapterPaths" db:"chapter_paths"`
}
