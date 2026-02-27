package translate_personal_user_book

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
	"github.com/nimyab/nim2book-back/pkg/parsers/epub_parser"
	pb "github.com/nimyab/nim2book-back/proto/word_aligner"
)

type S3 interface {
	Upload(path string, data []byte) error
	Check(path string) error
}

type PersonalBookRepository interface {
	GetByUserAndAuthorAndTitle(ctx context.Context, userID domain.ID, authorName, title string) (*domain.PersonalBook, error)
	Create(ctx context.Context, book *domain.PersonalBook) (*domain.PersonalBook, error)
	CreateChapter(ctx context.Context, chapter *domain.PersonalBookChapter) (*domain.PersonalBookChapter, error)
	GetChapterByPersonalBookIDAndOrder(ctx context.Context, personalBookID domain.ID, orderChapter int) (*domain.PersonalBookChapter, error)
	UpdateProcessStatus(ctx context.Context, id domain.ID, processStatus domain.ProcessStatus) (*domain.PersonalBook, error)
}

type AuthorRepository interface {
	GetOrCreate(ctx context.Context, name string) (*domain.Author, error)
}

type WordAligner interface {
	Align(ctx context.Context, req *pb.AlignRequest) (*pb.AlignResponse, error)
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

	s3               S3
	personalBookRepo PersonalBookRepository
	authorRepo       AuthorRepository
	wordAligner      pb.AlignmentServiceClient
	translator       Translator

	notificationSignal NotificationSender

	logic *translate_logic.Logic
}

var (
	ErrFailedTranslateChapter   = errors.New("failed translate chapter")
	ErrFailedSaveToStorage      = errors.New("failed save to storage")
	ErrFailedSaveBookToDatabase = errors.New("failed save book to database")
	ErrFailedGetOrCreateAuthor  = errors.New("failed get or create author")
)

func New(
	s3 S3,
	personalBookRepo PersonalBookRepository,
	authorRepo AuthorRepository,
	wordAligner pb.AlignmentServiceClient,
	translator Translator,
	config dto.Config,
	notificationSignal NotificationSender,
) *Service {
	return &Service{
		s3:                          s3,
		personalBookRepo:            personalBookRepo,
		authorRepo:                  authorRepo,
		wordAligner:                 wordAligner,
		translator:                  translator,
		config:                      config,
		currentCountBookTranslating: 0,
		notificationSignal:          notificationSignal,
		logic:                       translate_logic.NewLogic(translator, wordAligner),
	}
}

func (s *Service) Throttle() {
	time.Sleep(time.Duration(s.currentCountBookTranslating) * s.config.WaitDuration)
}

func (s *Service) TranslatePersonalUserBook(ctx context.Context, input *Input, book *multipart.FileHeader, userId domain.ID) (*Output, error) {
	const operation = "translate_personal_user_book.Service.TranslatePersonalUserBook"

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

	translationContext := &dto.TranslationContext[*domain.PersonalBook]{
		ParsedData: parsedData,
		UserID:     userId,
		From:       input.From,
		To:         input.To,
	}

	return s.startTranslate(ctx, translationContext, true)
}

