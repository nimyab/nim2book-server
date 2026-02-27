package translate_book

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/services/translate/dto"
	"github.com/nimyab/nim2book-back/internal/services/translate/flow"
	"github.com/nimyab/nim2book-back/internal/services/translate/translate_logic"
	"github.com/nimyab/nim2book-back/internal/services/word_aligner/align"
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	pb "github.com/nimyab/nim2book-back/proto/word_aligner"
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

type WordAligner interface {
	Align(input *align.Input) (*align.Output, error)
}

type Translator interface {
	Translate(input *translate.Input) (*translate.Output, error)
}

type NotificationSender interface {
	Emit(ctx context.Context, notification *domain.Notification)
}

type Service struct {
	config dto.Config

	mu                          sync.Mutex
	currentCountBookTranslating int

	s3                 S3
	bookRepo           BookRepository
	authorRepo         AuthorRepository
	wordAligner        pb.AlignmentServiceClient
	translator         Translator
	notificationSignal NotificationSender
	logic              *translate_logic.Logic
}

var (
	ErrFailedTranslateChapter   = errors.New("failed translate chapter")
	ErrFailedSaveToStorage      = errors.New("failed save to storage")
	ErrFailedSaveBookToDatabase = errors.New("failed save book to database")
)

func New(
	s3 S3,
	bookRepo BookRepository,
	authorRepo AuthorRepository,
	wordAligner pb.AlignmentServiceClient,
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
		logic:                       translate_logic.NewLogic(translator, wordAligner),
	}
}

func (s *Service) Throttle() {
	time.Sleep(time.Duration(s.currentCountBookTranslating) * s.config.WaitDuration)
}

func (s *Service) TranslateBook(ctx context.Context, input *Input, book *multipart.FileHeader, userId domain.ID) (*Output, error) {
	const operation = "translate_book.Service.TranslateBook"

	file, err := book.Open()
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to open book file")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to read book file")
	}

	startParse := time.Now()

	parsedData, err := epub_parser.Parse(data)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, errors.New("failed to parse book file")
	}

	slog.Info(
		"Book is parsed successfully",
		slog.Int64("duration ms", time.Since(startParse).Milliseconds()),
		slog.Int("chapters count", len(parsedData.FormattedChapter)),
		slog.String("author", parsedData.Book.Author),
		slog.String("title", parsedData.Book.Title),
	)

	translationContext := &dto.TranslationContext[*domain.Book]{
		ParsedData: parsedData,
		UserID:     userId,
		From:       input.From,
		To:         input.To,
	}

	return s.startTranslate(ctx, translationContext)
}

