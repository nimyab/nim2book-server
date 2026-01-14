package models

// Book represents a book in the library
type Book struct {
	BaseModel

	Title        string      `gorm:"column:title;type:varchar(255);not null;uniqueIndex:idx_title_author" json:"title"`
	Author       string      `gorm:"column:author;type:varchar(255);not null;uniqueIndex:idx_title_author" json:"author"`
	ChapterPaths StringArray `gorm:"column:chapter_paths;type:varchar(100)[];not null" json:"chapterPaths"`
	Cover        *string     `gorm:"column:cover;type:varchar(255)" json:"cover,omitempty"`
}

func (Book) TableName() string {
	return "books"
}

// GetChapterCount returns the number of chapters in a book
func (b Book) GetChapterCount() int {
	return len(b.ChapterPaths)
}

// HasCover checks if a book has a cover image
func (b Book) HasCover() bool {
	return b.Cover != nil && *b.Cover != ""
}