func (s *Service) startTranslate(
	ctx context.Context,
	translationContext *dto.TranslationContext[*domain.PersonalBook],
	runAsync bool,
) (*Output, error) {
	const operation = "translate_personal_user_book.Service.startTranslate"

	deps := flow.Deps[*domain.PersonalBook, *domain.PersonalBookChapter]{
		GetBook: func(ctx context.Context) (*domain.PersonalBook, error) {
			return s.personalBookRepo.GetByUserAndAuthorAndTitle(ctx, translationContext.UserID, translationContext.ParsedData.Book.Author, translationContext.ParsedData.Book.Title)
		},
		CreateBook: func(ctx context.Context) (*domain.PersonalBook, error) {
			author, err := s.authorRepo.GetOrCreate(ctx, translationContext.ParsedData.Book.Author)
			if err != nil {
				return nil, fmt.Errorf("failed to get or create author: %w", err)
			}

			var coverURL *string
			if translationContext.ParsedData.Cover != nil {
				coverPath, err := s.saveCoverToS3(translationContext.ParsedData.Cover, translationContext.ParsedData.Book.Title, translationContext.UserID)
				if err != nil {
					slog.Error(err.Error(), slog.String("operation", operation))
				} else {
					coverURL = &coverPath
				}
			}

			newBook := &domain.PersonalBook{
				Author:         author,
				Title:          translationContext.ParsedData.Book.Title,
				CoverURL:       coverURL,
				OriginalLang:   string(translationContext.From),
				TranslatedLang: string(translationContext.To),
				User: &domain.User{
					ID: translationContext.UserID,
				},
				ProcessStatus: domain.ProcessStatusInProgress,
			}
			newBook, err = s.personalBookRepo.Create(ctx, newBook)
			if err != nil {
				slog.Error("failed to create personal book", slog.Any("error", err), slog.String("operation", operation))
				return nil, ErrFailedSaveBookToDatabase
			}
			return newBook, nil
		},
		UpdateBookStatus: func(ctx context.Context, id domain.ID, status domain.ProcessStatus) (*domain.PersonalBook, error) {
			return s.personalBookRepo.UpdateProcessStatus(ctx, id, status)
		},
		GetChapter: func(ctx context.Context, bookID domain.ID, order int) (*domain.PersonalBookChapter, error) {
			return s.personalBookRepo.GetChapterByPersonalBookIDAndOrder(ctx, bookID, order)
		},
		CreateChapter: func(ctx context.Context, book *domain.PersonalBook, chapter *domain.ChapterAlignNode, contentURL string) (*domain.PersonalBookChapter, error) {
			newChapter := &domain.PersonalBookChapter{
				PersonalBook:    book,
				Order:           chapter.Order,
				ContentURL:      contentURL,
				Title:           chapter.Title,
				TranslatedTitle: chapter.TranslatedTitle,
			}
			return s.personalBookRepo.CreateChapter(ctx, newChapter)
		},
		SaveChapterToS3: func(chapter *domain.ChapterAlignNode) (string, error) {
			path, err := s.saveChapterToS3(chapter, translationContext.ParsedData.Book.Title, translationContext.UserID)
			if err != nil {
				return "", ErrFailedSaveToStorage
			}
			return path, nil
		},
		SaveImageToS3: func(data []byte) (string, error) {
			return s.saveImageToS3(data, translationContext.ParsedData.Book.Title, translationContext.UserID)
		},
		TranslateChapter: func(ctx context.Context, chapter epub_parser.FormattedChapter, imageSaver func([]byte) (string, error)) (*domain.ChapterAlignNode, error) {
			node, err := s.logic.TranslateChapter(ctx, chapter, translationContext.From, translationContext.To, s, s.config.MaxRequestCount, imageSaver)
			if err != nil {
				return nil, ErrFailedTranslateChapter
			}
			return node, nil
		},
		NotifyChapter: func(ctx context.Context, chapter *domain.PersonalBookChapter) {
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
		NotifyBook: func(ctx context.Context, book *domain.PersonalBook) {
			if s.notificationSignal != nil {
				s.notificationSignal.Emit(ctx, &domain.Notification{
					UserId: translationContext.UserID,
					Type:   domain.NotificationPersonalBookTranslated,
					Data: &domain.NotificationPersonalBookTranslatedData{
						Book: book,
					},
				})
			}
		},
		NotifyError: func(ctx context.Context, err error) {
			if s.notificationSignal != nil {
				var errorMessage string
				switch {
				case errors.Is(err, ErrFailedSaveToStorage):
					errorMessage = "Произошла ошибка во время сохранения главы, попробуйте перевести книгу позже."
				case errors.Is(err, ErrFailedTranslateChapter):
					errorMessage = "Произошла ошибка во время перевода главы, попробуйте перевести книгу позже."
				case errors.Is(err, ErrFailedSaveBookToDatabase):
					errorMessage = "Произошла ошибка во время сохранения книги, попробуйте перевести книгу позже, извините за неудобства."
				default:
					errorMessage = err.Error()
				}

				s.notificationSignal.Emit(ctx, &domain.Notification{
					UserId: translationContext.UserID,
					Type:   domain.NotificationError,
					Data: &domain.NotificationErrorData{
						Title:        translationContext.ParsedData.Book.Title,
						Author:       translationContext.ParsedData.Book.Author,
						ErrorMessage: errorMessage,
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

	if runAsync {
		go runChapters()
	} else {
		runChapters()
	}

	return &Output{Message: "start translate"}, nil
}

func (s *Service) saveImageToS3(data []byte, bookTitle string, userId domain.ID) (string, error) {
	const operation = "translate_personal_user_book.Service.saveImageToS3"

	// Генерируем уникальное имя файла
	filename := uuid.New().String() + ".jpg" // Предполагаем jpg, в идеале нужно определять тип
	path := fmt.Sprintf("user/%s/book/%s/images/%s", userId, strings.ReplaceAll(bookTitle, " ", "_"), filename)

	if err := s.s3.Upload(path, data); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
}

func (s *Service) saveChapterToS3(chapter *domain.ChapterAlignNode, bookTitle string, userId domain.ID) (string, error) {
	const operation = "translate_personal_user_book.Service.saveChapterToS3"

	path := fmt.Sprintf("user/%s/book/%s/%d.json", userId, strings.ReplaceAll(bookTitle, " ", "_"), chapter.Order)

	data, err := json.Marshal(chapter)
	if err != nil {
		return "", fmt.Errorf("%s: failed to marshal chapter: %w", operation, err)
	}

	if err := s.s3.Upload(path, data); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
}

func (s *Service) saveCoverToS3(coverData []byte, bookTitle string, userId domain.ID) (string, error) {
	const operation = "translate_personal_user_book.Service.saveCoverToS3"

	path := fmt.Sprintf("user/%s/cover/%s/%s", userId, strings.ReplaceAll(bookTitle, " ", "_"), uuid.New().String())

	if err := s.s3.Upload(path, coverData); err != nil {
		return "", fmt.Errorf("%s: failed upload to s3: %w", operation, err)
	}

	return path, nil
}
