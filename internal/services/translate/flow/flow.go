package flow

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/nimyab/nim2book-back/internal/services/translate/dto"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
)

// BookInterface определяет общие методы, необходимые для сущности книги
type BookInterface interface {
	GetID() domain.ID
	GetProcessStatus() domain.ProcessStatus
	SetProcessStatus(status domain.ProcessStatus)
}

// ChapterInterface определяет общие методы, необходимые для сущности главы
type ChapterInterface interface {
	GetOrder() int
}

// Deps содержит все зависимости, необходимые для потока перевода
type Deps[TBook BookInterface, TChapter ChapterInterface] struct {
	GetBook          func(ctx context.Context) (TBook, error)
	CreateBook       func(ctx context.Context) (TBook, error)
	UpdateBookStatus func(ctx context.Context, id domain.ID, status domain.ProcessStatus) (TBook, error)

	GetChapter    func(ctx context.Context, bookID domain.ID, order int) (TChapter, error)
	CreateChapter func(ctx context.Context, book TBook, chapter *domain.ChapterAlignNode, contentURL string) (TChapter, error)

	SaveChapterToS3 func(chapter *domain.ChapterAlignNode) (string, error)
	SaveImageToS3   func(data []byte) (string, error)

	TranslateChapter func(ctx context.Context, chapter epub_parser.FormattedChapter, imageSaver func([]byte) (string, error)) (*domain.ChapterAlignNode, error)

	NotifyChapter func(ctx context.Context, chapter TChapter)
	NotifyBook    func(ctx context.Context, book TBook)
	NotifyError   func(ctx context.Context, err error)
}

// Run выполняет подготовку к потоку перевода
func Run[TBook BookInterface, TChapter ChapterInterface](
	ctx context.Context,
	input *dto.TranslationContext[TBook],
	deps Deps[TBook, TChapter],
) (*dto.TranslationContext[TBook], error) {
	const operation = "translate.flow.Run"

	// 1. Проверка существования книги
	book, err := deps.GetBook(ctx)
	if err != nil {
		slog.Error("GetBook error", slog.String("err", err.Error()), slog.String("type", fmt.Sprintf("%T", err)))
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%s: не удалось получить книгу: %w", operation, err)
		}
	}

	// Если книга существует
	if !isNil(book) {
		// Если статус завершен или уже в процессе перевода, возвращаем её
		if book.GetProcessStatus() == domain.ProcessStatusCompleted || book.GetProcessStatus() == domain.ProcessStatusInProgress {
			input.BookEntity = book
			return input, nil
		}
		// Обновляем статус на "В процессе"
		book, err = deps.UpdateBookStatus(ctx, book.GetID(), domain.ProcessStatusInProgress)
		if err != nil {
			return nil, fmt.Errorf("%s: не удалось обновить статус книги: %w", operation, err)
		}
	} else {
		// Создаем книгу со статусом "В процессе"
		book, err = deps.CreateBook(ctx)
		if err != nil {
			return nil, fmt.Errorf("%s: не удалось создать книгу: %w", operation, err)
		}
	}
	input.BookEntity = book

	return input, nil
}

// TranslateChapters итерирует и переводит главы
func TranslateChapters[TBook BookInterface, TChapter ChapterInterface](
	ctx context.Context,
	input *dto.TranslationContext[TBook],
	deps Deps[TBook, TChapter],
) {
	const operation = "translate.flow.TranslateChapters"

	book := input.BookEntity

	for i, chapterData := range input.ParsedData.FormattedChapter {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Проверка существования главы в БД
		existChapter, err := deps.GetChapter(ctx, book.GetID(), i)
		if err == nil && !isNil(existChapter) {
			slog.Info(fmt.Sprintf("глава %d уже существует", i), slog.String("operation", operation))
			continue
		}

		// Перевод
		translatedChapter, err := deps.TranslateChapter(ctx, chapterData, deps.SaveImageToS3)
		if err != nil {
			slog.Error("не удалось перевести главу", slog.Any("error", err), slog.Int("chapter", i))
			deps.NotifyError(ctx, fmt.Errorf("не удалось перевести главу %d: %w", i, err))
			// Если перевод не удался, мы должны уведомить и, возможно, продолжить или остановиться.
			// В текущей реализации мы останавливаемся для этой главы, но цикл продолжается?
			// Нет, return прерывает весь процесс перевода книги.
			return
		}
		translatedChapter.Order = i

		// Сохранение в S3
		path, err := deps.SaveChapterToS3(translatedChapter)
		if err != nil {
			slog.Error("не удалось сохранить главу в s3", slog.Any("error", err), slog.Int("chapter", i))
			deps.NotifyError(ctx, fmt.Errorf("не удалось сохранить главу %d в s3: %w", i, err))
			return
		}

		// Сохранение в БД
		newChapter, err := deps.CreateChapter(ctx, book, translatedChapter, path)
		if err != nil {
			slog.Error("не удалось создать главу в бд", slog.Any("error", err), slog.Int("chapter", i))
			deps.NotifyError(ctx, fmt.Errorf("не удалось сохранить главу %d в бд: %w", i, err))
			return
		}

		// Уведомление
		deps.NotifyChapter(ctx, newChapter)
	}

	// Все главы готовы, обновляем статус на Завершено
	book, err := deps.UpdateBookStatus(ctx, book.GetID(), domain.ProcessStatusCompleted)
	if err != nil {
		slog.Error("не удалось обновить статус книги на завершено", slog.Any("error", err))
		// Не критично, но стоит залогировать
	}

	slog.Info("все главы переведены и сохранены", slog.String("operation", operation), slog.Any("book", book))

	deps.NotifyBook(ctx, book)
}

func isNil(i any) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Pointer, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
