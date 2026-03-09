package translate_book

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/services/translate/chaptertranslator"
	"github.com/nimyab/nim2book-back/internal/services/translate/dto"
	"github.com/nimyab/nim2book-back/internal/services/translate/parsefile"
	"github.com/nimyab/nim2book-back/proto/word_aligner"
)

type S3 interface {
	Upload(path string, data []byte) error
	Check(path string) error
}

type BookRepository interface {
	GetByAuthorAndTitle(ctx context.Context, authorName, title string) (*domain.Book, error)
	Create(ctx context.Context, book *domain.Book) (*domain.Book, error)
	CreateChapter(ctx context.Context, chapter *domain.BookChapter) (*domain.BookChapter, error)
	GetChapterByBookIDAndOrder(ctx context.Context, bookID domain.ID, orderChapter int) (*domain.BookChapter, error)
	UpdateProcessStatus(ctx context.Context, id domain.ID, processStatus domain.ProcessStatus) (*domain.Book, error)
}

type AuthorRepository interface {
	GetOrCreate(ctx context.Context, name string) (*domain.Author, error)
}

type Translator interface {
	Translate(input *translate.Input) (*translate.Output, error)
}

type NotificationSender interface {
	Emit(notification *domain.Notification)
}

type Service struct {
	config                      dto.Config
	currentCountBookTranslating int64
	s3                          S3
	bookRepo                    BookRepository
	authorRepo                  AuthorRepository
	wordAligner                 word_aligner.AlignmentServiceClient
	translator                  Translator
	notificationSignal          NotificationSender
}

func New(
	s3 S3,
	bookRepo BookRepository,
	authorRepo AuthorRepository,
	wordAligner word_aligner.AlignmentServiceClient,
	translator Translator,
	config dto.Config,
	notificationSignal NotificationSender,
) *Service {
	return &Service{
		s3:                          s3,
		bookRepo:                    bookRepo,
		authorRepo:                  authorRepo,
		wordAligner:                 wordAligner,
		translator:                  translator,
		config:                      config,
		notificationSignal:          notificationSignal,
		currentCountBookTranslating: 0,
	}
}

var (
	ErrFailedToGetBook          = errors.New("failed to get book")
	ErrFailedToCreateBook       = errors.New("failed to create book")
	ErrFailedGetAuthor          = errors.New("failed to get or create author")
	ErrFailedToUpdateStatusBook = errors.New("failed to update status book")
	ErrFailedSaveChapter        = errors.New("failed to save chapter")
	ErrFailedCreateBookChapter  = errors.New("failed to create book chapter")
)

func (s *Service) Throttle() {
	time.Sleep(time.Duration(s.currentCountBookTranslating) * s.config.WaitDuration)
}

func (s *Service) TranslateBook(ctx context.Context, input *Input, file *multipart.FileHeader, userId domain.ID) (*Output, error) {
	const operation = "translate_book.Service.TranslateBook"
	logger := slog.With(slog.String("operation", operation))

	parsedData, err := parsefile.ParseFile(file)
	if err != nil {
		return nil, err
	}

	coverUrl, err := s.saveCoverToS3(parsedData.Cover, parsedData.Book.Title)
	if err != nil {
		logger.Error("failed to save cover to s3", slog.String("error", err.Error()))
		// не возвращаем ошибку, так как обложка не является критичной частью процесса перевода книги
	}

	book, needTranslate, err := s.getBook(ctx,
		parsedData.Book.Author,
		parsedData.Book.Title,
		string(input.From),
		string(input.To),
		coverUrl,
	)
	if err != nil {
		return nil, err
	}
	if !needTranslate {
		if book.ProcessStatus == domain.ProcessStatusCompleted {
			return &Output{Book: book}, nil
		}
		if book.ProcessStatus == domain.ProcessStatusInProgress {
			return &Output{Message: "book is in progress"}, nil
		}
	}
	book, err = s.bookRepo.UpdateProcessStatus(ctx, book.ID, domain.ProcessStatusInProgress)
	if err != nil {
		logger.Error("failed to update book status to in progress", slog.String("error", err.Error()))
		return nil, ErrFailedToUpdateStatusBook
	}

	translationContext := &dto.TranslationContext[*domain.Book]{
		Book:       book,
		ParsedData: parsedData,
		UserID:     userId,
		From:       input.From,
		To:         input.To,
	}

	go s.translate(translationContext)

	return &Output{Message: "start translate"}, nil
}

