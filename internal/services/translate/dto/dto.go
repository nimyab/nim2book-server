package dto

import (
	"time"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
)

// Config содержит параметры конфигурации для сервиса перевода
type Config struct {
	MaxRequestCount int
	WaitDuration    time.Duration
	RunSync         bool // Если true, то перевод будет выполнен синхронно (для тестов)
}

// TranslationContext содержит все данные, необходимые для процесса перевода книги
type TranslationContext[T any] struct {
	UserID     domain.ID
	Book       T // *domain.Book or *domain.PersonalBook
	ParsedData *epub_parser.ParsedData
	From       domain.SupportedLang
	To         domain.SupportedLang
}

// ChapterResult содержит результат обработки одной главы
type ChapterResult[T any] struct {
	ExistChapter T // *domain.BookChapter or *domain.PersonalBookChapter
	Error        error
}