func (s *Service) startTranslate(
	ctx context.Context,
	translationContext *dto.TranslationContext[*domain.Book],
) (*Output, error) {
	const operation = "translate_book.Service.startTranslate"

	deps := flow.Deps[*domain.Book, *domain.BookChapter]{
		GetBook: func(ctx context.Context) (*domain.Book, error) {
			return s.bookRepo.GetByAuthorAndTitle(ctx, translationContext.ParsedData.Book.Author, translationContext.ParsedData.Book.Title)
		},
		CreateBook: func(ctx context.Context) (*domain.Book, error) {
			// Получаем или создаём автора
			author, err := s.authorRepo.GetOrCreate(ctx, translationContext.ParsedData.Book.Author)
			if err != nil {
				return nil, fmt.Errorf("failed to get or create author: %w", err)
			}

			var coverURL *string
			if translationContext.ParsedData.Cover != nil {
				coverPath, err := s.saveCoverToS3(translationContext.ParsedData.Cover, translationContext.ParsedData.Book.Title)
				if err != nil {
					slog.Error(err.Error(), slog.String("operation", operation))
				} else {
					coverURL = &coverPath
				}
			}

			newBook := &domain.Book{
				Author:         author,
				Title:          translationContext.ParsedData.Book.Title,
				CoverURL:       coverURL,
				OriginalLang:   string(translationContext.From),
				TranslatedLang: string(translationContext.To),
				ProcessStatus:  domain.ProcessStatusInProgress,
			}
			return s.bookRepo.Create(ctx, newBook)
		},
		UpdateBookStatus: func(ctx context.Context, id domain.ID, status domain.ProcessStatus) (*domain.Book, error) {
			return s.bookRepo.UpdateProcessStatus(ctx, id, status)
		},
		GetChapter: func(ctx context.Context, bookID domain.ID, order int) (*domain.BookChapter, error) {
			return s.bookRepo.GetChapterByBookIDAndOrder(ctx, bookID, order)
		},
		CreateChapter: func(ctx context.Context, book *domain.Book, chapter *domain.ChapterAlignNode, contentURL string) (*domain.BookChapter, error) {
			newChapter := &domain.BookChapter{
				Book:            book,
				Order:           chapter.Order,
				ContentURL:      contentURL,
				Title:           chapter.Title,
				TranslatedTitle: chapter.TranslatedTitle,
			}
			return s.bookRepo.CreateChapter(ctx, newChapter)
		},
		SaveChapterToS3: func(chapter *domain.ChapterAlignNode) (string, error) {
			return s.saveChapterToS3(chapter, translationContext.ParsedData.Book.Title)
		},
		SaveImageToS3: func(data []byte) (string, error) {
			return s.saveImageToS3(data, translationContext.ParsedData.Book.Title)
		},
		TranslateChapter: func(ctx context.Context, chapter epub_parser.FormattedChapter, imageSaver func([]byte) (string, error)) (*domain.ChapterAlignNode, error) {
			return s.logic.TranslateChapter(ctx, chapter, translationContext.From, translationContext.To, s, s.config.MaxRequestCount, imageSaver)
		},
		NotifyChapter: func(ctx context.Context, chapter *domain.BookChapter) {
			if s.notificationSignal != nil {
				s.notificationSignal.Emit(ctx, &domain.Notification{
					UserId: translationContext.UserID,
					Type:   domain.NotificationChapterTranslateSucceed,
					Data: &domain.NotificationChapterTranslateSucceedData{
						Author:            translationContext.ParsedData.Book.Author,
						ChapterPath:       chapter.ContentURL,
						Title:             translationContext.ParsedData.Book.Title,
						ChapterOrder:      chapter.Order,
						TotalChapterCount: len(translationContext.ParsedData.FormattedChapter),
					},
				})
			}
		},
		NotifyBook: func(ctx context.Context, book *domain.Book) {
			if s.notificationSignal != nil {
				s.notificationSignal.Emit(ctx, &domain.Notification{
					UserId: translationContext.UserID,
					Type:   domain.NotificationBookTranslated,
					Data: &domain.NotificationBookTranslatedData{
						Book: book,
					},
				})
			}
		},
		NotifyError: func(ctx context.Context, err error) {
			if s.notificationSignal != nil {
				s.notificationSignal.Emit(ctx, &domain.Notification{
					UserId: translationContext.UserID,
					Type:   domain.NotificationError,
					Data: &domain.NotificationErrorData{
						Title:        translationContext.ParsedData.Book.Title,
						Author:       translationContext.ParsedData.Book.Author,
						ErrorMessage: err.Error(),
					},
				})
			}
		},
	}

	resultCtx, err := flow.Run(ctx, translationContext, deps)
	if err != nil {
		slog.Error("failed to run translation flow", slog.Any("error", err), slog.String("operation", operation))
		return nil, err
	}

	if resultCtx.BookEntity.ProcessStatus == domain.ProcessStatusCompleted {
		return &Output{Book: resultCtx.BookEntity}, nil
	}

	if resultCtx.BookEntity.ProcessStatus == domain.ProcessStatusInProgress {
		return &Output{Message: "book is in progress"}, nil
	}

	runChapters := func() {
		s.mu.Lock()
		s.currentCountBookTranslating++
		s.mu.Unlock()
		defer func() {
			s.mu.Lock()
			s.currentCountBookTranslating--
			s.mu.Unlock()
		}()

		flow.TranslateChapters(context.Background(), resultCtx, deps)
	}

	go runChapters()

	return &Output{Message: "start translate"}, nil
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