func (s *Service) getBook(
	ctx context.Context,
	authorName, title, from, to string,
	coverUrl string,
) (*domain.Book, bool, error) {
	const operation = "translate_book.Service.getBook"
	logger := slog.With(slog.String("operation", operation))

	book, err := s.bookRepo.GetByAuthorAndTitle(ctx, authorName, title)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		logger.Error("failed to get book from repository", slog.String("error", err.Error()))
		return nil, false, ErrFailedToGetBook
	}
	if book != nil {
		if book.ProcessStatus == domain.ProcessStatusFailed {
			return book, true, nil
		}
		// если книга уже есть и она либо в процессе, либо завершена, то возвращаем её, не создавая новую
		return book, false, nil
	}

	// Если книги нет, то создаём новую со статусом "В процессе"
	author, err := s.authorRepo.GetOrCreate(ctx, authorName)
	if err != nil {
		logger.Error("failed to get or create author", slog.String("error", err.Error()))
		return nil, false, ErrFailedGetAuthor
	}
	book, err = s.bookRepo.Create(ctx, &domain.Book{
		Title:          title,
		Author:         author,
		OriginalLang:   from,
		TranslatedLang: to,
		CoverURL:       &coverUrl,
		ProcessStatus:  domain.ProcessStatusInProgress,
	})
	if err != nil {
		logger.Error("failed to create book in repository", slog.String("error", err.Error()))
		return nil, false, ErrFailedToCreateBook
	}

	return book, true, nil
}

