package dto

import (
	"time"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	"github.com/timsims/pamphlet"
)

// Config содержит параметры конфигурации для сервиса перевода
type Config struct {
	MaxRequestCount int
	WaitDuration    time.Duration
}

// TranslationContext содержит все данные, необходимые для процесса перевода книги
type TranslationContext struct {
	UserID       domain.ID
	Book         *pamphlet.Book
	Chapters     []epub_parser.FormattedChapter
	CoverData    []byte
	From         domain.SupportedLang
	To           domain.SupportedLang
	PersonalBook *domain.PersonalBook // Опционально, только для персональных книг
}

// ChapterResult содержит результат обработки одной главы
type ChapterResult struct {
	Chapter      *domain.ChapterAlignNode
	ExistChapter *domain.PersonalBookChapter // Опционально, если глава уже существует
	ChapterOrder int
	Path         string
	Error        error
}
