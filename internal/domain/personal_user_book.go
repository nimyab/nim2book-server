package domain

// PersonalUserBook это книги конкретных пользователей, которые они добавили сами, эти книги не из общего каталога
type PersonalUserBook struct {
	Id           Id       `json:"id"`
	Title        string   `json:"title"`
	Author       string   `json:"author"`
	ChapterPaths []string `json:"chapterPaths"`
	Cover        *string  `json:"cover,omitempty"`
	Genres       []Genre  `json:"genres,omitempty"`
	UserId       Id       `json:"userId"`
}