func (s *Service) translate(translationCtx *dto.TranslationContext[*domain.Book]) {
	const operation = "translate_book.Service.translate"
	logger := slog.With(
		slog.String("operation", operation),
		slog.String("bookId", translationCtx.Book.ID.String()),
		slog.String("userId", translationCtx.UserID.String()),
	)

	atomic.AddInt64(&s.currentCountBookTranslating, 1)
	defer func() {
		atomic.AddInt64(&s.currentCountBookTranslating, -1)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// используем буферизированный канал, чтобы избежать блокировки при отправке результатов
	resultChan := make(chan dto.ChapterResult[*domain.BookChapter], len(translationCtx.ParsedData.FormattedChapter))

	go s.translateChapters(ctx, translationCtx, resultChan)

	for result := range resultChan {
		if result.Error != nil {
			logger.Error("error translating chapter", slog.String("error", result.Error.Error()))
			// при ошибке отменяем весь процесс перевода книги
			cancel()
			// обновляем статус книги на "Ошибка"
			_, err := s.bookRepo.UpdateProcessStatus(context.Background(), translationCtx.Book.ID, domain.ProcessStatusFailed)
			if err != nil {
				logger.Error("failed to update book status to failed", slog.String("error", err.Error()))
			}
			s.notificationSignal.Emit(&domain.Notification{
				Type:   domain.NotificationError,
				UserId: translationCtx.UserID,
				Data: &domain.NotificationErrorData{
					Title:        translationCtx.ParsedData.Book.Title,
					Author:       translationCtx.ParsedData.Book.Author,
					ErrorMessage: result.Error.Error(),
				},
			})
			return
		}
		if result.ExistChapter != nil {
			logger.Info("chapter translated successfully", slog.Any("chapter", result.ExistChapter))
			s.notificationSignal.Emit(&domain.Notification{
				UserId: translationCtx.UserID,
				Type:   domain.NotificationChapterTranslateSucceed,
				Data: &domain.NotificationChapterTranslateSucceedData{
					Author:       translationCtx.ParsedData.Book.Author,
					Title:        translationCtx.ParsedData.Book.Title,
					ChapterOrder: result.ExistChapter.Order,
				},
			})
		}
	}

	logger.Info("translate complete")
	translatedBook, err := s.bookRepo.UpdateProcessStatus(context.Background(), translationCtx.Book.ID, domain.ProcessStatusCompleted)
	if err != nil {
		logger.Error("failed to update book status to completed", slog.String("error", err.Error()))
		return
	}
	s.notificationSignal.Emit(&domain.Notification{
		Type:   domain.NotificationBookTranslated,
		UserId: translationCtx.UserID,
		Data: &domain.NotificationBookTranslatedData{
			Book: translatedBook,
		},
	})
}

func (s *Service) translateChapters(
	ctx context.Context,
	translationCtx *dto.TranslationContext[*domain.Book],
	resultChan chan<- dto.ChapterResult[*domain.BookChapter],
) {
	const operation = "translate_book.Service.translateChapters"
	bookId := translationCtx.Book.ID
	bookTitle := translationCtx.Book.Title
	logger := slog.With(
		slog.String("operation", operation),
		slog.String("bookId", bookId.String()),
		slog.String("bookTitle", bookTitle),
	)

	// закрываем канал после обработки всех глав
	defer close(resultChan)

	chapterTranslator := chaptertranslator.New(
		func(data []byte) (string, error) {
			return s.saveImageToS3(data, bookTitle)
		},
		s.Throttle,
		s.translator,
		s.wordAligner,
	)
	for idx, chapter := range translationCtx.ParsedData.FormattedChapter {
		// chapter order будем считать как i, если брать значения из chapter order, то могут быть проблемы с порядком глав
		// в книге главы могут не содержать какой-то письменный контент и мы такие главы пропускаем при парсинге, а значит порядок глав будет нарушен
		chapterOrder := idx

		select {
		case <-ctx.Done():
			logger.Info("translation cancelled, stopping chapter translation")
			return
		default:
		}

		existChapter, err := s.bookRepo.GetChapterByBookIDAndOrder(ctx, bookId, chapterOrder)
		if err == nil && existChapter != nil {
			logger.Info("translation already translated, stopping chapter translation")
			continue
		}

		chapterAlignNode, err := chapterTranslator.TranslateChapter(ctx, chaptertranslator.Input{
			Chapter:        chapter,
			From:           translationCtx.From,
			To:             translationCtx.To,
			ChapterOrder:   chapterOrder,
			MaxConcurrency: s.config.MaxRequestCount,
		})
		if err != nil {
			logger.Error("failed to translate chapter", slog.String("error", err.Error()), slog.Int("chapterOrder", chapterOrder))
			resultChan <- dto.ChapterResult[*domain.BookChapter]{Error: err}
			return
		}

		chapterUrl, err := s.saveChapterToS3(chapterAlignNode, bookTitle)
		if err != nil {
			logger.Error("failed to save chapter to s3", slog.String("error", err.Error()))
			resultChan <- dto.ChapterResult[*domain.BookChapter]{Error: ErrFailedSaveChapter}
			return
		}

		bookChapter, err := s.bookRepo.CreateChapter(ctx, &domain.BookChapter{
			Book:            &domain.Book{ID: bookId},
			Title:           chapterAlignNode.Title,
			TranslatedTitle: chapterAlignNode.TranslatedTitle,
			Order:           chapterOrder,
			ContentURL:      chapterUrl,
		})
		if err != nil {
			logger.Error("failed to create chapter in repository", slog.String("error", err.Error()))
			resultChan <- dto.ChapterResult[*domain.BookChapter]{Error: ErrFailedCreateBookChapter}
			return
		}

		resultChan <- dto.ChapterResult[*domain.BookChapter]{ExistChapter: bookChapter}
	}
}

func (s *Service) saveImageToS3(data []byte, bookTitle string) (string, error) {
	const operation = "translate_book.Service.saveImageToS3"

	path := fmt.Sprintf("book/%s/images/%s", strings.ReplaceAll(bookTitle, " ", "_"), uuid.New().String())

	if err := s.s3.Upload(path, data); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
}

func (s *Service) saveChapterToS3(chapter *domain.ChapterAlignNode, bookTitle string) (string, error) {
	const operation = "translate_book.Service.saveChapterToS3"

	path := fmt.Sprintf("book/%s/%d.json", strings.ReplaceAll(bookTitle, " ", "_"), chapter.Order)

	data, err := json.Marshal(chapter)
	if err != nil {
		return "", fmt.Errorf("%s: failed to marshal chapter: %w", operation, err)
	}

	if err := s.s3.Upload(path, data); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
}

func (s *Service) saveCoverToS3(coverData []byte, bookTitle string) (string, error) {
	const operation = "translate_book.Service.saveCoverToS3"

	path := fmt.Sprintf("cover/%s/%s", strings.ReplaceAll(bookTitle, " ", "_"), uuid.New().String())

	if err := s.s3.Upload(path, coverData); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
}
