package domain

type NotificationType string

const (
	NotificationError                   NotificationType = "error"
	NotificationChapterTranslateSucceed NotificationType = "chapter_translate_succeed"
	NotificationBookTranslated          NotificationType = "book_translated"
	NotificationPersonalBookTranslated  NotificationType = "personal_book_translated"
	NotificationTest                    NotificationType = "test"
)

type NotificationErrorData struct {
	Author       string `json:"author"`
	Title        string `json:"title"`
	ErrorMessage string `json:"errorMessage"`
}

type NotificationChapterTranslateSucceedData struct {
	ChapterPath       string `json:"chapterPath"`
	Author            string `json:"author"`
	Title             string `json:"title"`
	ChapterOrder      int    `json:"chapterOrder"`
	TotalChapterCount int    `json:"totalChapterCount"`
}

type NotificationBookTranslatedData struct {
	Book *Book `json:"book"`
}

type NotificationPersonalBookTranslatedData struct {
	Book *PersonalUserBook `json:"book"`
}

type NotificationTestData struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Notification struct {
	UserId Id               `json:"user_id"`
	Type   NotificationType `json:"type"`
	Data   any              `json:"data"`
}
